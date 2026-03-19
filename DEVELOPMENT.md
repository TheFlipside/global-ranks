# Development

1.1 Install PostgreSQL

sudo apt-get install postgresql postgresql-contrib

1.2 Create the database

createdb globalranks

2.1 Install Go

sudo apt-get install golang

2.2 Build the service

go build ./...

2.3 Start the server (migrations run automatically)

GR_DB_DSN="postgres://localhost/globalranks?sslmode=disable" make run

3.1 Test

curl localhost:8080/api/v1/health
curl -X POST localhost:8080/api/v1/scores \
  -H 'Content-Type: application/json' \
  -d '{"uuid":"550e8400-e29b-41d4-a716-446655440000","game_slug":"demo","score":1000}'
curl localhost:8080/api/v1/games/demo/leaderboard
