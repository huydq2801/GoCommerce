# Portable commands; run all targets from the repository root.
.PHONY: run test integration vet build up stop migrate-up sqlc
run:
	go run ./cmd/api
test:
	go test ./...
integration:
	go test -v ./tests/integration
vet:
	go vet ./...
build:
	go build -o bin/api ./cmd/api
up:
	docker compose up --build --wait
stop:
	docker compose stop
migrate-up:
	docker compose run --rm migrate
sqlc:
	sqlc generate
