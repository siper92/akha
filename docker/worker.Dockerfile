FROM golang:1.27-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/akha ./cmd/akha

FROM alpine:3.22
RUN adduser -D -u 1000 akha && mkdir -p /data/root /scripts && chown -R akha:akha /data /scripts
COPY --from=build /out/akha /usr/local/bin/akha
COPY docker/config.wk.yaml /etc/akha/config.wk.yaml
COPY docker/scripts /scripts
USER akha
WORKDIR /data
ENTRYPOINT ["akha", "worker"]
CMD ["run", "--config", "/etc/akha/config.wk.yaml", "/scripts/hello.ak"]
