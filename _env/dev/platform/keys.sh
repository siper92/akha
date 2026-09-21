#!/bin/bash

openssl genrsa -out ca.key 2048
openssl req -new -x509 -days 3650 -key ca.key -out ca.crt \
  -subj "/CN=BeizCA"

# Generate server cert with SAN
openssl genrsa -out server.key 2048

cat > server.conf <<EOF
[req]
distinguished_name = req_distinguished_name
req_extensions = v3_req
prompt = no

[req_distinguished_name]
CN = localhost

[v3_req]
subjectAltName = DNS:localhost,DNS:127.0.0.1,IP:127.0.0.1
EOF

openssl req -new -key server.key -out server.csr \
  -config server.conf -extensions v3_req

openssl x509 -req -days 365 -in server.csr \
  -CA ca.crt -CAkey ca.key -CAcreateserial \
  -out server.crt -extensions v3_req \
  -extfile server.conf

# Generate CLI client cert with SAN

openssl genrsa -out cli-client.key 2048

cat > client.conf <<EOF
[req]
distinguished_name = req_distinguished_name
req_extensions = v3_req
prompt = no

[req_distinguished_name]
CN = beiz-cli

[v3_req]
subjectAltName = DNS:beiz-cli
EOF

openssl req -new -key cli-client.key -out cli-client.csr \
  -config client.conf -extensions v3_req

# Sign client cert
openssl x509 -req -days 365 -in cli-client.csr \
  -CA ca.crt -CAkey ca.key -CAcreateserial \
  -out cli-client.crt -extensions v3_req \
  -extfile client.conf

# Cleanup
rm -f server.csr server.conf client.csr client.conf