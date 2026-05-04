package app

import (
	"crypto/sha256"
	"crypto/subtle"
	"database/sql"
	"embed"
	"fmt"
	"log"
	"net/http"
	"runtime"
	"strconv"
	"strings"

	"github.com/mugtree/grounds/app/db"

	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
)

//go:embed public/css/*.css
//go:embed public/js/*.js
//go:embed public/js/manifest.json
//go:embed public/img/*.png
var staticFS embed.FS

func ServerSetup(queries *db.Queries, sqldb *sql.DB, uploadsDir string, sessions *scs.SessionManager, user string, password string) chi.Router {

	r := chi.NewRouter()
	r.Use(sessions.LoadAndSave)
	r.Handle("/public/*", neuterDirectoryHandler(http.FileServer(http.FS(staticFS))))
	r.Group(func(site chi.Router) {
		site.Use(basicAuthHandler(user, password))
		setupRoutes(site, queries, sqldb, sessions, uploadsDir)
	})

	return r
}

func errorHandler(w http.ResponseWriter, r *http.Request, msg string, statusCode ...int) {
	status := 500
	if len(statusCode) > 0 {
		status = statusCode[0]
	}

	_, file, line, ok := runtime.Caller(1) // 1 = caller of this function
	if ok {
		msg = fmt.Sprintf("%s (at %s:%d)", msg, file, line)
	}
	LogError(msg)
	w.WriteHeader(status)
	ErrorPageTemplate().Render(r.Context(), w)
}

func neuterDirectoryHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/") {
			http.NotFound(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func basicAuthHandler(user string, user_password string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			username, password, ok := r.BasicAuth()

			if ok {
				usernameHash := sha256.Sum256([]byte(username))
				passwordHash := sha256.Sum256([]byte(password))
				expectedUsernameHash := sha256.Sum256([]byte(user))
				expectedPasswordHash := sha256.Sum256([]byte(user_password))

				usernameMatch := subtle.ConstantTimeCompare(usernameHash[:], expectedUsernameHash[:]) == 1
				passwordMatch := subtle.ConstantTimeCompare(passwordHash[:], expectedPasswordHash[:]) == 1

				if usernameMatch && passwordMatch {
					next.ServeHTTP(w, r)
					return
				}
			}

			w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
		})
	}
}

func pathValueAsIntOrErr(w http.ResponseWriter, r *http.Request, key string) (int64, bool) {

	formVal := r.PathValue(key)

	if formVal == "" {
		errorHandler(w, r, fmt.Sprintf("http: incorrect path value %s on page %v", key, r.URL.Path))
		return 0, false
	}

	val, err := strconv.ParseInt(formVal, 10, 64)
	if err != nil {
		errorHandler(w, r, fmt.Sprintf("http: incorrect path value %v, should be numeric - on page %v", formVal, r.URL.Path))
		return 0, false
	}

	return val, true

}

func formValueAsIntOrErr(w http.ResponseWriter, r *http.Request, key string) (int64, bool) {

	formVal := r.FormValue(key)

	if formVal == "" {
		errorHandler(w, r, fmt.Sprintf("http: incorrect form value %s on page %v", key, r.URL.Path))
		return 0, false
	}

	val, err := strconv.ParseInt(formVal, 10, 64)
	if err != nil {
		errorHandler(w, r, fmt.Sprintf("http: incorrect form value %v, should be numeric - on page %v", formVal, r.URL.Path))
		return 0, false
	}

	return val, true

}

func parseMultipart(r *http.Request) (*http.Request, error) {

	ct := r.Header.Get("Content-Type")
	if strings.HasPrefix(ct, "multipart/form-data") {
		err := r.ParseMultipartForm(10 << 20)
		if err != nil {
			return r, err
		}
	} else {
		err := r.ParseForm()
		if err != nil {
			return r, err
		}
	}

	return r, nil
}

func LogInfo(msg string)  { log.Println("INFO: " + msg) }
func LogError(msg string) { log.Println("ERROR: " + msg) }
