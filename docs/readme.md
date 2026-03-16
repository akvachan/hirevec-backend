# Hirevec Backend

## Setup

### Requirements

- go >= 1.25.5 
- postgres == 17.8

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
