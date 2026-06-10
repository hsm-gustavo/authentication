package middlewares

import (
	"log/slog"
	"net/http"
)

// um arquivo wrapper para eu não ter que colocar um por

func Wrap(next http.Handler, log *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// o request ID tem que ser o primeiro middleware para garantir que ele esteja no contexto para os outros middlewares e handlers
		// apesar de contraintuitivo, o request ID é o ultimo a ser chamado, pois ele é o primeiro a ser aplicado na cadeia de middlewares

		clientInfoMiddleware := ExtractClientInfo(next, log)
		reqLoggerMiddleware := RequestLogger(clientInfoMiddleware, log)
		
		reqIDMiddleware := RequestID(reqLoggerMiddleware)
		reqIDMiddleware.ServeHTTP(w, r)
	})
}