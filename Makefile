include .env
export DB_HOST DB_PORT DB_NAME DB_USER DB_PASSWORD DB_SSLMODE
MIGRATION_DIRS = internal/db/migrations

MIGRATION_DB = "postgresql://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=${DB_SSLMODE}"

importdb:
	docker exec -i postgres-db psql -U root -d master-golang < ./backupdb-master-golang.sql

exportdb:
	docker exec -i postgres-db pg_dump -U root -d master-golang > ./backupdb-master-golang.sql

server:
	go run .

migrate-create:
	migrate create -ext sql -dir $(MIGRATION_DIRS) -seq $(name)

migrate-up:
	migrate -path $(MIGRATION_DIRS) -database $(MIGRATION_DB) up

migrate-down:
	migrate -path $(MIGRATION_DIRS) -database $(MIGRATION_DB) down 1

migrate-force:
	migrate -path $(MIGRATION_DIRS) -database $(MIGRATION_DB) force $(version)

migrate-drop:
	migrate -path $(MIGRATION_DIRS) -database $(MIGRATION_DB) drop

migrate-goto:
	migrate -path $(MIGRATION_DIRS) -database $(MIGRATION_DB) goto $(version)

migrate-reset:
	migrate -path $(MIGRATION_DIRS) -database $(MIGRATION_DB) down
	migrate -path $(MIGRATION_DIRS) -database $(MIGRATION_DB) up
sqlc:
	sqlc generate

run-server:
	cd cmd\api && go run . 
.PHONY: importdb exportdb server migrate-create migrate-up migrate-down migrate-force migrate-drop migrate-goto migrate-reset sqlc run-server