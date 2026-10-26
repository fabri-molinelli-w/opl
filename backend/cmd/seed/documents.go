package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type seedDocument struct {
	title    string
	category string
	tags     []string
	content  string
}

// Documentación ficticia para la demo (CU-05). Hosts e IPs de rangos de documentación.
var documents = []seedDocument{
	{
		title: "Reiniciar servicio Nginx", category: "runbook", tags: []string{"nginx", "web", "linux"},
		content: "## Cuándo usarlo\n\nEl portal responde 502/504 o el health check de `web-01.example.com` está DOWN.\n\n" +
			"## Pasos\n\n1. Validar la configuración antes de reiniciar:\n\n   ```bash\n   sudo nginx -t\n   ```\n\n" +
			"2. Si la validación es correcta, recargar sin cortar conexiones:\n\n   ```bash\n   sudo systemctl reload nginx\n   ```\n\n" +
			"3. Si sigue sin responder, reiniciar el servicio:\n\n   ```bash\n   sudo systemctl restart nginx\n   sudo systemctl status nginx --no-pager\n   ```\n\n" +
			"## Verificación\n\n```bash\ncurl -I https://web-01.example.com/health\n```\n\nDebe devolver `200`. Si no, revisar `/var/log/nginx/error.log` y escalar a Plataforma.\n",
	},
	{
		title: "Alta de usuario en Active Directory", category: "procedimiento", tags: []string{"active-directory", "usuarios", "windows"},
		content: "## Requisitos\n\n- Ticket aprobado por el responsable del área.\n- Nombre, apellido, legajo y grupo de seguridad.\n\n" +
			"## Pasos\n\n1. Abrir **Usuarios y equipos de Active Directory**.\n2. Ir a la OU del área (`OU=Usuarios,DC=lab,DC=local`).\n" +
			"3. Crear el usuario con el formato `nombre.apellido`.\n4. Marcar *El usuario debe cambiar la contraseña en el próximo inicio de sesión*.\n" +
			"5. Agregarlo a los grupos del área.\n\nAlternativa por PowerShell:\n\n```powershell\nNew-ADUser -Name \"Juan Perez\" -SamAccountName juan.perez " +
			"-Path \"OU=Usuarios,DC=lab,DC=local\" -ChangePasswordAtLogon $true -Enabled $true\n```\n\n" +
			"## Cierre\n\nInformar al usuario por el canal habitual (nunca enviar la contraseña por mail).\n",
	},
	{
		title: "Configurar VPN en notebook", category: "guia", tags: []string{"vpn", "notebooks", "wireguard"},
		content: "## Objetivo\n\nDejar la notebook conectada a la VPN corporativa con WireGuard.\n\n" +
			"## Pasos\n\n1. Instalar el cliente WireGuard desde el portal de software.\n2. Solicitar el archivo de perfil `.conf` a SysAdmin (se genera uno por equipo).\n" +
			"3. Importar el perfil: **Agregar túnel → Importar desde archivo**.\n4. Activar el túnel y probar:\n\n   ```bash\n   ping 192.0.2.1\n   ```\n\n" +
			"## Problemas frecuentes\n\n| Síntoma | Causa probable | Acción |\n|---|---|---|\n| No hay handshake | Puerto 51820/udp bloqueado | Probar desde otra red |\n" +
			"| Conecta pero no resuelve nombres | DNS del perfil | Verificar `DNS = 192.0.2.53` |\n",
	},
	{
		title: "Restaurar backup de PostgreSQL", category: "runbook", tags: []string{"postgresql", "backup", "db"},
		content: "## Contexto\n\nLos backups diarios de `db-01.example.com` se guardan en `/backups/postgres/` con formato `opl-AAAAMMDD.dump`.\n\n" +
			"## Pasos (staging)\n\n1. Detener las aplicaciones que usan la base.\n2. Crear una base vacía:\n\n   ```bash\n   createdb -U postgres opl_restore\n   ```\n\n" +
			"3. Restaurar el dump elegido:\n\n   ```bash\n   pg_restore -U postgres -d opl_restore --no-owner /backups/postgres/opl-20261020.dump\n   ```\n\n" +
			"4. Validar conteos de tablas críticas y renombrar la base.\n\n" +
			"> En **producción** se requiere aprobación del responsable y ventana de mantenimiento.\n",
	},
	{
		title: "Renovar certificado TLS", category: "procedimiento", tags: []string{"tls", "certificados", "nginx"},
		content: "## Cuándo\n\nCuando el dashboard marca un certificado con menos de 30 días de vigencia.\n\n" +
			"## Pasos\n\n1. Generar el CSR:\n\n   ```bash\n   openssl req -new -newkey rsa:2048 -nodes -keyout portal.key -out portal.csr " +
			"-subj \"/CN=web-01.example.com\"\n   ```\n\n2. Enviar el CSR a la CA interna y descargar el certificado.\n" +
			"3. Copiar certificado y clave a `/etc/nginx/tls/` (permisos `600` para la clave).\n4. `sudo nginx -t && sudo systemctl reload nginx`.\n\n" +
			"## Verificación\n\n```bash\necho | openssl s_client -connect web-01.example.com:443 2>/dev/null | openssl x509 -noout -dates\n```\n",
	},
	{
		title: "Impresora no imprime: checklist", category: "referencia", tags: []string{"impresoras", "mesa-de-ayuda"},
		content: "## Checklist rápido\n\n- [ ] La impresora está encendida y sin errores en el panel.\n- [ ] Tiene papel y tóner.\n" +
			"- [ ] Responde a ping (`192.0.2.80` para la impresora del piso 2).\n- [ ] La cola de impresión no tiene trabajos trabados.\n\n" +
			"## Vaciar la cola en Windows\n\n```powershell\nStop-Service Spooler\nRemove-Item C:\\Windows\\System32\\spool\\PRINTERS\\* -Force\nStart-Service Spooler\n```\n\n" +
			"Si después de esto no imprime, registrar el modelo y el error en el ticket y escalar a SysAdmin.\n",
	},
}

func seedDocuments(ctx context.Context, pool *pgxpool.Pool) error {
	var authorID string
	err := pool.QueryRow(ctx, `
		SELECT id::text FROM users
		WHERE active
		ORDER BY CASE role WHEN 'sysadmin' THEN 0 WHEN 'admin' THEN 1 ELSE 2 END, created_at
		LIMIT 1`).Scan(&authorID)
	if errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("no hay usuarios: arrancá el backend una vez para crear el admin inicial")
	}
	if err != nil {
		return fmt.Errorf("find document author: %w", err)
	}

	for _, doc := range documents {
		tx, err := pool.Begin(ctx)
		if err != nil {
			return fmt.Errorf("begin seed document: %w", err)
		}
		var id string
		err = tx.QueryRow(ctx, `
			INSERT INTO documents (title, content_md, category, tags, created_by, updated_by)
			VALUES ($1, $2, $3, $4, $5::uuid, $5::uuid)
			ON CONFLICT (lower(title)) WHERE deleted_at IS NULL DO NOTHING
			RETURNING id::text`,
			doc.title, doc.content, doc.category, doc.tags, authorID,
		).Scan(&id)
		if errors.Is(err, pgx.ErrNoRows) {
			_ = tx.Rollback(ctx)
			fmt.Printf("document already exists: %s\n", doc.title)
			continue
		}
		if err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("insert document %q: %w", doc.title, err)
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO document_versions (document_id, version, title, content_md, category, tags, edited_by)
			VALUES ($1::uuid, 1, $2, $3, $4, $5, $6::uuid)`,
			id, doc.title, doc.content, doc.category, doc.tags, authorID,
		); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("insert first version %q: %w", doc.title, err)
		}
		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("commit seed document: %w", err)
		}
		fmt.Printf("seeded document %s\n", doc.title)
	}
	return nil
}
