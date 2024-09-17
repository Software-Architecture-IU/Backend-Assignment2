package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"os"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestAddMessage(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	msg := PostMessage{Text: "Hello, world!"}

	t.Run("Successful insertion", func(t *testing.T) {
		mock.ExpectQuery(`INSERT INTO messages \(text, timestamp\) VALUES \(\$1, \$2\) RETURNING id`).
			WithArgs(msg.Text, sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

		err = addMessage(db, msg)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Insert query error", func(t *testing.T) {
		mock.ExpectQuery(`INSERT INTO messages \(text, timestamp\) VALUES \(\$1, \$2\) RETURNING id`).
			WithArgs(msg.Text, sqlmock.AnyArg()).
			WillReturnError(sql.ErrConnDone)

		err = addMessage(db, msg)
		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestGetMessages(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	t.Run("Successful retrieval", func(t *testing.T) {
		mockMessages := []Message{
			{ID: 1, Text: "Hello, world!", Timestamp: time.Now().UTC()},
		}

		rows := sqlmock.NewRows([]string{"id", "text", "timestamp"}).
			AddRow(mockMessages[0].ID, mockMessages[0].Text, mockMessages[0].Timestamp)

		mock.ExpectQuery(`SELECT id, text, timestamp FROM messages ORDER BY timestamp ASC OFFSET \$1`).
			WithArgs(0).
			WillReturnRows(rows)

		messages, err := getMessages(db, 0)
		assert.NoError(t, err)
		assert.Equal(t, mockMessages, messages)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Query error", func(t *testing.T) {
		mock.ExpectQuery(`SELECT id, text, timestamp FROM messages ORDER BY timestamp ASC OFFSET \$1`).
			WithArgs(0).
			WillReturnError(sql.ErrConnDone)

		messages, err := getMessages(db, 0)
		assert.Error(t, err)
		assert.Nil(t, messages)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Scan error", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "text", "timestamp"}).
			AddRow("invalid_id", "Hello, world!", time.Now().UTC())

		mock.ExpectQuery(`SELECT id, text, timestamp FROM messages ORDER BY timestamp ASC OFFSET \$1`).
			WithArgs(0).
			WillReturnRows(rows)

		messages, err := getMessages(db, 0)
		assert.Error(t, err)
		assert.Nil(t, messages)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostMessageHandler(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	handler := postMessageHandler(db)

	t.Run("Valid message", func(t *testing.T) {
		mock.ExpectQuery(`INSERT INTO messages \(text, timestamp\) VALUES \(\$1, \$2\) RETURNING id`).
			WithArgs("Hello, world!", sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

		recorder := httptest.NewRecorder()
		requestBody := bytes.NewBufferString(`{"text":"Hello, world!"}`)
		request, _ := http.NewRequest("POST", "/messages", requestBody)

		handler.ServeHTTP(recorder, request)

		assert.Equal(t, http.StatusCreated, recorder.Code)
		assert.NoError(t, mock.ExpectationsWereMet())

		var response map[string]string
		err := json.NewDecoder(recorder.Body).Decode(&response)
		assert.NoError(t, err)
		assert.Equal(t, "Message received", response["status"])
	})

	t.Run("Invalid message - empty text", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		requestBody := bytes.NewBufferString(`{"text":""}`)
		request, _ := http.NewRequest("POST", "/messages", requestBody)

		handler.ServeHTTP(recorder, request)

		assert.Equal(t, http.StatusBadRequest, recorder.Code)
	})

	t.Run("Invalid message - malformed JSON", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		requestBody := bytes.NewBufferString(`{"text":}`)
		request, _ := http.NewRequest("POST", "/messages", requestBody)

		handler.ServeHTTP(recorder, request)

		assert.Equal(t, http.StatusBadRequest, recorder.Code)
	})

	t.Run("Database error", func(t *testing.T) {
		mock.ExpectQuery(`INSERT INTO messages \(text, timestamp\) VALUES \(\$1, \$2\) RETURNING id`).
			WithArgs("Hello, world!", sqlmock.AnyArg()).
			WillReturnError(sql.ErrConnDone)

		recorder := httptest.NewRecorder()
		requestBody := bytes.NewBufferString(`{"text":"Hello, world!"}`)
		request, _ := http.NewRequest("POST", "/messages", requestBody)

		handler.ServeHTTP(recorder, request)

		assert.Equal(t, http.StatusInternalServerError, recorder.Code)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestGetMessagesHandler(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	mockMessages := []Message{
		{ID: 1, Text: "Hello, world!", Timestamp: time.Now().UTC()},
	}

	rows := sqlmock.NewRows([]string{"id", "text", "timestamp"}).
		AddRow(mockMessages[0].ID, mockMessages[0].Text, mockMessages[0].Timestamp)

	mock.ExpectQuery(`SELECT id, text, timestamp FROM messages ORDER BY timestamp ASC OFFSET \$1`).
		WithArgs(0).
		WillReturnRows(rows)

	handler := getMessagesHandler(db)

	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/messages?OFFSET=0", nil)

	handler.ServeHTTP(recorder, request)

	assert.Equal(t, http.StatusOK, recorder.Code)

	var messages []Message
	err = json.NewDecoder(recorder.Body).Decode(&messages)
	assert.NoError(t, err)

	for i := range messages {
		messages[i].Timestamp = messages[i].Timestamp.UTC()
	}

	assert.Equal(t, mockMessages, messages)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetMessagesCountHandler(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	handler := getMessagesCountHandler(db)

	t.Run("Successful count retrieval", func(t *testing.T) {
		mock.ExpectQuery(`SELECT COUNT\(\*\) FROM messages`).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

		recorder := httptest.NewRecorder()
		request, _ := http.NewRequest("GET", "/messages/count", nil)

		handler.ServeHTTP(recorder, request)

		assert.Equal(t, http.StatusCreated, recorder.Code)

		var count MessagesCount
		err := json.NewDecoder(recorder.Body).Decode(&count)
		assert.NoError(t, err)
		assert.Equal(t, 1, count.Count)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Database error", func(t *testing.T) {
		mock.ExpectQuery(`SELECT COUNT\(\*\) FROM messages`).
			WillReturnError(sql.ErrConnDone)

		recorder := httptest.NewRecorder()
		request, _ := http.NewRequest("GET", "/messages/count", nil)

		handler.ServeHTTP(recorder, request)

		assert.Equal(t, http.StatusInternalServerError, recorder.Code)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestCORSMiddleware(t *testing.T) {
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := corsMiddleware(nextHandler)

	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("OPTIONS", "/", nil)

	handler.ServeHTTP(recorder, request)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "*", recorder.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "GET, POST, PUT, DELETE, OPTIONS", recorder.Header().Get("Access-Control-Allow-Methods"))
	assert.Equal(t, "Content-Type, Authorization", recorder.Header().Get("Access-Control-Allow-Headers"))
}

func TestMain(m *testing.M) {
	code := m.Run()
	os.Exit(code)
}
