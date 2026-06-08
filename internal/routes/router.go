package routes

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/hsm-gustavo/authentication/internal/middlewares"
)

func Setup(log *slog.Logger) http.Handler {
	mux := http.NewServeMux()

	// cada handler deve ser registrado aqui
	
	healthHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{"message": "Hello, World!"})
		}
	})

	mux.Handle("/health", middlewares.Wrap(healthHandler, log))

	return mux
}