package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

type application struct {
	serviceName string
	dbName      string
	db          *sql.DB
}

type response struct {
	Code      int         `json:"code"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp int64       `json:"timestamp"`
}

func runServer(addr string, app *application) error {
	return http.ListenAndServe(addr, newRouter(app))
}

func newRouter(app *application) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", app.handleIndex)
	mux.HandleFunc("/healthz", app.handleHealthz)
	mux.HandleFunc("/api/v1/hello", app.handleHello)
	mux.HandleFunc("/api/v1/mysql", app.handleMySQL)
	mux.HandleFunc("/api/v1/users", app.handleUsers)
	return mux
}

func (app *application) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		writeJSON(w, http.StatusNotFound, response{
			Code:      http.StatusNotFound,
			Message:   "not found",
			Timestamp: time.Now().Unix(),
		})
		return
	}
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}

	writeJSON(w, http.StatusOK, response{
		Code:    http.StatusOK,
		Message: "ok",
		Data: map[string]interface{}{
			"service":  app.serviceName,
			"mysql_db": app.dbName,
			"endpoints": []string{
				"GET /",
				"GET /healthz",
				"GET /api/v1/hello?name=Tom",
				"GET /api/v1/mysql",
				"GET /api/v1/users",
			},
		},
		Timestamp: time.Now().Unix(),
	})
}

func (app *application) handleHealthz(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}
	if app.db == nil {
		writeJSON(w, http.StatusServiceUnavailable, response{
			Code:      http.StatusServiceUnavailable,
			Message:   "mysql is not connected",
			Timestamp: time.Now().Unix(),
		})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	if err := app.db.PingContext(ctx); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, response{
			Code:      http.StatusServiceUnavailable,
			Message:   "mysql ping failed",
			Data:      map[string]string{"error": err.Error()},
			Timestamp: time.Now().Unix(),
		})
		return
	}

	writeJSON(w, http.StatusOK, response{
		Code:    http.StatusOK,
		Message: "ok",
		Data: map[string]string{
			"service":  app.serviceName,
			"mysql_db": app.dbName,
			"status":   "healthy",
		},
		Timestamp: time.Now().Unix(),
	})
}

func (app *application) handleHello(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}

	name := r.URL.Query().Get("name")
	if name == "" {
		name = "World"
	}

	writeJSON(w, http.StatusOK, response{
		Code:    http.StatusOK,
		Message: "ok",
		Data: map[string]string{
			"name":     name,
			"message":  "Hello, " + name,
			"service":  app.serviceName,
			"mysql_db": app.dbName,
		},
		Timestamp: time.Now().Unix(),
	})
}

func (app *application) handleMySQL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}
	if app.db == nil {
		writeJSON(w, http.StatusServiceUnavailable, response{
			Code:      http.StatusServiceUnavailable,
			Message:   "mysql is not connected",
			Timestamp: time.Now().Unix(),
		})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	dbName, err := currentDatabase(ctx, app.db)
	if err != nil {
		writeJSON(w, http.StatusServiceUnavailable, response{
			Code:      http.StatusServiceUnavailable,
			Message:   "mysql query failed",
			Data:      map[string]string{"error": err.Error()},
			Timestamp: time.Now().Unix(),
		})
		return
	}

	writeJSON(w, http.StatusOK, response{
		Code:    http.StatusOK,
		Message: "ok",
		Data: map[string]string{
			"service":  app.serviceName,
			"mysql_db": dbName,
		},
		Timestamp: time.Now().Unix(),
	})
}

func (app *application) handleUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}
	if app.db == nil {
		writeJSON(w, http.StatusServiceUnavailable, response{
			Code:      http.StatusServiceUnavailable,
			Message:   "mysql is not connected",
			Timestamp: time.Now().Unix(),
		})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	usernames, err := listUsernames(ctx, app.db)
	if err != nil {
		writeJSON(w, http.StatusServiceUnavailable, response{
			Code:      http.StatusServiceUnavailable,
			Message:   "mysql query failed",
			Data:      map[string]string{"error": err.Error()},
			Timestamp: time.Now().Unix(),
		})
		return
	}

	writeJSON(w, http.StatusOK, response{
		Code:    http.StatusOK,
		Message: "ok",
		Data: map[string]interface{}{
			"service":   app.serviceName,
			"mysql_db":  app.dbName,
			"usernames": usernames,
		},
		Timestamp: time.Now().Unix(),
	})
}

func writeMethodNotAllowed(w http.ResponseWriter) {
	writeJSON(w, http.StatusMethodNotAllowed, response{
		Code:      http.StatusMethodNotAllowed,
		Message:   "method not allowed",
		Timestamp: time.Now().Unix(),
	})
}

func writeJSON(w http.ResponseWriter, status int, payload response) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("write json response failed: %v", err)
	}
}
