package auth

import (
	"net/http"

	"github.com/hsm-gustavo/authentication/internal/database"
	"github.com/hsm-gustavo/authentication/internal/domains/email"
	"github.com/hsm-gustavo/authentication/internal/domains/jwt"
)

func RegisterModule(queries *database.Queries, jwtService *jwt.Service, emailService *email.Service) http.Handler {
	service := NewService(queries, jwtService, emailService)
	controller := NewController(service)

	mux := http.NewServeMux()

	mux.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
			return
		}
		controller.RegisterHandler(w, r)
	})
	
	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
			return
		}
		controller.LoginHandler(w, r)
	})

	mux.HandleFunc("/confirm-email", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
			return
		}
		controller.ConfirmEmailHandler(w, r)
	})

	return mux
}