package auth

import (
	"encoding/json"
	"net/http"
)

type Controller struct {
	service *Service
}

func NewController(service *Service) *Controller {
	return &Controller{service: service}
}

func (c *Controller) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var dto RegisterDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	if dto.Email == "" || dto.Password == "" {
		http.Error(w, "Email e senha são obrigatórios", http.StatusBadRequest)
		return
	}

	err := c.service.Register(r.Context(), dto)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (c *Controller) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var dto LoginDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	response, err := c.service.Login(r.Context(), dto)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	responseCleaned := struct {
		AccessToken string `json:"access_token"`
	}{
		AccessToken: response.AccessToken,
	}

	http.SetCookie(w, &http.Cookie{
		Name: "session_id",
		Value: response.SessionID,
		Path: "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	http.SetCookie(w, &http.Cookie{
		Name: "session_secret",
		Value: response.SessionSecret,
		Path: "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responseCleaned)
}

func (c *Controller) ConfirmEmailHandler(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	email := r.URL.Query().Get("email")

	if code == "" || email == "" {
		http.Error(w, "Código e email são obrigatórios", http.StatusBadRequest)
		return
	}

	err := c.service.ConfirmEmail(r.Context(), code, email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (c *Controller) RefreshHandler(w http.ResponseWriter, r *http.Request) {
	idCookie, errID := r.Cookie("session_id")
	secretCookie, errSecret := r.Cookie("session_secret")

	if errID != nil || errSecret != nil {
		http.Error(w, "Sessão ausente", http.StatusUnauthorized)
		return
	}

	newAccessToken, err := c.service.RefreshAccess(r.Context(), idCookie.Value, secretCookie.Value)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	response := RefreshResponse{
		AccessToken: newAccessToken,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}