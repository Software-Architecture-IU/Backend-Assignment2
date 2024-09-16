package main

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	_ "github.com/lib/pq"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/stretchr/testify/assert"
)

func setupTestContainer(t *testing.T) (testcontainers.Container, *sql.DB) {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "postgres:13",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "testuser",
			"POSTGRES_PASSWORD": "testpassword",
			"POSTGRES_DB":       "testdb",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp"),
	}

	postgresContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatal(err)
	}

	host, err := postgresContainer.Host(ctx)
	if err != nil {
		t.Fatal(err)
	}

	port, err := postgresContainer.MappedPort(ctx, "5432")
	if err != nil {
		t.Fatal(err)
	}

	dsn := "postgres://testuser:testpassword@" + host + ":" + port.Port() + "/testdb?sslmode=disable"
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}

	err = db.Ping()
	if err != nil {
		t.Fatal(err)
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		t.Fatal(err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"postgres", driver)
	if err != nil {
		t.Fatal(err)
	}

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		t.Fatal("Failed to apply migrations:", err)
	}

	return postgresContainer, db
}

func TestAddMessage(t *testing.T) {
	postgresContainer, db := setupTestContainer(t)
	defer postgresContainer.Terminate(context.Background())
	defer db.Close()

	msg := PostMessage{Text: "Hello, World!"}
	err := addMessage(db, msg)
	assert.NoError(t, err)

	var id int
	var text string
	var timestamp time.Time
	err = db.QueryRow("SELECT id, text, timestamp FROM messages WHERE text=$1", msg.Text).Scan(&id, &text, &timestamp)
	assert.NoError(t, err)
	assert.Equal(t, msg.Text, text)
	assert.WithinDuration(t, time.Now(), timestamp, time.Second)
}
