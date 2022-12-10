build:
	go build -o bin/go-mysql-kafka ./cmd/go-mysql-kafka/main.go

run:
	go run ./cmd/go-mysql-kafka/main.go

devenv:
	docker compose -f ./docker-compose-devenv.yaml down
	docker compose -f ./docker-compose-devenv.yaml up -d
	docker compose -f ./docker-compose-devenv.yaml ps

test:
	~/go/bin/ginkgo --fail-fast -vv run pipeline