[group('up')]
up:
    docker start akha_platform.dev 2>/dev/null || docker run -d \
      --name akha_platform.dev \
      -v "{{justfile_directory()}}:/akha" \
      -w /platform \
      akha_platform.dev

[group('dev')]
gen: gen-sqlc gen-proto

### sqlc code generation
[group('generate')]
gen-sqlc: up
    docker exec -it akha_platform.dev bash -c "sqlc generate"

### protobuf code generation
[group('generate')]
gen-proto: up
    docker exec -it akha_platform.dev bash -c "protoc \
      --proto_path=./_defs/proto/ \
      --go_out=./sdk \
      --go-grpc_out=./sdk \
      ./_defs/proto/*.proto"