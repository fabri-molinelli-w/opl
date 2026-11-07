#!/bin/sh
# Laboratorio CU-06: nginx con un certificado autofirmado de vida corta (LAB_CERT_DAYS, por defecto 20 días),
# para que OPL lo muestre como "por vencer". Solo para la demo: no es un certificado real.
set -e
DAYS="${LAB_CERT_DAYS:-20}"
if [ ! -f /etc/nginx/tls/lab.crt ]; then
  command -v openssl >/dev/null 2>&1 || apk add --no-cache openssl >/dev/null
  mkdir -p /etc/nginx/tls
  openssl req -x509 -newkey rsa:2048 -nodes -days "$DAYS" \
    -subj "/CN=lab-tls.lab.local/O=OPL Laboratorio" \
    -keyout /etc/nginx/tls/lab.key -out /etc/nginx/tls/lab.crt 2>/dev/null
fi
cat > /etc/nginx/conf.d/default.conf <<'CONF'
server {
    listen 443 ssl;
    ssl_certificate     /etc/nginx/tls/lab.crt;
    ssl_certificate_key /etc/nginx/tls/lab.key;
    location / { default_type text/plain; return 200 "OPL lab-tls\n"; }
}
CONF
exec nginx -g 'daemon off;'
