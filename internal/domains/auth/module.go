package auth

import (
	"net/http"

	"github.com/hsm-gustavo/authentication/internal/database"
	"github.com/hsm-gustavo/authentication/internal/domains/jwt"
)

func RegisterModule(queries *database.Queries, jwtService *jwt.Service) http.Handler {
	service := NewService(queries, jwtService)
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

	return mux
}