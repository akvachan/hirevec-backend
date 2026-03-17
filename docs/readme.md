# Hirevec Backend

## Philosophy

- The server strives to be simple and lightweight, we do not use any heavy fullstack frameworks on purpose.
- The server tries to follow HATEOAS philosophy, meaning that we provide next available actions in the response where it is appropriate.
- The server does not use any external build systems, package managers or shell scripts, thus trying to be as cross-platform as possible.
- The server depends only on the Go compiler, Go builder/runner and Postgres database. No Redis. No vector database. No key-value store.
- The server does follow best practices and implements RFCs wherever it can. We do not make up our own concepts or conventions.

## Setup

### Requirements
- go >= 1.25.5 
- postgres == 17.8

1. Setup required environment variables in `.env` as shown in [.example.env](.example.env).
2. Setup server (creates a new database, with a new user) dependencies:
```bash
go run cmd/setup/main.go
```
3. Run Go server:
```
go run cmd/server/main.go
```
4. Open [http://localhost:8080/health](http://localhost:8080/health).

## Cleanup
In case, for whatever reason, you want to completely remove the database and everything created by the setup script, run cleanup script:
```bash
go run cmd/cleanup/main.go
```

## API Usage (Client)
1. Generate access token (token gives access to a test user with some data binded to it already):
```
go run cmd/token/main.go
```
2. Set `ACCESS_TOKEN` either in shell environment variables or `.env`.
3. Call the script by providing the path, e.g.:
```
go run cmd/api/main.go "/v1/me/recommendations"
```
or 
```
go run cmd/api/main.go "/v1/me/recommendations/{id}/reaction" POST '{"reaction_type":"positive"}'
```

