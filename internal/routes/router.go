package routes

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/hsm-gustavo/authentication/internal/app"
	"github.com/hsm-gustavo/authentication/internal/domains/auth"
	"github.com/hsm-gustavo/authentication/internal/domains/jwt"
	"github.com/hsm-gustavo/authentication/internal/middlewares"
)

func Setup(app *app.Application) http.Handler {
	mux := http.NewServeMux()

	// cada handler deve ser registrado aqui
	
	healthHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{"message": "Hello, World!"})
		}
	})
	authHandler := auth.RegisterModule(app.DB, jwt.NewJWTService(app.Config.JWTSecret, 1 * time.Hour))

	mux.Handle("/health", middlewares.Wrap(healthHandler, app.Logger))
	mux.Handle("/auth/", http.StripPrefix("/auth", middlewares.Wrap(authHandler, app.Logger)))

	return mux
}