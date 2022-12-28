build:
	go build -o bin/prensio ./cmd/prensio/main.go

run:
	go run ./cmd/prensio/main.go $(args)

devenv:
	docker compose -f ./docker-compose-devenv.yaml down
	docker compose -f ./docker-compose-devenv.yaml up -d
	docker compose -f ./docker-compose-devenv.yaml ps

rundev:
	go run ./cmd/prensio/main.go run -c sample_configs/prensio_dev_default.toml

test:
	~/go/bin/ginkgo --fail-fast -vv run pipeline