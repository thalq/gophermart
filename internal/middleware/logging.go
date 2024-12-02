package middleware

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

var Sugar *zap.SugaredLogger

type (
	responseData struct {
		status int
		size   int
	}

	loggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}
)

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(status int) {
	r.ResponseWriter.WriteHeader(status)
	r.responseData.status = status
}

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		responseData := &responseData{
			status: 0,
			size:   0,
		}
		lrw := &loggingResponseWriter{
			ResponseWriter: w,
			responseData:   responseData,
		}
		next.ServeHTTP(lrw, r)

		duration := time.Since(start)

		Sugar.Infow("request",
			"uri", r.RequestURI,
			"method", r.Method,
			"status", responseData.status,
			"size", responseData.size,
			"duration", duration,
		)
	})
}

func InitLogger() {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()
	Sugar = logger.Sugar()
}
