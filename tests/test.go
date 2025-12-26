package tests

import (
	"log/slog"
	"net/http"
)

const url string = "http://localhost:8000/api/v0"

// TODO ./todos/261225-155400-ImplementBasicTests.md
func simpleTests() {
	slog.Info("running test #1...")
	path := url + "/positions/1"
	resp, err := http.Get(path)
	if err != nil {
		slog.Error("test #1 failed")
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
}

func main() {
	simpleTests()
}
