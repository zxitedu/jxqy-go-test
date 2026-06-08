package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHelloReturnsJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/hello?name=Tom", nil)
	rec := httptest.NewRecorder()

	newRouter(testApplication()).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d", rec.Code, http.StatusOK)
	}

	if got := rec.Header().Get("Content-Type"); got != "application/json; charset=utf-8" {
		t.Fatalf("content type = %q, want json", got)
	}

	var body struct {
		Code int `json:"code"`
		Data struct {
			Name    string `json:"name"`
			Message string `json:"message"`
			Service string `json:"service"`
			MySQLDB string `json:"mysql_db"`
		} `json:"data"`
		Message string `json:"message"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if body.Code != http.StatusOK || body.Message != "ok" {
		t.Fatalf("response = %+v, want success", body)
	}
	if body.Data.Name != "Tom" || body.Data.Message != "Hello, Tom" {
		t.Fatalf("data = %+v, want hello Tom", body.Data)
	}
	if body.Data.Service != "api" || body.Data.MySQLDB != "api" {
		t.Fatalf("data = %+v, want api service/mysql db", body.Data)
	}
}

func TestNotFoundReturnsJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/missing", nil)
	rec := httptest.NewRecorder()

	newRouter(testApplication()).ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status code = %d, want %d", rec.Code, http.StatusNotFound)
	}

	var body struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Code != http.StatusNotFound || body.Message != "not found" {
		t.Fatalf("response = %+v, want not found", body)
	}
}

func TestUsersRequiresMySQLConnection(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	rec := httptest.NewRecorder()

	newRouter(testApplication()).ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status code = %d, want %d", rec.Code, http.StatusServiceUnavailable)
	}

	var body struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Code != http.StatusServiceUnavailable || body.Message != "mysql is not connected" {
		t.Fatalf("response = %+v, want mysql not connected", body)
	}
}

func TestConfigDBMustMatchServiceName(t *testing.T) {
	cfg := config{
		MySQL: mysqlConfig{
			Host:     "127.0.0.1",
			Port:     3306,
			DB:       "admin",
			User:     "root",
			Password: "secret",
		},
	}

	if err := cfg.validate("api"); err == nil {
		t.Fatal("validate returned nil, want mysql db mismatch error")
	}

	cfg.MySQL.DB = "api"
	if err := cfg.validate("api"); err != nil {
		t.Fatalf("validate returned error: %v", err)
	}
}

func TestParseServerOptionsAcceptsDockerCMD(t *testing.T) {
	opts, err := parseServerOptions([]string{"server", "-c", "/app/config/settings.dev.yml", "-a", "true"})
	if err != nil {
		t.Fatalf("parse server options: %v", err)
	}
	if opts.configPath != "/app/config/settings.dev.yml" {
		t.Fatalf("config path = %q, want docker config path", opts.configPath)
	}
}

func testApplication() *application {
	return &application{
		serviceName: "api",
		dbName:      "api",
	}
}
