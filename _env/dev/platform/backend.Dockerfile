FROM golang:1.27-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY ../../docker .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/akha ./cmd/akha

FROM alpine:3.22
RUN adduser -D -u 1000 akha && mkdir -p /data && chown akha:akha /data
COPY --from=build /out/akha /usr/local/bin/akha
COPY docker/config.be.yaml /etc/akha/config.be.yaml
USER akha
WORKDIR /data
EXPOSE 50051
ENTRYPOINT ["akha", "backend"]
CMD ["serve", "--config", "/etc/akha/config.be.yaml"]
