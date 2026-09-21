FROM golang:1.27-alpine

RUN apk add --no-cache \
    protobuf-dev git \
    iputils unzip wget vim \
    bash bash-completion

RUN go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
RUN go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest

RUN apk add --no-cache gcc musl-dev sqlite-dev
RUN apk add --no-cache openssl

WORKDIR /root

COPY .bashrc /root/.bashrc
COPY ./platform/.bashrc_container /root/.bashrc_container
COPY ./platform/keys.sh /root/keys.sh

RUN chmod +x /root/keys.sh && /bin/bash /root/keys.sh

RUN date +%d/%m/%Y-%H:%M:%S > /build-timestamp && echo $(date +%s) >> /build-timestamp

WORKDIR /platform

CMD ["tail", "-f", "/dev/null"]