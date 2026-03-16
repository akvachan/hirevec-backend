# Hirevec Backend

Hirevec is a simple job application app. This project implements server for the Hirevec app.

## Setup

### Standard

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
3. Run Go server:
```
go run cmd/server/main.go
```
4. Open [http://localhost:8080/health](http://localhost:8080/health).

#### Cleanup

In case, for whatever reason, you want to completely remove the database and everything created by the setup script, run cleanup script:
```bash
go run cmd/cleanup/main.go
```
