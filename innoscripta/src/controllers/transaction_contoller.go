package controllers

import (
	"context"
	"net/http"
	"time"

	"INNOSCRIPTA/src/database"
	"INNOSCRIPTA/src/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

// ProcessTransaction handles transaction processing
func ProcessTransaction(c *gin.Context) {
	var transaction models.Transaction
	if err := c.ShouldBindJSON(&transaction); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Insert transaction into MongoDB
	collection := database.MongoClient.Database("innoscripta").Collection("transactions")
	_, err := collection.InsertOne(context.Background(), transaction)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to log transaction"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Transaction processed successfully"})
}

// GetTransactionHistory retrieves transaction history (ledger details) for a user
func GetTransactionHistory(c *gin.Context) {
	userID := c.Param("id") // Get user ID from the URL parameter

	// Create a context with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Query MongoDB for transactions
	var transactions []models.Transaction
	collection := database.MongoClient.Database("innoscripta").Collection("transactions")
	filter := bson.M{"user_id": userID} // Filter by user ID

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch transactions"})
		return
	}
	defer cursor.Close(ctx)

	// Decode transactions
	for cursor.Next(ctx) {
		var transaction models.Transaction
		if err := cursor.Decode(&transaction); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode transaction"})
			return
		}
		transactions = append(transactions, transaction)
	}

	// Check for cursor errors
	if err := cursor.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Cursor error"})
		return
	}

	// Return the transaction history
	c.JSON(http.StatusOK, gin.H{
		"user_id":      userID,
		"transactions": transactions,
	})
}
