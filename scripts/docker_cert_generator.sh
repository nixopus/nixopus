#!/bin/bash

IP_ADDRESS="192.168.1.3"  
CERT_PATH="/etc/docker/certs"

sudo mkdir -p $CERT_PATH

cat > server-ext.cnf << EOF
subjectAltName = IP:${IP_ADDRESS},IP:127.0.0.1,DNS:localhost
EOF

cat > client-ext.cnf << EOF
extendedKeyUsage = clientAuth
EOF

sudo openssl genrsa -out $CERT_PATH/ca-key.pem 4096
sudo openssl req -new -x509 -days 365 -key $CERT_PATH/ca-key.pem -sha256 -out $CERT_PATH/ca.pem \
    -subj "/C=US/ST=State/L=City/O=Organization/CN=$IP_ADDRESS"

sudo openssl genrsa -out $CERT_PATH/server-key.pem 4096

sudo openssl req -new -key $CERT_PATH/server-key.pem -out $CERT_PATH/server.csr \
    -subj "/CN=$IP_ADDRESS"

sudo openssl x509 -req -days 365 -sha256 \
    -in $CERT_PATH/server.csr \
    -CA $CERT_PATH/ca.pem \
    -CAkey $CERT_PATH/ca-key.pem \
    -CAcreateserial \
    -out $CERT_PATH/server-cert.pem \
    -extfile server-ext.cnf

sudo openssl genrsa -out $CERT_PATH/key.pem 4096

sudo openssl req -new -key $CERT_PATH/key.pem -out $CERT_PATH/client.csr \
    -subj "/CN=client"

sudo openssl x509 -req -days 365 -sha256 \
    -in $CERT_PATH/client.csr \
    -CA $CERT_PATH/ca.pem \
    -CAkey $CERT_PATH/ca-key.pem \
    -CAcreateserial \
    -out $CERT_PATH/cert.pem \
    -extfile client-ext.cnf

rm server-ext.cnf client-ext.cnf

sudo chmod 0444 $CERT_PATH/ca-key.pem $CERT_PATH/server-key.pem $CERT_PATH/key.pem
sudo chmod 0444 $CERT_PATH/ca.pem $CERT_PATH/server-cert.pem $CERT_PATH/cert.pem
sudo chown root:docker $CERT_PATH/*.pem

echo "Certificates generated in $CERT_PATH"
ls -l $CERT_PATH