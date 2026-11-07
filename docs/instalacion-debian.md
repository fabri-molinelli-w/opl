# Instalación en producción (Debian 12/13)

Guía de la Fase 5. OPL corre con Docker Compose en un servidor de la red interna; HTTPS lo termina un
reverse proxy en el host (nginx o Caddy). El ai-service y PostgreSQL nunca quedan expuestos.

```
usuarios ──HTTPS──▶ nginx/Caddy (host :443) ──▶ 127.0.0.1:8088 frontend (nginx) ──/api──▶ backend ──▶ postgres
                                                                                        └──────▶ ai-service
```

## 1. Requisitos

```bash
sudo apt update && sudo apt install -y ca-certificates curl git
# Docker Engine + plugin compose (repositorio oficial de Docker)
curl -fsSL https://get.docker.com | sudo sh
sudo usermod -aG docker "$USER"   # cerrar sesión y volver a entrar
```

## 2. Código y configuración

```bash
sudo mkdir -p /opt/opl && sudo chown "$USER" /opt/opl
git clone https://github.com/fabri-molinelli-w/opl.git /opt/opl
cd /opt/opl
cp .env.example .env
chmod 600 .env
```

Editar `.env` (nunca se commitea):

| Variable | Valor en producción |
|---|---|
| `POSTGRES_PASSWORD` | contraseña larga y aleatoria (`openssl rand -hex 24`) |
| `JWT_SECRET` | `openssl rand -hex 32` |
| `AI_INTERNAL_TOKEN` | `openssl rand -hex 32` |
| `DEPLOY_WEBHOOK_TOKEN` | `openssl rand -hex 32` (o vacío para deshabilitar el webhook) |
| `ADMIN_EMAIL` / `ADMIN_PASSWORD` | admin inicial; cambiar la contraseña después del primer login |
| `LLM_*` | proveedor y modelo de OPSBot |

`DATABASE_URL`, `AI_SERVICE_URL` y `BACKEND_URL` los fija `docker-compose.prod.yml` (red interna).

## 3. Levantar

```bash
docker compose -f docker-compose.prod.yml up -d --build
docker compose -f docker-compose.prod.yml ps
curl -s http://127.0.0.1:8088/api/health    # {"status":"ok","database":"up"}
```

El backend aplica las migraciones al iniciar. **No** correr el seed en producción (son datos ficticios).

## 4. HTTPS con reverse proxy

### Opción A: Caddy (certificado automático si hay DNS público)

```bash
sudo apt install -y caddy
sudo tee /etc/caddy/Caddyfile >/dev/null <<'CADDY'
opl.example.com {
    reverse_proxy 127.0.0.1:8088
}
CADDY
sudo systemctl reload caddy
```

Para un dominio solo interno, usar `tls internal` dentro del bloque (CA propia de Caddy) o el certificado de la CA
de la empresa con `tls /etc/ssl/opl/fullchain.pem /etc/ssl/opl/privkey.pem`.

### Opción B: nginx con el certificado de la empresa

```nginx
# /etc/nginx/sites-available/opl
server {
    listen 80;
    server_name opl.example.com;
    return 301 https://$host$request_uri;
}
server {
    listen 443 ssl;
    server_name opl.example.com;
    ssl_certificate     /etc/ssl/opl/fullchain.pem;
    ssl_certificate_key /etc/ssl/opl/privkey.pem;
    ssl_protocols TLSv1.2 TLSv1.3;
    add_header Strict-Transport-Security "max-age=31536000" always;
    location / {
        proxy_pass http://127.0.0.1:8088;
        proxy_set_header Host $host;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto https;
        proxy_read_timeout 120s;   # OPSBot puede tardar
    }
}
```

```bash
sudo ln -s /etc/nginx/sites-available/opl /etc/nginx/sites-enabled/opl
sudo nginx -t && sudo systemctl reload nginx
```

Firewall: solo 443 (y 80 para la redirección) desde la red interna; 8088 queda en `127.0.0.1`.
El propio certificado de OPL se puede vigilar desde **Certificados** (CU-06).

## 5. Backups de PostgreSQL

```bash
sudo mkdir -p /var/backups/opl && sudo chown "$USER" /var/backups/opl
./scripts/backup-postgres.sh                     # prueba manual
crontab -e
# 15 3 * * * cd /opt/opl && ./scripts/backup-postgres.sh >> /var/log/opl-backup.log 2>&1
```

Guarda `opl-AAAAMMDD-HHMMSS.sql.gz` (permisos 600) y borra los de más de `KEEP_DAYS` (14). Copiar
periódicamente `/var/backups/opl` fuera del servidor.

### Restaurar un backup

> ⚠️ Reemplaza los datos actuales. Hacerlo en una ventana acordada y con un backup reciente a mano.

```bash
cd /opt/opl
docker compose -f docker-compose.prod.yml stop backend ai-service frontend
gunzip -c /var/backups/opl/opl-AAAAMMDD-HHMMSS.sql.gz \
  | docker compose -f docker-compose.prod.yml exec -T postgres sh -c 'psql -U "$POSTGRES_USER" -d "$POSTGRES_DB"'
docker compose -f docker-compose.prod.yml start backend ai-service frontend
curl -s http://127.0.0.1:8088/api/health
```

El dump se genera con `--clean --if-exists`, así que se puede restaurar sobre la base existente.
Probar la restauración al menos una vez por mes en un entorno de prueba. Este procedimiento conviene cargarlo
también como runbook en la documentación de OPL (CU-05) para que OPSBot lo encuentre.

## 6. Actualizar

```bash
cd /opt/opl
./scripts/backup-postgres.sh
git pull
docker compose -f docker-compose.prod.yml up -d --build
```

Opcional: avisar el deploy a OPL (CU-07) con `OPL_URL=https://opl.example.com make demo-deploy SERVICE=... ENV=prod VERSION=...`
o desde el pipeline con `.github/workflows/deploy-notify.yml`.
