.PHONY: run mock-gen clean-mock test migrate-up migrate-down migrate-new seed-new load-test-data load-test-profile-cpu-memory load-test-profile-wallclock load-test-profile monitor-pool

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

load-test-data:
	./scripts/load_test.sh

# pprof CPU + heap/goroutine/allocs/block/mutex (see PROFILE_SECONDS env)
load-test-profile-cpu-memory:
	./scripts/load_test_with_profile_cpu_memory.sh

# fgprof wall-clock (PROFILE_SECONDS env)
load-test-profile-wallclock:
	./scripts/load_test_with_profile_wallclock.sh

# Dispatcher: run `make load-test-profile ARGS='cpu-memory'` or `'wallclock'`
load-test-profile:
	./scripts/load_test_with_profile.sh $(ARGS)

mock-gen:
	mockery

clean-mock:
	rm -rf ./mocks

test:
	go test -v -cover ./...

monitor-pool:
	./scripts/monitor_pool.sh