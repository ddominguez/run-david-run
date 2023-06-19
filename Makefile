MIGRATIONS_DIR=./migrations
DB_CONNECTION=postgres postgres://${DB_USER}:${DB_PASSWORD}@localhost:5432/${DB_NAME}

db-start:
	docker compose up db

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
