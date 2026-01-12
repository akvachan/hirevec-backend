// Copyright (c) 2026 Arsenii Kvachan. All Rights Reserved. MIT License.

package hirevec_test

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	hirevec "github.com/akvachan/hirevec-backend/src"
	_ "github.com/jackc/pgx/v5/stdlib"
)

var (
	testDB  *sql.DB
	baseURL string
)

func TestMain(m *testing.M) {
	root, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	err = hirevec.LoadDotEnv(filepath.Join(root, "..", ".dev.env"))
	if err != nil {
		panic(err)
	}

	hirevec.HirevecLogger = slog.New(
		slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}),
	)
	slog.SetDefault(hirevec.HirevecLogger)

	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		fmt.Println("TEST_DATABASE_URL is not set")
		os.Exit(1)
	}

	testDB, err = sql.Open("pgx", dsn)
	if err != nil {
		fmt.Println("failed to connect to test DB:", err)
		os.Exit(1)
	}
	if err := testDB.Ping(); err != nil {
		fmt.Println("failed to ping test DB:", err)
		os.Exit(1)
	}

	hirevec.HirevecDatabase = testDB

	_, err = testDB.Exec(`
	CREATE SCHEMA IF NOT EXISTS general;

	CREATE TABLE IF NOT EXISTS general.users (
		id SERIAL PRIMARY KEY,
		email VARCHAR(512) NOT NULL UNIQUE,
		user_name VARCHAR(64) NOT NULL UNIQUE,
		full_name VARCHAR(128) NOT NULL
	);

	CREATE TABLE IF NOT EXISTS general.candidates (
		id SERIAL PRIMARY KEY,
		user_id INT NOT NULL REFERENCES general.users(id) ON DELETE CASCADE,
		about TEXT NOT NULL
	);

	CREATE TABLE IF NOT EXISTS general.recruiters (
		id SERIAL PRIMARY KEY,
		user_id INT NOT NULL REFERENCES general.users(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS general.positions (
		id SERIAL PRIMARY KEY,
		title TEXT NOT NULL,
		description TEXT NOT NULL,
		company TEXT
	);

	CREATE TABLE IF NOT EXISTS general.candidates_reactions (
		candidate_id INT NOT NULL REFERENCES general.candidates(id) ON DELETE CASCADE,
		position_id INT NOT NULL REFERENCES general.positions(id) ON DELETE CASCADE,
		reaction_type reaction_type NOT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT NOW(),
		PRIMARY KEY (candidate_id, position_id)
	);

	CREATE TABLE IF NOT EXISTS general.recruiters_reactions (
		recruiter_id INT NOT NULL REFERENCES general.recruiters(id) ON DELETE CASCADE,
		position_id INT NOT NULL REFERENCES general.positions(id) ON DELETE CASCADE,
		candidate_id INT NOT NULL REFERENCES general.candidates(id) ON DELETE CASCADE,
		reaction_type reaction_type NOT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT NOW(),
		PRIMARY KEY (recruiter_id, position_id, candidate_id)
	);

	CREATE TABLE IF NOT EXISTS general.matches (
		candidate_id INT NOT NULL REFERENCES general.candidates(id) ON DELETE CASCADE,
		position_id INT NOT NULL REFERENCES general.positions(id) ON DELETE CASCADE,
		timestamp TIMESTAMP NOT NULL DEFAULT NOW(),
		PRIMARY KEY (candidate_id, position_id)
	);
	`)
	if err != nil {
		fmt.Println("failed to create schema/tables:", err)
		os.Exit(1)
	}

	ts := httptest.NewServer(hirevec.GetMainHandler())
	defer ts.Close()
	baseURL = ts.URL + "/api/v0"

	code := m.Run()
	os.Exit(code)
}

func truncateAll() {
	tables := []string{
		"matches",
		"recruiters_reactions",
		"candidates_reactions",
		"positions",
		"recruiters",
		"candidates",
		"users",
	}
	for _, tbl := range tables {
		_, _ = testDB.Exec(fmt.Sprintf("TRUNCATE TABLE general.%s CASCADE", tbl))
	}
}

func createPosition(t *testing.T, title, desc, company string) int {
	var id int
	err := testDB.QueryRow(
		`INSERT INTO general.positions (title, description, company) VALUES ($1,$2,$3) RETURNING id`,
		title, desc, company,
	).Scan(&id)
	if err != nil {
		t.Fatalf("failed to create position: %v", err)
	}
	return id
}

func TestGetPositionHandler(t *testing.T) {
	t.Cleanup(truncateAll)

	posID := createPosition(t, "Dev", "Developer role", "Acme Corp")

	resp, err := http.Get(fmt.Sprintf("%s/positions/%d", baseURL, posID))
	if err != nil {
		t.Fatalf("GET request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
	}

	var apiResp hirevec.SuccessAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		t.Fatalf("failed to decode JSON: %v", err)
	}

	data := apiResp.Data.(map[string]any)
	if data["title"] != "Dev" {
		t.Errorf("expected title 'Dev', got %v", data["title"])
	}
	if data["company"] != "Acme Corp" {
		t.Errorf("expected company 'Acme Corp', got %v", data["company"])
	}
}
