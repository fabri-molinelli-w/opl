#!/bin/sh
# Backup diario de PostgreSQL (Fase 5). Guarda un dump comprimido y conserva los últimos KEEP_DAYS días.
# Uso (desde la raíz del repo, en el servidor):  ./scripts/backup-postgres.sh
# Cron (todos los días 03:15):
#   15 3 * * * cd /opt/opl && ./scripts/backup-postgres.sh >> /var/log/opl-backup.log 2>&1
# Restauración: ver docs/instalacion-debian.md ("Restaurar un backup").
set -eu
COMPOSE_FILE="${COMPOSE_FILE:-docker-compose.prod.yml}"
BACKUP_DIR="${BACKUP_DIR:-/var/backups/opl}"
KEEP_DAYS="${KEEP_DAYS:-14}"

# Solo se leen las variables de la base (no se imprime nada del .env).
POSTGRES_USER=$(grep -E '^POSTGRES_USER=' .env | cut -d= -f2-)
POSTGRES_DB=$(grep -E '^POSTGRES_DB=' .env | cut -d= -f2-)

mkdir -p "$BACKUP_DIR"
FILE="$BACKUP_DIR/opl-$(date +%Y%m%d-%H%M%S).sql.gz"
docker compose -f "$COMPOSE_FILE" exec -T postgres pg_dump -U "$POSTGRES_USER" -d "$POSTGRES_DB" --clean --if-exists \
  | gzip > "$FILE"
chmod 600 "$FILE"
find "$BACKUP_DIR" -name 'opl-*.sql.gz' -mtime +"$KEEP_DAYS" -delete
echo "$(date -Is) backup OK: $FILE"
