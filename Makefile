BUILD_DIR=./build
MIGRATIONS_DIR=./migrations
SQLITE_DB=./strava.db

run-static:
	python -m http.server --directory dist

build-static: clean-dist
	go run cmd/genhtml/main.go

clean-dist:
	rm -rf ./dist/*

test:
	go test ./...

migrate-status:
	goose -dir $(MIGRATIONS_DIR) sqlite3 $(SQLITE_DB) status

migrate-up:
	goose -dir $(MIGRATIONS_DIR) sqlite3 $(SQLITE_DB) up

migrate-reset:
	goose -dir $(MIGRATIONS_DIR) sqlite3 $(SQLITE_DB) reset

# -- make migrate-create NAME=migration_name
migrate-create:
	goose -dir $(MIGRATIONS_DIR) sqlite3 $(SQLITE_DB) create $(NAME) sql
