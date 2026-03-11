.PHONY: run mock-gen clean-mock test migrate-up migrate-down migrate-new seed-new

migrate-up:
	cd cmd && go run . migrate up

migrate-down:
	cd cmd && go run . migrate down --all

migrate-new:
	@if [ -z "$(name)" ]; then \
		echo "Error: name is required"; \
		echo "Usage: make migrate-new name=<migration-name>"; \
		echo "Example: make migrate-new name=create_users_table"; \
		exit 1; \
	fi
	cd cmd && go run . migrate new $(name)

seed-new:
	@if [ -z "$(name)" ]; then \
		echo "Error: name is required"; \
		echo "Usage: make seed-new name=<seeder-name>"; \
		echo "Example: make seed-new name=admin_account"; \
		exit 1; \
	fi
	cd cmd && go run . seeder new $(name)

seed:
	cd cmd && go run . seeder seed
run:
	cd cmd && go run .

run-db:
	docker compose up -d

clear-db:
	docker compose down -v --remove-orphans

mock-gen:
	mockery --all

clean-mock:
	rm -rf ./internal/mocks

test:
	go test -v -cover ./...