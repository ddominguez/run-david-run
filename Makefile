BUILD_DIR=./build
MIGRATIONS_DIR=./migrations
DB_CONNECTION=postgres postgresql://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}

run: tw-build
	go run main.go

run-static:
	python -m http.server --directory dist

build: clean tw-build
	env GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/main main.go

build-static: clean-dist tw-build
	go run cmd/genhtml/main.go

clean:
	rm -rf $(BUILD_DIR)

clean-dist:
	rm -rf ./dist/*

test:
	go test ./...

db-start:
	docker compose up -d db

db-shell:
	docker compose exec db psql -d ${DB_NAME} -h localhost -U ${DB_USER}

migrate-status:
	goose -dir $(MIGRATIONS_DIR) $(DB_CONNECTION) status

migrate-up:
	goose -dir $(MIGRATIONS_DIR) $(DB_CONNECTION) up

migrate-reset:
	goose -dir $(MIGRATIONS_DIR) $(DB_CONNECTION) reset

# -- make migrate-create NAME=migration_name
migrate-create:
	goose -dir $(MIGRATIONS_DIR) $(DB_CONNECTION) create $(NAME) sql

tw-build:
	tailwindcss -i ./css/input.css -o ./dist/styles.css --minify

tw-watch:
	tailwindcss -i ./css/input.css -o ./dist/styles.css --watch
