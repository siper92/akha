set shell := ["bash", "-cu"]

bin := "./bin/akha"
be_config := "config.be.yaml"
wk_config := "config.wk.yaml"
hello := "_env/examples/akha/hello.ak"

default:
    @just --list

### container

[group('up')]
up:
    docker start akha_platform.dev 2>/dev/null || docker run -d \
      --name akha_platform.dev \
      -v "{{justfile_directory()}}:/platform" \
      -w /platform \
      akha_platform.dev

[group('up')]
down:
    docker stop akha_platform.dev 2>/dev/null || true

[group('dev')]
dev: up
    docker exec -it akha_platform.dev bash

### build and test

[group('dev')]
build:
    mkdir -p bin
    go build -o {{bin}} ./cmd/akha

[group('dev')]
test:
    go test ./...

[group('dev')]
test-lang:
    go test ./lang/...

[group('dev')]
lint:
    go vet ./...
    gofmt -l .

[group('dev')]
tidy:
    go mod tidy

### run

[group('run')]
run-backend *args: build
    {{bin}} backend serve --config {{be_config}} {{args}}

[group('run')]
run-worker *args: build
    {{bin}} worker --config {{wk_config}} {{args}}

[group('run')]
run-ak file: build
    {{bin}} worker run --config {{wk_config}} {{file}}

[group('run')]
check-ak file: build
    {{bin}} worker check --config {{wk_config}} {{file}}

[group('run')]
login: build
    {{bin}} worker login --config {{wk_config}}

[group('run')]
run-hello: build
    {{bin}} worker run --config {{wk_config}} {{hello}}

[group('run')]
integration script=hello: build
    #!/usr/bin/env bash
    set -u
    {{bin}} backend serve --config {{be_config}} &
    pid=$!
    {{bin}} worker run --config {{wk_config}} {{script}}
    rc=$?
    kill $pid
    wait $pid 2>/dev/null || true
    exit $rc

### generate

[group('dev')]
clear:
    rm -rf ./sdk/*

[group('dev')]
gen: clear gen-sqlc gen-proto

[group('generate')]
gen-sqlc: up
    docker exec -i akha_platform.dev bash -c "sqlc generate && cp ./_defs/db/schema.sql ./backend/auth/schema.sql"

[group('generate')]
gen-proto: up
    docker exec -i akha_platform.dev bash -c "protoc \
      --proto_path=./_defs/proto/ \
      --go_out=./sdk \
      --go-grpc_out=./sdk \
      ./_defs/proto/*.proto"

### cleanup

[group('clean')]
clean:
    rm -rf ./bin

[group('clean')]
clean-cache:
    rm -rf ./.cache
