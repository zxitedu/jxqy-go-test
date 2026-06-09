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
	if body.Data.Service != "api" || body.Data.MySQLDB != "apidb" {
		t.Fatalf("data = %+v, want api service/apidb mysql db", body.Data)
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
		Settings: settingsConfig{
			Application: applicationConfig{
				Name: "api",
				Port: 8989,
			},
			Database: databaseConfig{
				Driver: "mysql",
				Source: "root:secret@tcp(mysql:3306)/admindb?charset=utf8mb4&parseTime=True&loc=Local&timeout=1000ms",
			},
			Gen: genConfig{
				DBName: "admindb",
			},
		},
	}

	if err := cfg.validate("api"); err == nil {
		t.Fatal("validate returned nil, want mysql db mismatch error")
	}

	cfg.Settings.Database.Source = "root:secret@tcp(mysql:3306)/apidb?charset=utf8mb4&parseTime=True&loc=Local&timeout=1000ms"
	cfg.Settings.Gen.DBName = "apidb"
	if err := cfg.validate("api"); err != nil {
		t.Fatalf("validate returned error: %v", err)
	}
}

func TestMySQLDatabaseName(t *testing.T) {
	dbName, err := mysqlDatabaseName("root:secret@tcp(mysql:3306)/house_admindb?charset=utf8mb4&parseTime=True&loc=Local&timeout=1000ms")
	if err != nil {
		t.Fatalf("mysqlDatabaseName returned error: %v", err)
	}
	if dbName != "house_admindb" {
		t.Fatalf("dbName = %q, want house_admindb", dbName)
	}
}

func TestServiceDatabaseName(t *testing.T) {
	tests := map[string]string{
		"api":         "apidb",
		"admin":       "admindb",
		"house_admin": "house_admindb",
	}

	for serviceName, want := range tests {
		if got := serviceDatabaseName(serviceName); got != want {
			t.Fatalf("serviceDatabaseName(%q) = %q, want %q", serviceName, got, want)
		}
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
		dbName:      "apidb",
	}
}
