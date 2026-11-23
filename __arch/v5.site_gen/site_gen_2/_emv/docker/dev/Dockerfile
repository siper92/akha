FROM golang:1.23-bookworm AS go

FROM node:22-bookworm-slim

COPY --from=go /usr/local/go /usr/local/go

ENV GOPATH=/go \
    GOCACHE=/go/cache \
    PATH=/usr/local/go/bin:/go/bin:$PATH

RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates git \
    && rm -rf /var/lib/apt/lists/* \
    && npm install -g \
        tailwindcss@4 \
        @tailwindcss/cli@4 \
        @tailwindcss/postcss@4 \
        postcss@8 \
        postcss-cli@11 \
        autoprefixer@10 \
    && npm cache clean --force \
    && mkdir -p /go /src

ENV NODE_PATH=/usr/local/lib/node_modules

WORKDIR /src

VOLUME ["/go"]

RUN go version && node --version && npm --version && tailwindcss --help > /dev/null && postcss --version

CMD ["bash"]
