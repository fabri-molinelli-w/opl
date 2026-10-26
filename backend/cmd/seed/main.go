package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type seedServer struct {
	hostname    string
	ip          string
	environment string
	os          string
	location    string
	owner       string
	tags        []string
	services    []seedService
}

type seedService struct {
	name               string
	serviceType        string
	port               int
	checkType          string
	checkTarget        string
	expectedStatus     int
	intervalSeconds    int
	latencyThresholdMs int
	currentStatus      string
	// monitored=false deja el monitoreo pausado (check_enabled): los servidores ficticios no
	// existen, así que se muestra el estado sembrado. Solo el laboratorio se chequea de verdad.
	monitored bool
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "seed:", err)
		os.Exit(1)
	}
}

func run() error {
	url := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if url == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		return fmt.Errorf("connect database: %w", err)
	}
	defer pool.Close()

	servers := []seedServer{
		{
			hostname: "web-01.example.com", ip: "192.0.2.10", environment: "prod",
			os: "debian-12", location: "onprem", owner: "Plataforma",
			tags: []string{"web", "critical"},
			services: []seedService{
				{name: "Portal", serviceType: "web", port: 443, checkType: "http",
					checkTarget: "https://web-01.example.com/health", expectedStatus: 200,
					intervalSeconds: 30, latencyThresholdMs: 1000, currentStatus: "UP"},
			},
		},
		{
			hostname: "db-01.example.com", ip: "198.51.100.20", environment: "prod",
			os: "debian-12", location: "onprem", owner: "Base de datos",
			tags: []string{"db", "critical"},
			services: []seedService{
				{name: "PostgreSQL", serviceType: "db", port: 5432, checkType: "tcp",
					checkTarget: "db-01.example.com:5432", expectedStatus: 200,
					intervalSeconds: 30, latencyThresholdMs: 500, currentStatus: "UP"},
			},
		},
		{
			hostname: "queue-01.example.com", ip: "203.0.113.30", environment: "staging",
			os: "debian-12", location: "cloud", owner: "Mensajería",
			tags: []string{"queue", "staging"},
			services: []seedService{
				{name: "RabbitMQ", serviceType: "queue", port: 5672, checkType: "tcp",
					checkTarget: "queue-01.example.com:5672", expectedStatus: 200,
					intervalSeconds: 60, latencyThresholdMs: 1500, currentStatus: "DEGRADED"},
			},
		},
		{
			hostname: "app-01.example.com", ip: "198.51.100.30", environment: "prod",
			os: "debian-12", location: "onprem", owner: "Aplicaciones",
			tags: []string{"app", "critical"},
			services: []seedService{
				{name: "API interna", serviceType: "web", port: 8080, checkType: "http",
					checkTarget: "http://app-01.example.com:8080/health", expectedStatus: 200,
					intervalSeconds: 30, latencyThresholdMs: 800, currentStatus: "UP"},
			},
		},
		{
			hostname: "lab-01.lab.local", ip: "192.0.2.100", environment: "dev",
			os: "debian-12", location: "onprem", owner: "Laboratorio OPL",
			tags: []string{"lab", "demo"},
			services: []seedService{
				{name: "lab-nginx", serviceType: "web", port: 8081, checkType: "http",
					checkTarget: "http://localhost:8081/health", expectedStatus: 200,
					intervalSeconds: 10, latencyThresholdMs: 500, currentStatus: "UNKNOWN", monitored: true},
			},
		},
		{
			hostname: "cache-01.example.com", ip: "203.0.113.40", environment: "dev",
			os: "debian-12", location: "cloud", owner: "Plataforma",
			tags: []string{"cache", "dev"},
			services: []seedService{
				{name: "Redis", serviceType: "db", port: 6379, checkType: "tcp",
					checkTarget: "cache-01.example.com:6379", expectedStatus: 200,
					intervalSeconds: 60, latencyThresholdMs: 300, currentStatus: "UNKNOWN"},
			},
		},
	}

	for _, server := range servers {
		var serverID string
		err := pool.QueryRow(ctx, `
			INSERT INTO servers (hostname, ip, environment, os, location, owner, tags)
			VALUES ($1, $2::inet, $3, $4, $5, $6, $7)
			ON CONFLICT (lower(hostname)) WHERE deleted_at IS NULL
			DO UPDATE SET
				ip = EXCLUDED.ip,
				environment = EXCLUDED.environment,
				os = EXCLUDED.os,
				location = EXCLUDED.location,
				owner = EXCLUDED.owner,
				tags = EXCLUDED.tags,
				updated_at = now()
			RETURNING id::text`,
			server.hostname, server.ip, server.environment, server.os,
			server.location, server.owner, server.tags,
		).Scan(&serverID)
		if err != nil {
			return fmt.Errorf("upsert server %s: %w", server.hostname, err)
		}

		for _, item := range server.services {
			if _, err := pool.Exec(ctx, `
				INSERT INTO services (
					server_id, name, type, port, check_type, check_target,
					expected_status, interval_seconds, latency_threshold_ms, current_status, check_enabled
				)
				SELECT $1::uuid, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
				WHERE NOT EXISTS (
					SELECT 1 FROM services
					WHERE server_id = $1::uuid AND name = $2 AND deleted_at IS NULL
				)`,
				serverID, item.name, item.serviceType, item.port, item.checkType, item.checkTarget,
				item.expectedStatus, item.intervalSeconds, item.latencyThresholdMs, item.currentStatus, item.monitored,
			); err != nil {
				return fmt.Errorf("insert service %s/%s: %w", server.hostname, item.name, err)
			}
			// Bases sembradas antes de CU-03: se pausa el monitoreo de los servicios ficticios.
			if _, err := pool.Exec(ctx, `
				UPDATE services
				SET check_enabled = $3,
				    current_status = CASE WHEN $3 THEN current_status ELSE $4 END,
				    consecutive_failures = CASE WHEN $3 THEN consecutive_failures ELSE 0 END
				WHERE server_id = $1::uuid AND name = $2 AND deleted_at IS NULL`,
				serverID, item.name, item.monitored, item.currentStatus,
			); err != nil {
				return fmt.Errorf("update service monitoring %s/%s: %w", server.hostname, item.name, err)
			}
		}
		fmt.Printf("seeded %s\n", server.hostname)
	}
	return seedDocuments(ctx, pool)
}
