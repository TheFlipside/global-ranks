# Development

1.1 Install PostgreSQL

```sh
sudo apt-get install postgresql postgresql-contrib
```

1.2 Create the database

```sh
createdb globalranks
```

2.1 Install Go

```sh
sudo apt-get install golang
```

2.2 Build the service

```sh
go build -o bin/global-ranks ./cmd/global-ranks
```

2.3 Start the server (migrations run automatically)

```sh
GR_DB_DSN="postgres://localhost/globalranks?sslmode=disable" make run
```

3.1 Test

```sh
curl localhost:8080/api/v1/health

curl -X POST localhost:8080/api/v1/register \
  -H 'Content-Type: application/json' \
  -d '{"uuid":"550e8400-e29b-41d4-a716-446655440000"}'

curl localhost:8080/api/v1/games/demo/leaderboard
```
