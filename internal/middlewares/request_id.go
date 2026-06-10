package middlewares

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"sync/atomic"
)

// Vamos seguir a ideia do middleware do Chi: https://github.com/go-chi/chi/blob/3b171578ca44dfd75ca3c5cbddc7b44c600a7b49/middleware/request_id.go
// vamos injetar um ID no contexto de cada request.
// seguindo os padroes da W3C, vamos ter um header traceparent
// <version>-<trace id>--<span id>-<flags>
// ref: https://www.w3.org/TR/trace-context-1/#traceparent-header

// trace_id é o ID único para cada request, seguindo o padrão W3C Trace Context
const traceIDKey contextKey = "trace_id"

type TraceParent struct {
	Version string
	TraceID string
	SpanID string
	Flags string
}

func (tp TraceParent) String() string {
	return fmt.Sprintf("%s-%s-%s-%s", tp.Version, tp.TraceID, tp.SpanID, tp.Flags)
}

var (
	processPrefix string
	reqSequence atomic.Uint64
)



func init() {
	buf := make([]byte, 8)
	_, err := rand.Read(buf)
	if err != nil {
		// fallback estático se não for possível gerar um ID aleatório
		processPrefix = "static"
	} else {
		processPrefix = hex.EncodeToString(buf)
	}
}

func GetTraceID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if id, ok := ctx.Value(traceIDKey).(string); ok {
		return id
	}
	return ""
}

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("traceparent")
		var tp TraceParent

		if header != "" && len(header) == 55 && strings.Count(header, "-") == 3 {
			parts := strings.Split(header, "-")
			// esse código usa a versão 00, então vamos checar se a versão é 00 antes de tentar parsear o header
			if parts[0] == "00" {
				tp = TraceParent{
					Version: parts[0],
					TraceID: parts[1],
					SpanID: parts[2],
					Flags: parts[3],
				}
			}
		}

		if tp.TraceID == "" {
			currentID := reqSequence.Add(1)

			// o trace ID é composto pelo prefixo do processo (8 bytes) + um contador sequencial (8 bytes), totalizando 16 bytes (32 caracteres hexadecimais)
			seqHex := fmt.Sprintf("%016x", currentID)
			traceID := processPrefix + seqHex

			// o span ID é o complemento do trace ID, para garantir que seja único mesmo em casos de reinício do processo
			spanHex := fmt.Sprintf("%016x", ^currentID)

			tp = TraceParent{
				Version: "00",
				TraceID: traceID,
				SpanID: spanHex,
				Flags: "01",
			}
		}

		ctx := context.WithValue(r.Context(), traceIDKey, tp.TraceID)

		w.Header().Set("traceparent", tp.String())

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}