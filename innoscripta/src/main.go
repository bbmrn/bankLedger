package main

import (
	"INNOSCRIPTA/src/database"
	"INNOSCRIPTA/src/routes"
	"INNOSCRIPTA/src/util"

	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize databases
	database.InitPostgres()
	database.InitMongoDB()

	// Create Gin router
	router := gin.Default()

	// Set up routes
	routes.SetupRoutes(router)

	// Start the server
	util.Logger.Info("Server is running on :8080")
	if err := router.Run(":8080"); err != nil {
		util.Logger.Fatalf("Could not start server: %v", err)
	}
}
