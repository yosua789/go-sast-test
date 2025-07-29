# Define variables
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=assist_tix_dev
DB_SSLMODE=disable
MIGRATE_CMD=$(GOPATH)/bin/migrate

# Define the database URL
DB_URL=postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)

# Create the database
createdb:
	@echo "Creating database..."
	psql -U $(DB_USER) -h $(DB_HOST) -c "CREATE DATABASE $(DB_NAME);"
	@echo "Database created."

# Drop the database
dropdb:
	@echo "Dropping database..."
	psql -U $(DB_USER) -h $(DB_HOST) -c "DROP DATABASE IF EXISTS $(DB_NAME);"
	@echo "Database dropped."

# Run all migrations up
migrateup:
	@echo "Running migrations..."
	$(MIGRATE_CMD) -database $(DB_URL) -path ./database/migrations up
	@echo "Migrations completed."

# Roll back the last migration
migratedown:
	@echo "Rolling back last migration..."
	$(MIGRATE_CMD) -database $(DB_URL) -path ./database/migrations down 1
	@echo "Migration rolled back."

# Create a new migration file
create_migration:
	@echo "Creating new migration..."
	$(MIGRATE_CMD) create -ext sql -dir ./database/migrations -seq $(name)
	@echo "Migration created."