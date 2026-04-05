// Copyright (c) 2026 Arsenii Kvachan
// SPDX-License-Identifier: MIT

// Package common implements common helper functions for the scripts
package common

import (
	"bufio"
	"os"
	"os/exec"
	"strings"
)

func Loadenv(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func() {
		err := f.Close()
		if err != nil {
			Log.Error(
				"failed to properly close file",
				"err", err,
			)
			os.Exit(0)
		}
	}()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, found := strings.Cut(line, "=")
		if !found {
			continue
		}

		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)

		value = strings.Trim(value, `"'`)

		err = os.Setenv(key, value)
		if err != nil {
			return err
		}
	}

	return scanner.Err()
}

func Getenv(key string, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	return value
}

func OsUsername() (string, error) {
	out, err := exec.Command("id", "-un").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func DetectSuperuser() string {
	if v := os.Getenv("POSTGRES_SUPERUSER"); v != "" {
		return v
	}

	host := Getenv("POSTGRES_HOST", "localhost")
	port := Getenv("POSTGRES_PORT", "5432")

	candidates := []string{"postgres"}
	if u, err := OsUsername(); err == nil && u != "postgres" {
		candidates = append(candidates, u)
	}

	for _, u := range candidates {
		cmd := exec.Command("psql", "-h", host, "-p", port, "-U", u, "-d", "postgres", "-c", "SELECT 1;")
		if err := cmd.Run(); err == nil {
			return u
		}
	}

	if u, err := OsUsername(); err == nil {
		return u
	}
	return "postgres"
}

func CheckEnvVars(requiredVars []string) {
	var missing []string
	for _, v := range requiredVars {
		if os.Getenv(v) == "" {
			missing = append(missing, v)
		}
	}
	if len(missing) > 0 {
		Exit("missing required environment variables", "vars", missing)
	}
}
