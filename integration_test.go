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

dbName := "users"
dbUser := "user"
dbPassword := "password"

postgresContainer, err := postgres.Run(ctx,
    "docker.io/postgres:16-alpine",
    postgres.WithDatabase(dbName),
    postgres.WithUsername(dbUser),
    postgres.WithPassword(dbPassword),
    testcontainers.WithWaitStrategy(
        wait.ForLog("database system is ready to accept connections").
            WithOccurrence(2).
            WithStartupTimeout(5*time.Second)),
)
defer func() {
    if err := testcontainers.TerminateContainer(postgresContainer); err != nil {
        log.Printf("failed to terminate container: %s", err)
    }
}()
if err != nil {
    log.Printf("failed to start container: %s", err)
    return
}
}
