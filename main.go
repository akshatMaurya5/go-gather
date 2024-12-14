package main

import (
	// "go-meta/services"
	"go-meta/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	router.GET("/", rootHandler)
	router.GET("/health", healthHandler)
	router.GET("/test-db", testDbHandler)

	go func() {
		if err := router.Run(":3000"); err != nil {
			panic("Error starting server: " + err.Error())
		}
	}()

	if err := utils.TestDbConnection(); err != nil {
		panic("Database connection failed: " + err.Error())
	}

	go func() {
		socketService := services.NewSocketService()
		socketService.Start()
	}()

	// function to test the socket service
	// services.TestSocketService()

	// Block the main goroutine to keep the server running
	select {}
}

func rootHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Hello, World!"})
}

func healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "healthy"})
}

func testDbHandler(c *gin.Context) {
	collections, err := utils.GetAllCollections()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"collections": collections})
}
