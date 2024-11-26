package main

import (
	"fmt"
	"go-meta/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	fmt.Println("Server running on localhost:3000")
	router.GET("/", rootHandler)

	go func() {
		if err := router.Run(":3000"); err != nil {
			fmt.Println("Error starting server:", err)
		}
	}()

	utils.TestDbConnection()

	select {}
}

func rootHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Hello, World!"})
}
