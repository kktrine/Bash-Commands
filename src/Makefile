run:
	@go run ./cmd/bash-commands/main.go

docker_run:
	@docker compose up

docker_stop:
	@docker compose down

check:
	@go vet -vettool=$(which staticcheck -f) ./...