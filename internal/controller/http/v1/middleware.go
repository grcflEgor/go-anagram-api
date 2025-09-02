package v1

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"github.com/grcflEgor/go-anagram-api/pkg/logger"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

func LoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := uuid.New().String()
		reqLogger := logger.AppLogger.With(
			zap.String("request_id", requestID),
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
		)

		ctx := logger.WithLogger(r.Context(), reqLogger)

		reqLogger.Info("req started")
		start := time.Now()

		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		next.ServeHTTP(ww, r.WithContext(ctx))

		reqLogger.Info("req finished",
			zap.Int("status", ww.Status()),
			zap.Int("bytes", ww.BytesWritten()),
			zap.Duration("duration", time.Since(start)),
		)
	})
}

func OTelMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tracer := otel.Tracer("anagram-api/http")

		ctx, span := tracer.Start(
			r.Context(),
			r.Method+" "+r.URL.Path,
			trace.WithTimestamp(time.Now()),
			trace.WithSpanKind(trace.SpanKindServer),
		)
		defer span.End()

		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		next.ServeHTTP(ww, r.WithContext(ctx))

		span.SetAttributes(
			attribute.String("http.method", r.Method),
			attribute.String("http.url", r.URL.Path),
			attribute.Int("http.status_code", ww.Status()),
		)

		if ww.Status() >= 400 {
			span.SetStatus(codes.Error, http.StatusText(ww.Status()))
		}
	})
}
