package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/0x24CaptainParrot/gophermart-service/internal/logger"
	"github.com/0x24CaptainParrot/gophermart-service/internal/models"
	"github.com/0x24CaptainParrot/gophermart-service/internal/pkg/repository"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OrderProcessingService struct {
	repo        repository.WorkerPoolRepository
	accrualAddr string
	pool        *pgxpool.Pool
	channel     string
	workers     int
	cancel      context.CancelFunc
	client      *http.Client
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
		repo:        repository.NewWorkerPoolRepo(pool),
		accrualAddr: accrualAddr,
		pool:        pool,
		channel:     channel,
		client:      client,
	}, nil
}

func (s *OrderProcessingService) StartProcessing(ctx context.Context, workers int) {
	ctx, s.cancel = context.WithCancel(ctx)
	s.workers = workers

	for i := 0; i < workers; i++ {
		go s.worker(ctx, i)
	}

	go s.processExistingOrders(ctx)
}

func (s *OrderProcessingService) worker(ctx context.Context, workerID int) {
	conn, err := s.pool.Acquire(ctx)
	if err != nil {
		logger.Log.Sugar().Errorf("worker %d: failed to acquire connection: %v", workerID, err)
		return
	}
	defer conn.Release()

	if _, err := conn.Exec(ctx, fmt.Sprintf("LISTEN %s", s.channel)); err != nil {
		logger.Log.Sugar().Errorf("worker %d: failed to listen channel: %v", workerID, err)
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
				continue
			}

			var orderNumber int64
			if _, err := fmt.Sscan(notification.Payload, &orderNumber); err != nil {
				logger.Log.Sugar().Errorf("Failed to parse order number: %v", err)
				continue
			}

			locked, err := s.tryLockOrder(ctx, orderNumber)
			if err != nil {
				logger.Log.Sugar().Errorf("Failed to lock order: %d: %v", orderNumber, err)
				continue
			}

			if !locked {
				continue
			}

			if err := s.processOrder(ctx, orderNumber); err != nil {
				logger.Log.Sugar().Errorf("Worker %d failed to process order %d: %v", workerID, orderNumber, err)
				continue
			}
			logger.Log.Sugar().Infof("Worker %d successfully processed order: %d", workerID, orderNumber)
		}
	}
}

func (s *OrderProcessingService) tryLockOrder(ctx context.Context, orderNumber int64) (bool, error) {
	var locked bool
	err := s.pool.QueryRow(ctx, `SELECT pg_try_advisory_xact_lock($1)`, orderNumber).Scan(&locked)
	return locked, err
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
				if err := s.processOrder(ctx, order.Number); err != nil {
					logger.Log.Sugar().Errorf("Failed to process order %d: %v", order.Number, err)
				}
			}
		}
	}
}

func (s *OrderProcessingService) processOrder(ctx context.Context, orderNumber int64) error {
	status, err := s.repo.LockAndGetOrderStatus(ctx, orderNumber)
	if err != nil {
		if errors.Is(err, repository.ErrOrderNotFound) {
			logger.Log.Sugar().Warnf("Order: %d not found", orderNumber)
			return nil
		}
		return fmt.Errorf("failed to check order status: %w", err)
	}

	if status == "PROCESSED" {
		return nil
	}

	accrualData, err := s.getAccrual(ctx, orderNumber)
	if err != nil {
		logger.Log.Sugar().Errorf("failed to fetch data from accrual: Code: ")
		return fmt.Errorf("failed to get data from accrual: %v", err)
	}

	if accrualData.Accrual == 0 {
		logger.Log.Sugar().Infof("Order with number: %d has 0 loyalty points", accrualData.Order)
		return nil
	}

	order := models.Order{
		Number: orderNumber,
		Status: accrualData.Status,
	}

	logger.Log.Sugar().Infof("updating with: number: %d, status: %s, accrual: %d", order.Number, accrualData.Status, accrualData.Accrual)
	if err := s.repo.UpdateOrderAndBalance(ctx, order, accrualData.Accrual); err != nil {
		logger.Log.Sugar().Errorf("failed to update order with number: %d. error: %v", order.Number, err)
		return fmt.Errorf("failed to update order: %v", err)
	}

	return nil
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

	var accrualData models.AccrualResponse
	if err := json.NewDecoder(resp.Body).Decode(&accrualData); err != nil {
		logger.Log.Sugar().Errorf("Failed to decode accrual json data: %v", err)
		return nil, err
	}

	accrualData.StatusCode = resp.StatusCode
	return &accrualData, nil
}

func (s *OrderProcessingService) EnqueueOrder(ctx context.Context, order models.Order) error {
	_, err := s.pool.Exec(ctx, `SELECT pg_notify($1, $2)`, s.channel, strconv.FormatInt(order.Number, 10))
	return err
}

func (s *OrderProcessingService) StopProcessing() {
	if s.cancel != nil {
		s.cancel()
	}
}
