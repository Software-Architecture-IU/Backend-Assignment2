package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
	"os"
	"strconv"

	_ "github.com/lib/pq"
	"github.com/gorilla/mux"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4"
)


// Message structure
type Message struct {
	ID        int       `json:"id"`
	Text      string    `json:"text"`
	Timestamp time.Time `json:"timestamp"`
}

type PostMessage struct {
	Text      string    `json:"text"`
}

type MessagesCount struct {
	Count      int    	`json:"int"`
}

func setupDB() *sql.DB {
	// Get database connection information from environment variables
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	// Create the connection string
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	// Connect to the database
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal(err)
	}

	// Check the connection
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Successfully connected to the database")

	driver, err := postgres.WithInstance(db, &postgres.Config{})
    m, err := migrate.NewWithDatabaseInstance(
        "file:///migrations",
        "postgres", driver)
	
	if err != nil {
		log.Fatal(err)
	}

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		log.Fatal("Failed to apply migrations:", err)
	}
	fmt.Println("Successfully sat up migration on the database")

	return db
}


// Add a new message to the database
func addMessage(db *sql.DB, msg PostMessage) error {
	query := `INSERT INTO messages (text, timestamp) VALUES ($1, $2) RETURNING id`

	var id int64
	err := db.QueryRow(query, msg.Text, time.Now()).Scan(&id)
	if err != nil {
		return err
	}
	fmt.Printf("Message added with ID: %d\n", id)
	return nil
}

// Retrieve all messages from the database
func getMessages(db *sql.DB, offset int) ([]Message, error) {
	// Correct SQL query with OFFSET and placeholder
	query := "SELECT id, text, timestamp FROM messages ORDER BY timestamp ASC OFFSET $1"

	// Execute the query with the OFFSET argument
	rows, err := db.Query(query, offset)
	if err != nil {
		fmt.Printf("Query error: %s\n", err)
		return nil, err
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var msg Message
		if err := rows.Scan(&msg.ID, &msg.Text, &msg.Timestamp); err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}

	// Check if there were any errors during row iteration
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return messages, nil
}

// Handler for posting a new message
func postMessageHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse the message from request body
		var msg PostMessage
		err := json.NewDecoder(r.Body).Decode(&msg)
		if err != nil || msg.Text == "" {
			http.Error(w, "Invalid message", http.StatusBadRequest)
			return
		}

		// Store the message in the database
		err = addMessage(db, msg)
		if err != nil {
			http.Error(w, "Failed to store message", http.StatusInternalServerError)
			return
		}

		// Respond with success
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"status": "Message received"})
	}
}

// Handler for retrieving all messages
func getMessagesHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		offsetParam := r.URL.Query().Get("OFFSET")

		offset, err := strconv.Atoi(offsetParam)
		if err != nil {
			// Return an error if 'OFFSET' is not a valid integer
			http.Error(w, "Invalid OFFSET parameter. Must be an integer.", http.StatusBadRequest)
			return
		}

		msgs, err := getMessages(db, offset)
		if err != nil {
			http.Error(w, "Failed to retrieve messages", http.StatusInternalServerError)
			return
		}

		// Return the messages as JSON
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(msgs)
	}
}


func getMessagesCountHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var count int
		query := "SELECT COUNT(*) FROM messages"
		
		err := db.QueryRow(query).Scan(&count)
		if err != nil {
			http.Error(w, "Failed to query the db for count..", http.StatusInternalServerError)
			return
		}
		var result MessagesCount 

		result.Count = count

		// Return the messages as JSON
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(result)
	}
}

func corsMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Set CORS headers
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

        // Handle preflight OPTIONS request
        if r.Method == "OPTIONS" {
            w.WriteHeader(http.StatusOK)
            return
        }

        // Call the next handler
        next.ServeHTTP(w, r)
    })
}

func main() {
	// Setup the database
	db := setupDB()
	defer db.Close()

	// Setup the router
	r := mux.NewRouter()

	r.HandleFunc("/messages", postMessageHandler(db)).Methods("POST")
	r.HandleFunc("/messages", getMessagesHandler(db)).Methods("GET")
	r.HandleFunc("/messages/count", getMessagesCountHandler(db)).Methods("GET")

	// Start the server
	fmt.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", corsMiddleware(r)))
}

