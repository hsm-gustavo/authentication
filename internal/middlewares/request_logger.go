package middlewares

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

// ref: Chi
// https://github.com/go-chi/chi/blob/3b171578ca44dfd75ca3c5cbddc7b44c600a7b49/middleware/wrap_writer.go
// https://github.com/go-chi/chi/blob/3b171578ca44dfd75ca3c5cbddc7b44c600a7b49/middleware/logger.go

// http.ResponseWriter não deixa a gente acessar o status code e o tamanho da resposta, então vamos criar um wrapper para isso
type responseWriterWrapper struct {
	http.ResponseWriter
	statusCode int
	bytesWritten int
	wroteHeader bool
}

func newResponseWriterWrapper(w http.ResponseWriter) *responseWriterWrapper {
	return &responseWriterWrapper{ResponseWriter: w, statusCode: http.StatusOK}
}

// sobrescreve o método WriteHeader para capturar o status code e marcar que o header já foi escrito
func (rw *responseWriterWrapper) WriteHeader(statusCode int) {
	if rw.wroteHeader {
		return
	}
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
	rw.wroteHeader = true
}

// sobrescreve o método Write para capturar o tamanho da resposta e garantir que o header seja escrito antes de escrever o corpo
func (rw *responseWriterWrapper) Write(b []byte) (int, error) {
	if !rw.wroteHeader {
		rw.WriteHeader(http.StatusOK)
	}
	n, err := rw.ResponseWriter.Write(b)
	rw.bytesWritten += n
	return n, err
}

func RequestLogger(next http.Handler, log *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		ww := newResponseWriterWrapper(w)

		traceID := GetTraceID(r.Context())

		next.ServeHTTP(ww, r)

		elapsed := time.Since(start)

		attributes := []any{
			slog.String("method", r.Method),
			slog.String("uri", r.URL.RequestURI()),
			slog.Int("status", ww.statusCode),
			slog.Duration("elapsed", elapsed),
			slog.Int("bytes", ww.bytesWritten),
			slog.String("remote_ip", r.RemoteAddr),
		}

		if traceID != "" {
			attributes = append(attributes, slog.String("trace_id", traceID))
		}

		fmt.Println(traceID)

		ctx := r.Context()
		msg := "Requisição recebida"

		switch {
			case ww.statusCode >= 500:
				log.ErrorContext(ctx, msg, attributes...)
			case ww.statusCode >= 400:
				log.WarnContext(ctx, msg, attributes...)
			default:
				log.InfoContext(ctx, msg, attributes...)
		}
	})
}