#!/bin/bash

# TLS Certificate Generation Script
# This script generates self-signed certificates for testing TLS/mTLS

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$SCRIPT_DIR"

echo "=== Generating TLS certificates for testing ==="

# Configuration
DAYS=365
KEY_SIZE=2048

# 1. Generate CA certificate
echo "Generating CA certificate..."
openssl genrsa -out ca-key.pem $KEY_SIZE

openssl req -new -x509 -days $DAYS -key ca-key.pem -out ca.pem \
    -subj "/C=BR/ST=SP/L=Sao Paulo/O=Test CA/CN=Test CA"

# 2. Generate server certificate
echo "Generating server certificate..."
openssl genrsa -out server-key.pem $KEY_SIZE

openssl req -new -key server-key.pem -out server.csr \
    -subj "/C=BR/ST=SP/L=Sao Paulo/O=Test Server/CN=localhost"

# Create server extensions file for SAN
cat > server-ext.cnf << EOF
basicConstraints=CA:FALSE
subjectAltName=DNS:localhost,IP:127.0.0.1
extendedKeyUsage=serverAuth
EOF

openssl x509 -req -days $DAYS -in server.csr -CA ca.pem -CAkey ca-key.pem \
    -CAcreateserial -out server.pem -extfile server-ext.cnf

# 3. Generate client certificate (for mTLS)
echo "Generating client certificate..."
openssl genrsa -out client-key.pem $KEY_SIZE

openssl req -new -key client-key.pem -out client.csr \
    -subj "/C=BR/ST=SP/L=Sao Paulo/O=Test Client/CN=test-client"

# Create client extensions file
cat > client-ext.cnf << EOF
basicConstraints=CA:FALSE
extendedKeyUsage=clientAuth
EOF

openssl x509 -req -days $DAYS -in client.csr -CA ca.pem -CAkey ca-key.pem \
    -CAcreateserial -out client.pem -extfile client-ext.cnf

# Cleanup temporary files
rm -f *.csr *.cnf ca.srl

# Set permissions
chmod 644 *.pem
chmod 600 *-key.pem

echo ""
echo "=== Certificates generated successfully ==="
echo ""
echo "Files created:"
echo "  ca.pem         - CA certificate (share with clients)"
echo "  ca-key.pem     - CA private key (keep secure)"
echo "  server.pem     - Server certificate"
echo "  server-key.pem - Server private key"
echo "  client.pem     - Client certificate (for mTLS)"
echo "  client-key.pem - Client private key (for mTLS)"
echo ""
echo "To verify certificates:"
echo "  openssl x509 -in ca.pem -text -noout"
echo "  openssl x509 -in server.pem -text -noout"
echo "  openssl x509 -in client.pem -text -noout"

