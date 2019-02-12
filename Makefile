CIRCLE_SHA1 ?= latest

default: run

run:
	go run main.go

build:
	mkdir -p ./dist
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
		go build -v \
		-ldflags "-X main.CommitSha=${CIRCLE_SHA1}" \
		-o ./dist/buran \
		./

test:
	go test -v ./...
