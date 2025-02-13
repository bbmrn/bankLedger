package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
	"github.com/streadway/amqp"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	postgresDB  *sql.DB
	mongoClient *mongo.Client
)

type Transaction struct {
	UserID int     `json:"user_id"`
	Amount float64 `json:"amount"`
	Type   string  `json:"type"`
}

// Initialize PostgreSQL
func initPostgres() {
	connStr := "user=youruser dbname=yourdb sslmode=disable password=yourpassword"
	var err error
	postgresDB, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Error connecting to PostgreSQL: %v", err)
	}

	err = postgresDB.Ping()
	if err != nil {
		log.Fatalf("Error pinging PostgreSQL: %v", err)
	}

	fmt.Println("Successfully connected to PostgreSQL!")
}

// Initialize MongoDB
func initMongoDB() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	var err error
	mongoClient, err = mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatalf("Error connecting to MongoDB: %v", err)
	}

	err = mongoClient.Ping(ctx, nil)
	if err != nil {
		log.Fatalf("Error pinging MongoDB: %v", err)
	}

	fmt.Println("Successfully connected to MongoDB!")
}

// Process a transaction
func processTransaction(transaction Transaction) {
	// Update PostgreSQL (account balance)
	var balance float64
	err := postgresDB.QueryRow("SELECT balance FROM users WHERE id = $1", transaction.UserID).Scan(&balance)
	if err != nil {
		log.Printf("Error fetching user balance: %v", err)
		return
	}

	if transaction.Type == "debit" {
		balance -= transaction.Amount
	} else {
		balance += transaction.Amount
	}

	_, err = postgresDB.Exec("UPDATE users SET balance = $1 WHERE id = $2", balance, transaction.UserID)
	if err != nil {
		log.Printf("Error updating user balance: %v", err)
		return
	}

	// Insert transaction into MongoDB
	collection := mongoClient.Database("innoscripta").Collection("transactions")
	_, err = collection.InsertOne(context.Background(), transaction)
	if err != nil {
		log.Printf("Error inserting transaction into MongoDB: %v", err)
		return
	}

	log.Printf("Processed transaction: %+v", transaction)
}

func main() {
	// Initialize databases
	initPostgres()
	initMongoDB()

	// Connect to RabbitMQ
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatalf("Error connecting to RabbitMQ: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Error opening RabbitMQ channel: %v", err)
	}
	defer ch.Close()

	// Declare the queue
	q, err := ch.QueueDeclare(
		"transactions", // Queue name
		false,          // Durable
		false,          // Delete when unused
		false,          // Exclusive
		false,          // No-wait
		nil,            // Arguments
	)
	if err != nil {
		log.Fatalf("Error declaring queue: %v", err)
	}

	// Consume messages
	msgs, err := ch.Consume(
		q.Name, // Queue
		"",     // Consumer
		true,   // Auto-ack
		false,  // Exclusive
		false,  // No-local
		false,  // No-wait
		nil,    // Args
	)
	if err != nil {
		log.Fatalf("Error registering consumer: %v", err)
	}

	log.Println("Waiting for transactions...")

	// Process messages
	for msg := range msgs {
		var transaction Transaction
		err := json.Unmarshal(msg.Body, &transaction)
		if err != nil {
			log.Printf("Error decoding transaction: %v", err)
			continue
		}

		processTransaction(transaction)
	}
}
