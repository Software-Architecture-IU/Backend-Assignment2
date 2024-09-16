package main

import (
    "context"
    "database/sql"
    "fmt"
    "log"
    "path/filepath"
    "testing"
    "time"

    "github.com/stretchr/testify/assert"
    "github.com/testcontainers/testcontainers-go"
    "github.com/testcontainers/testcontainers-go/wait"
    _ "github.com/lib/pq"
)

func TestPostgresContainer(t *testing.T) {
    ctx := context.Background()

    dbName := "users"
    dbUser := "user"
    dbPassword := "password"

    req := testcontainers.ContainerRequest{
        Image:        "postgres:16-alpine",
        Env:          map[string]string{"POSTGRES_DB": dbName, "POSTGRES_USER": dbUser, "POSTGRES_PASSWORD": dbPassword},
        WaitingFor:   wait.ForLog("database system is ready to accept connections").WithOccurrence(2).WithStartupTimeout(5 * time.Second),
        ExposedPorts: []string{"5432/tcp"},
        Mounts: []testcontainers.ContainerMount{
            testcontainers.BindMount(filepath.Join("testdata", "init-user-db.sh"), "/docker-entrypoint-initdb.d/init-user-db.sh"),
            testcontainers.BindMount(filepath.Join("testdata", "my-postgres.conf"), "/etc/postgresql/postgresql.conf"),
        },
    }

    postgresContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
        ContainerRequest: req,
        Started:          true,
    })

    if err != nil {
        log.Printf("failed to start container: %s", err)
        return
    }
    defer func() {
        if err := postgresContainer.Terminate(ctx); err != nil {
            log.Printf("failed to terminate container: %s", err)
        }
    }()

    host, err := postgresContainer.Host(ctx)
    if err != nil {
        log.Printf("failed to get container host: %s", err)
        return
    }

    port, err := postgresContainer.MappedPort(ctx, "5432")
    if err != nil {
        log.Printf("failed to get container port: %s", err)
        return
    }

    dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port.Port(), dbUser, dbPassword, dbName)
    db, err := sql.Open("postgres", dsn)
    if err != nil {
        log.Printf("failed to connect to database: %s", err)
        return
    }
    defer db.Close()

    // Perform your database tests here
    assert.NotNil(t, db)
}
