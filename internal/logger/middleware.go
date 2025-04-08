package logger

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

type TrackRequestWriter struct {
	w       http.ResponseWriter
	reqInfo *ResponseInfo
}

type ResponseInfo struct {
	code int
	size int
}

func (tw *TrackRequestWriter) Write(b []byte) (int, error) {
	n, err := tw.w.Write(b)
	tw.reqInfo.size += n
	return n, err
}

func (tw *TrackRequestWriter) WriteHeader(statusCode int) {
	tw.reqInfo.code = statusCode
	tw.w.WriteHeader(statusCode)
}

func (tw *TrackRequestWriter) Header() http.Header {
	return tw.w.Header()
}

func LoggingReqResMiddleware(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			reqInfo := &ResponseInfo{}
			trw := &TrackRequestWriter{
				w:       w,
				reqInfo: reqInfo,
			}

			start := time.Now()
			h.ServeHTTP(trw, r)
			elapsed := time.Since(start)

			logger.Info("HTTP Request was made",
				zap.String("method:", r.Method),
				zap.String("uri:", r.RequestURI),
				zap.Int("status code:", trw.reqInfo.code),
				zap.Int("size:", trw.reqInfo.size),
				zap.Duration("elapsed:", elapsed))
		})
	}
}
