#!/usr/bin/env bash
# Generate a small CA + server cert for TLS between Envoy (client) and Go services.
# SAN includes host.docker.internal so Envoy can verify when connecting from Docker.
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
CERT_DIR="${ROOT}/envoy/certs"
mkdir -p "${CERT_DIR}"

CNF="${CERT_DIR}/upstream-openssl.cnf"
cat > "${CNF}" <<'EOF'
[req]
distinguished_name = req_distinguished_name
req_extensions = v3_req
prompt = no

[req_distinguished_name]
CN = host.docker.internal

[v3_req]
subjectAltName = @alt_names
basicConstraints = CA:FALSE
keyUsage = digitalSignature, keyEncipherment
extendedKeyUsage = serverAuth

[alt_names]
DNS.1 = host.docker.internal
DNS.2 = localhost
IP.1 = 127.0.0.1
EOF

echo "==> Generating upstream CA (envoy/certs/upstream-ca.{key,crt})"
openssl genrsa -out "${CERT_DIR}/upstream-ca.key" 2048
openssl req -x509 -new -nodes -key "${CERT_DIR}/upstream-ca.key" \
  -sha256 -days 825 -out "${CERT_DIR}/upstream-ca.crt" \
  -subj "/CN=EnvoyUpstreamDevCA"

echo "==> Generating upstream server cert (signed by CA)"
openssl genrsa -out "${CERT_DIR}/upstream-server.key" 2048
openssl req -new -key "${CERT_DIR}/upstream-server.key" \
  -out "${CERT_DIR}/upstream-server.csr" -config "${CNF}"
openssl x509 -req -in "${CERT_DIR}/upstream-server.csr" \
  -CA "${CERT_DIR}/upstream-ca.crt" -CAkey "${CERT_DIR}/upstream-ca.key" -CAcreateserial \
  -out "${CERT_DIR}/upstream-server.crt" -days 825 -sha256 \
  -extensions v3_req -extfile "${CNF}"

rm -f "${CERT_DIR}/upstream-server.csr" "${CERT_DIR}/upstream-ca.srl"

echo "Done. Mount envoy/certs in Envoy; Go services use upstream-server.{crt,key} and trust is via upstream-ca.crt on Envoy."
