package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type serverOptions struct {
	configPath string
	addr       string
}

func main() {
	opts, err := parseServerOptions(os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}

	serviceName, err := serviceNameFromExecutable()
	if err != nil {
		log.Fatal(err)
	}

	cfg, err := loadConfig(opts.configPath)
	if err != nil {
		log.Fatalf("load config failed: %v", err)
	}
	if err := cfg.validate(serviceName); err != nil {
		log.Fatalf("invalid config: %v", err)
	}

	db, dbName, err := connectMySQL(cfg.MySQL, 10*time.Second)
	if err != nil {
		log.Fatalf("mysql connect failed: %v", err)
	}
	defer db.Close()

	if dbName != serviceName {
		log.Fatalf("mysql database mismatch: connected=%q expected=%q", dbName, serviceName)
	}

	app := &application{
		serviceName: serviceName,
		dbName:      dbName,
		db:          db,
	}

	log.Printf("service=%s mysql_db=%s listening on %s", serviceName, dbName, opts.addr)
	if err := runServer(opts.addr, app); err != nil {
		log.Fatal(err)
	}
}

func parseServerOptions(args []string) (serverOptions, error) {
	opts := serverOptions{
		configPath: "settings.dev.yml",
		addr:       defaultListenAddr(),
	}

	if len(args) > 0 {
		switch args[0] {
		case "server":
			args = args[1:]
		default:
			if !strings.HasPrefix(args[0], "-") {
				return opts, fmt.Errorf("unknown command %q", args[0])
			}
		}
	}

	fs := flag.NewFlagSet("server", flag.ContinueOnError)
	fs.StringVar(&opts.configPath, "c", opts.configPath, "config file path")
	fs.StringVar(&opts.addr, "addr", opts.addr, "listen address")
	fs.String("a", "false", "compatibility flag")

	if err := fs.Parse(args); err != nil {
		return opts, err
	}
	if fs.NArg() != 0 {
		return opts, fmt.Errorf("unexpected args: %v", fs.Args())
	}

	return opts, nil
}

func defaultListenAddr() string {
	port := strings.TrimSpace(os.Getenv("PORT"))
	if port == "" {
		port = "8989"
	}
	if strings.HasPrefix(port, ":") {
		return port
	}
	return ":" + port
}

func serviceNameFromExecutable() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}

	name := strings.TrimSuffix(filepath.Base(exe), ".exe")
	if name == "" || name == "." {
		return "", fmt.Errorf("cannot detect service name from executable path %q", exe)
	}

	if envName := strings.TrimSpace(os.Getenv("SERVICE_NAME")); envName != "" && envName != name {
		return "", fmt.Errorf("SERVICE_NAME=%q does not match executable name %q", envName, name)
	}

	return name, nil
}
