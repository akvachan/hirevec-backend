// Copyright (c) 2026 Arsenii Kvachan
// SPDX-License-Identifier: MIT

package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	hirevec "github.com/akvachan/hirevec-backend/internal"
)

var requiredVars = []string{
	"POSTGRES_USER",
	"POSTGRES_PASSWORD",
}

func main() {
	if err := hirevec.Loadenv(".env"); err != nil {
		fmt.Println("could not load .env, using system environment")
	}
	checkPostgres()
	checkEnvVars()
	build()
}

func checkPostgres() {
	if _, err := exec.LookPath("psql"); err != nil {
		var hint string
		switch runtime.GOOS {
		case "darwin":
			hint = "brew install postgresql"
		case "linux":
			hint = "sudo apt install postgresql-client"
		default:
			hint = "https://www.postgresql.org/download/"
		}
		die("psql not found, install PostgreSQL: " + hint)
	}

	out, _ := exec.Command("psql", "--version").Output()
	fmt.Println("psql found:", strings.TrimSpace(string(out)))

	if path, err := exec.LookPath("pg_isready"); err == nil {
		host := envOr("POSTGRES_HOST", "localhost")
		port := envOr("POSTGRES_PORT", "5432")

		args := []string{"-h", host, "-p", port}
		if u := os.Getenv("POSTGRES_USER"); u != "" {
			args = append(args, "-U", u)
		}
		if d := os.Getenv("POSTGRES_DB"); d != "" {
			args = append(args, "-d", d)
		}

		if err := exec.Command(path, args...).Run(); err != nil {
			fmt.Printf("postgres not reachable at %s:%s, start it before the server needs it\n", host, port)
		} else {
			fmt.Printf("postgres is reachable at %s:%s\n", host, port)
		}
	}
}

func checkEnvVars() {
	var missing []string
	for _, v := range requiredVars {
		if os.Getenv(v) == "" {
			missing = append(missing, v)
		}
	}
	if len(missing) > 0 {
		fmt.Fprintln(os.Stderr, "missing required environment variables:")
		for _, v := range missing {
			fmt.Fprintf(os.Stderr, "  %s\n", v)
		}
		os.Exit(1)
	}
	fmt.Println("all required environment variables are set")
}

func build() {
	fmt.Println("building...")
	cmd := exec.Command("go", "build", "-ldflags=-w -s", "-o", "./bin/server", "./cmd/server/main.go")
	cmd.Env = append(os.Environ(), "CGO_ENABLED=0")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		die("build failed: " + err.Error())
	}
	fmt.Println("build successful, binary at ./bin/server")
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func die(msg string) {
	fmt.Fprintln(os.Stderr, msg)
	os.Exit(1)
}
