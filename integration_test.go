package main

import (
    "context"
    "database/sql"
    "fmt"
    "log"
    "testing"
    "time"

    "github.com/stretchr/testify/assert"
    "github.com/testcontainers/testcontainers-go"
    "github.com/testcontainers/testcontainers-go/wait"
    _ "github.com/lib/pq"
)

func TestPostgresContainer(t *testing.T) {
    ctx := context.Background()

    req := testcontainers.ContainerRequest{
        Image:        "postgres:latest",
        ExposedPorts: []string{"5432/tcp"},
        Env: map[string]string{
            "POSTGRES_PASSWORD": "example",
        },
        WaitingFor: wait.ForListeningPort("5432/tcp").WithStartupTimeout(60 * time.Second),
    }

    postgresC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
        ContainerRequest: req,
        Started:          true,
    })
    if err != nil {
        log.Fatalf("Failed to start container: %v", err)
    }
    defer postgresC.Terminate(ctx)

    host, err := postgresC.Host(ctx)
    if err != nil {
        log.Fatalf("Failed to get container host: %v", err)
    }

    port, err := postgresC.MappedPort(ctx, "5432")
    if err != nil {
        log.Fatalf("Failed to get mapped port: %v", err)
    }

    dsn := fmt.Sprintf("postgres://postgres:example@%s:%s/postgres?sslmode=disable", host, port.Port())

    db, err := sql.Open("postgres", dsn)
    if err != nil {
        log.Fatalf("Failed to open database connection: %v", err)
    }
    defer db.Close()

    err = db.Ping()
    assert.NoError(t, err, "Database should be accessible")
}
