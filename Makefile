build:
	@go build -o bin/ecom cmd/main.go

test:
	@go test -v ./...

run: build
	@./bin/ecom

# migrate create -ext sql -dir cmd/migrate/migrations add-user-table
migration:
	@migration create -ext sql -dir cmd/migrate/migrations $(filter-out $@,$(MAKECMDGOALS))

# go run cmd/migrate/main.go up
@migrate-up:
	@go run cmd/migrate/main.go up

# go run cmd/migrate/main.go down
@migrate-down:
	@go run cmd/migrate/main.go down

dev:
	@if [ "$(ENV)" = "development" ]; then air; else echo "Use the run command for production"; fi
