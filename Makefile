include .env

postgres:
	docker run --name expense-postgres -p5432:5432 -e POSTGRES_USER=${POSTGRES_USER} -e POSTGRES_PASSWORD=${POSTGRES_PASSWORD} -d postgres:alpine

createdb:
	docker exec -it expense-postgres createdb --username=${POSTGRES_USER} --owner=${POSTGRES_USER} development

dropdb:
	docker exec -it expense-postgres dropdb --username=${POSTGRES_USER} development

test:
	go test -v -cover ./...

migrateup:
	migrate -path ./migrations -database ${DB_URI} -verbose up

migrateup1:
	migrate -path ./migrations -database ${DB_URI} -verbose up 1

migratedown:
	migrate -path ./migrations -database ${DB_URI} -verbose down

migratedown1:
	migrate -path ./migrations -database ${DB_URI} -verbose down 1

serve:
	go run ./cmd/api -port=${PORT} -db-uri=${DB_URI} -cors-trusted-origins="http://localhost:4001"

.PHONY: postgres createdb migrateup migratedown migrateup1 migratedown1 dropdb test serve