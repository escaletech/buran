CIRCLE_SHA1 ?= latest

default: run

run:
	go run cmd/server/main.go

build:
	mkdir -p ./dist
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
		go build -v \
		-ldflags "-X main.CommitSha=${CIRCLE_SHA1}" \
		-o ./dist/buran \
		./cmd/server

test-unit:
	go test -v ./...

test-memory:
	CACHE_PROVIDER=memory \
		go test ./cmd/server/handler/** -v

test-redis:
	CACHE_PROVIDER=redis \
	REDIS_URL="redis://127.0.0.1:6379" \
	TTL=432000 \
		go test ./cmd/server/handler/** -v

test-rediscluster:
	CACHE_PROVIDER=redis-cluster \
	REDIS_URL="redis://127.0.0.1:30001" \
	TTL=432000 \
		go test ./cmd/server/handler/** -v
