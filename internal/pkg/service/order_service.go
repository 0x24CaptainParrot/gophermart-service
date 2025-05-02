package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/0x24CaptainParrot/gophermart-service/internal/logger"
	"github.com/0x24CaptainParrot/gophermart-service/internal/models"
	"github.com/0x24CaptainParrot/gophermart-service/internal/pkg/repository"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OrderProcessing interface {
	StartProcessing(ctx context.Context, workers int)
	EnqueueOrder(ctx context.Context, order models.Order) error
	StopProcessing()
}

type OrderProcessingService struct {
	repo         repository.WorkerPoolRepository
	accrualAddr  string
	pool         *pgxpool.Pool
	channel      string
	workers      int
	cancel       context.CancelFunc
	client       *http.Client
	ordersQueue  chan int64
	updatesQueue chan models.Order
	workersWg    sync.WaitGroup
	orderLocks   *sync.Map
}

func NewOrderProcessingService(pool *pgxpool.Pool, accrualAddr string, channel string) (*OrderProcessingService, error) {
	client := &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:        50,
			MaxIdleConnsPerHost: 20,
			IdleConnTimeout:     60 * time.Second,
		},
		Timeout: 15 * time.Second,
	}

	return &OrderProcessingService{
		repo:         repository.NewWorkerPoolRepo(pool),
		accrualAddr:  accrualAddr,
		pool:         pool,
		channel:      channel,
		client:       client,
		ordersQueue:  make(chan int64, 1000),
		updatesQueue: make(chan models.Order, 1000),
		workersWg:    sync.WaitGroup{},
		orderLocks:   &sync.Map{},
	}, nil
}

func (s *OrderProcessingService) StartProcessing(ctx context.Context, workers int) {
	ctx, s.cancel = context.WithCancel(ctx)
	s.workers = workers

	go s.notificationListener(ctx)

	for i := 0; i < workers; i++ {
		s.workersWg.Add(1)
		go s.worker(ctx, i)
	}

	s.workersWg.Add(1)
	go s.updateWorker(ctx)

	go s.processExistingOrders(ctx)
}

func (s *OrderProcessingService) notificationListener(ctx context.Context) {
	conn, err := s.pool.Acquire(ctx)
	if err != nil {
		logger.Log.Sugar().Errorf("failed to acquire listener conn: %v", err)
		return
	}
	defer conn.Release()

	if _, err := conn.Exec(ctx, fmt.Sprintf("LISTEN %s", s.channel)); err != nil {
		logger.Log.Sugar().Fatalf("failed to listen: %v", err)
		return
	}

	for {
		select {
		case <-ctx.Done():
			return
		default:
			notification, err := conn.Conn().WaitForNotification(ctx)
			if err != nil {
				if errors.Is(err, context.Canceled) {
					return
				}
				logger.Log.Sugar().Errorf("Notification error: %v", err)
				time.Sleep(100 * time.Millisecond)
				continue
			}

			var orderNumber int64
			if _, err := fmt.Sscan(notification.Payload, &orderNumber); err != nil {
				logger.Log.Sugar().Errorf("Failed to parse order number: %v", err)
				continue
			}

			select {
			case s.ordersQueue <- orderNumber:
			default:
				logger.Log.Sugar().Warnf("queue is full, drop order %d", orderNumber)
			}
		}
	}
}

func (s *OrderProcessingService) worker(ctx context.Context, workerID int) {
	defer s.workersWg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case orderNumber := <-s.ordersQueue:
			if _, loaded := s.orderLocks.LoadOrStore(orderNumber, struct{}{}); loaded {
				continue
			}

			go func(order int64) {
				defer s.orderLocks.Delete(order)

				orderForUpdate, err := s.processOrder(ctx, order)
				if err != nil {
					logger.Log.Sugar().Errorf("Worker %d failed to process order %d: %v", workerID, orderNumber, err)
					return
				}
				if orderForUpdate.IsEmpty() {
					logger.Log.Sugar().Infof("Worker %d: skipping empty order for %d", workerID, orderNumber)
					return
				}

				s.updatesQueue <- orderForUpdate
				logger.Log.Sugar().Infof("Worker %d successfully processed order: %d", workerID, orderNumber)
			}(orderNumber)
		}
	}
}

func (s *OrderProcessingService) updateWorker(ctx context.Context) {
	defer s.workersWg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case order := <-s.updatesQueue:
			logger.Log.Sugar().Infof("updating with: number: %d, status: %s, accrual: %.2f", order.Number, order.Status, order.Accrual)

			var lastErr error
			for i := 0; i < 3; i++ {
				err := s.repo.UpdateOrderAndBalance(ctx, order, order.Accrual)
				if err == nil {
					logger.Log.Sugar().Infof("Order with number %d was updated successfully", order.Number)
					break
				}
				var pgError *pgconn.PgError
				if errors.As(err, &pgError) && pgerrcode.IsTransactionRollback(pgError.Code) {
					lastErr = err
					time.Sleep(time.Duration(i+1) * 100 * time.Millisecond)
					continue
				}
				lastErr = err
				break
			}

			if lastErr != nil {
				logger.Log.Sugar().Errorf("failed to update order with number: %d. error: %v", order.Number, lastErr)
			}
		}
	}
}

func (s *OrderProcessingService) processExistingOrders(ctx context.Context) {
	const batchSize = 100
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			orders, err := s.repo.GetPendingOrders(ctx, batchSize)
			if err != nil {
				logger.Log.Sugar().Errorf("Failed to get pending orders: %v", err)
				continue
			}

			for _, order := range orders {
				select {
				case s.ordersQueue <- order.Number:
				default:
					logger.Log.Sugar().Warnf("queue is full, skip %d", order.Number)
				}
			}
		}
	}
}

func (s *OrderProcessingService) processOrder(ctx context.Context, orderNumber int64) (models.Order, error) {
	status, err := s.repo.LockAndGetOrderStatus(ctx, orderNumber)
	if err != nil {
		if errors.Is(err, repository.ErrOrderNotFound) {
			status = "NEW"
		} else {
			return models.Order{}, fmt.Errorf("failed to check order status: %w", err)
		}
	}

	accrualData, err := s.getAccrual(ctx, orderNumber)
	if err != nil {
		return models.Order{}, fmt.Errorf("failed to get data from accrual: %v", err)
	}

	if accrualData.Accrual == 0 {
		logger.Log.Sugar().Infof("Order with number: %d has 0 loyalty points", accrualData.Order)
	}

	if accrualData.Status == "INVALID" {
		logger.Log.Sugar().Warnf("accrual service returned INVALID status for order: %d", orderNumber)
		return models.Order{}, nil
	}

	order := models.Order{
		Number:  orderNumber,
		Status:  accrualData.Status,
		Accrual: accrualData.Accrual,
	}

	if status == "NEW" {
		err := s.insertMissingOrder(ctx, orderNumber)
		if err != nil {
			return models.Order{}, fmt.Errorf("failed to insert missing order: %w", err)
		}
	}

	return order, nil
}

func (s *OrderProcessingService) getAccrual(ctx context.Context, orderNumber int64) (*models.AccrualResponse, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/api/orders/%d", s.accrualAddr, orderNumber), nil)
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent {
		logger.Log.Sugar().Infof("accrual service returned 204 No Content for order: %d", orderNumber)
		return nil, fmt.Errorf("accrual service returned 204 No Content")
	}

	if resp.StatusCode != http.StatusOK {
		logger.Log.Sugar().Errorf("accrual service returned status %d for order: %d", resp.StatusCode, orderNumber)
		return nil, fmt.Errorf("accrual service returned status: %d", resp.StatusCode)
	}

	if resp.ContentLength == 0 {
		logger.Log.Sugar().Warnf("accrual service returned empty body for order: %d", orderNumber)
		return nil, fmt.Errorf("accrual response body is empty")
	}

	var accrualData models.AccrualResponse
	if err := json.NewDecoder(resp.Body).Decode(&accrualData); err != nil {
		logger.Log.Sugar().Errorf("Failed to decode accrual json data: %v", err)
		return nil, err
	}

	accrualData.StatusCode = resp.StatusCode
	return &accrualData, nil
}

func (s *OrderProcessingService) insertMissingOrder(ctx context.Context, orderNumber int64) error {
	return s.repo.InsertMissingOrder(ctx, orderNumber)
}

func (s *OrderProcessingService) EnqueueOrder(ctx context.Context, order models.Order) error {
	_, err := s.pool.Exec(ctx, `SELECT pg_notify($1, $2)`, s.channel, strconv.FormatInt(order.Number, 10))
	return err
}

func (s *OrderProcessingService) StopProcessing() {
	if s.cancel != nil {
		s.cancel()
	}

	s.workersWg.Wait()
}
