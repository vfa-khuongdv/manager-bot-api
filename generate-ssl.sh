#!/bin/bash

# Create self-signed SSL certificate for localhost
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
    -keyout certs/localhost.key \
    -out certs/localhost.crt \
    -subj "/C=VN/ST=HoChiMinh/L=HoChiMinh/O=LocalDev/OU=IT/CN=localhost" \
    -addext "subjectAltName=DNS:localhost,DNS:*.localhost,IP:127.0.0.1"

echo "SSL certificates generated successfully!"
echo "Certificate: certs/localhost.crt"
echo "Private key: certs/localhost.key"