DB_URL=postgres://user:password@localhost:5432/mydb?sslmode=disable

migrateup:
    migrate -path db/migrations -database $(DB_URL) up

migratedown:
    migrate -path db/migrations -database $(DB_URL) down

.PHONY: migrateup migratedown