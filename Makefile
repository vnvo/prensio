build:
	go build -o bin/go-mysql-kafka ./cmd/go-mysql-kafka/main.go

run:
	go run ./cmd/go-mysql-kafka/main.go

test:
	go test -v tests/functional/pipeline_test.go -run TestMain