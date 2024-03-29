SHELL:=/bin/bash

ifdef test_run
	TEST_ARGS := -run $(test_run)
endif

migrate_up=go run main.go migrate --direction=up --step=0
migrate_down=go run main.go migrate --direction=down --step=0
run_worker_command=go run main.go worker
run_command=go run main.go server

run: check-modd-exists
	@modd -f ./.modd/server.modd.conf

lint: check-cognitive-complexity
	golangci-lint run --print-issued-lines=false --exclude-use-default=false --enable=revive --enable=goimports  --enable=unconvert --enable=unparam --concurrency=2

check-gotest:
ifeq (, $(shell which richgo))
	$(warning "richgo is not installed, falling back to plain go test")
	$(eval TEST_BIN=go test)
else
	$(eval TEST_BIN=richgo test)
endif

ifdef test_run
	$(eval TEST_ARGS := -run $(test_run))
endif
	$(eval test_command=$(TEST_BIN) ./... $(TEST_ARGS) --cover)

test-only: check-gotest mockgen
	SVC_DISABLE_CACHING=true $(test_command)

test: lint test-only

check-modd-exists:
	@modd --version > /dev/null	

run-worker: check-modd-exists
	@modd -f ./.modd/worker.modd.conf

run-telegram-bot: check-modd-exists
	@modd -f ./.modd/telegram-bot.modd.conf

check-cognitive-complexity:
	find . -type f -name '*.go' -not -name "*.pb.go" -not -name "mock*.go" -not -name "generated.go" -not -name "federation.go" \
      -exec gocognit -over 15 {} +

cacher/mock/redis.go:
		mockgen -destination=cacher/mock/mock_redis.go -package=cacher_mock github.com/sweet-go/stdlib/cacher Cacher

mail/mock/mail_utility_utility.go:
		mockgen -destination=mail/mock/mock_mail_utility_utility.go -package=mail_mock github.com/sweet-go/stdlib/mail Utility

mail/mock/mail_utility_client.go:
		mockgen -destination=mail/mock/mock_mail_utility_client.go -package=mail_mock github.com/sweet-go/stdlib/mail Client

encryption/mock/mock_jwt_token_generator.go:
		mockgen -destination=encryption/mock/mock_jwt_token_generator.go -package=encryption_mock github.com/sweet-go/stdlib/encryption JWTTokenGenerator

http/mock/mock_response.go:
		mockgen -destination=http/mock/mock_response.go -package=http_mock github.com/sweet-go/stdlib/http APIResponseGenerator

worker/mock/mock_worker_client.go:
	mockgen -destination=worker/mock/mock_worker_client.go -package=worker_mock github.com/sweet-go/stdlib/worker Client

worker/mock/mock_worker_server.go:
	mockgen -destination=worker/mock/mock_worker_server.go -package=worker_mock github.com/sweet-go/stdlib/worker Server

mockgen: cacher/mock/redis.go \
	mail/mock/mail_utility_utility.go \
	mail/mock/mail_utility_client.go \
	encryption/mock/mock_jwt_token_generator.go \
	http/mock/mock_response.go \
	worker/mock/mock_worker_client.go \
	worker/mock/mock_worker_server.go

clean:
	find -type f -name 'mock_*.go' -delete