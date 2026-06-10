package routes

import (
	"encoding/json"
	"io"
	"io/fs"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/hsm-gustavo/authentication/internal/app"
	"github.com/hsm-gustavo/authentication/internal/domains/auth"
	"github.com/hsm-gustavo/authentication/internal/domains/email"
	"github.com/hsm-gustavo/authentication/internal/domains/jwt"
	"github.com/hsm-gustavo/authentication/internal/middlewares"
	"github.com/hsm-gustavo/authentication/ui"
)

func spaHandler(staticFS fs.FS) http.Handler {
	fileServer := http.FileServer(http.FS(staticFS))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// limpa url para evitar ataques de directory traversal (../../)
		path := filepath.Clean(r.URL.Path)

		file, err := staticFS.Open(strings.TrimPrefix(path, "/"))
		if err == nil {
			file.Close()
			fileServer.ServeHTTP(w, r)
		}

		indexFile, err := staticFS.Open("index.html")
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		defer indexFile.Close()

		stat, err := indexFile.Stat()
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		http.ServeContent(w, r, "index.html", stat.ModTime(), indexFile.(io.ReadSeeker))
	})
}

func Setup(app *app.Application) http.Handler {
	mainMux := http.NewServeMux()

	// cada handler deve ser registrado aqui

	v1 := http.NewServeMux()
	
	healthHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{"message": "Hello, World!"})
		}
	})
	authHandler := auth.RegisterModule(app.DB, jwt.NewJWTService(app.Config.JWTSecret, 1 * time.Hour), email.NewService(&app.Config.SMTPConfig))

	v1.Handle("/health", healthHandler)
	v1.Handle("/auth/", http.StripPrefix("/auth", authHandler))

	mainMux.Handle("/api/v1/", http.StripPrefix("/api/v1", v1))

	reactFS := ui.GetFS()
	mainMux.Handle("/", spaHandler(reactFS))

	return middlewares.Wrap(mainMux, app.Logger)
}