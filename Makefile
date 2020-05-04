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

test-all: test-unit test-memory test-redis test-rediscluster

test-unit: clear-coverage
	go test -v ./... \
		-race -covermode=atomic -coverpkg=./... -coverprofile=coverage/unit.out

test-memory: clear-coverage
	CACHE_PROVIDER=memory \
		go test ./cmd/server/handler/** -v \
		-race -covermode=atomic -coverpkg=./... -coverprofile=coverage/memory.out

test-redis: clear-coverage
	CACHE_PROVIDER=redis \
	REDIS_URL="redis://127.0.0.1:6379" \
	TTL=432000 \
		go test ./cmd/server/handler/** -v \
		-race -covermode=atomic -coverpkg=./... -coverprofile=coverage/redis.out

test-rediscluster: clear-coverage
	CACHE_PROVIDER=redis-cluster \
	REDIS_URL="redis://127.0.0.1:30001" \
	TTL=432000 \
		go test ./cmd/server/handler/** -v \
		-race -covermode=atomic -coverpkg=./... -coverprofile=coverage/redis-cluster.out

clear-coverage:
	rm -rf ./coverage || echo 'no coverage folder'
	mkdir ./coverage