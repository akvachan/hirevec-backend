# Hirevec Backend

Hirevec is a simple job application app. This project implements server for the Hirevec app.

## Setup

We support bare metal setup (downloading postgres and other dependencies is on you) and Docker setup (dependencies are handled automatically).

### Bare Metal

> [!NOTE]
> Bare metal scripts were tested only on macOS 15.7.3.

#### Requirements

- go >= 1.25.5 
- postgres == 17.8

#### Steps

1. Setup required environment variables in `.env` as shown in [.example.env](.example.env).
2. Run Go database and cache setup script:
```bash
go run cmd/setup/main.go
```
3. Generate and copy a development access token:
```bash
go run cmd/token/main.go
```
4. Run Go server:
```
go run cmd/server/main.go
```
5. Open [http://localhost:8080/v1/health](http://localhost:8080/v1/health).

> [!TIP]
> When accessing protected resources, use your development access token. You can modify certain aspects of the token and add scopes via [cmd/token/main.go](./cmd/token/main.go).

#### Cleanup

In case, for whatever reason, you want to completely remove the database and everything created by the setup script, run cleanup script:
```bash
go run cmd/cleanup/main.go
```

### Via Docker

#### Requirements

- docker >= 29.0.1

#### Steps

1. Setup required environment variables in `.env` as shown in [.example.env](.example.env).
2. Run:
```bash
docker compose up
```
3. Generate and copy a development access token:
```bash
go run cmd/token/main.go
```
4. Open [http://localhost:8080/v1/health](http://localhost:8080/v1/health).
