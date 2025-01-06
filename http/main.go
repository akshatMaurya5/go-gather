package main

import (
	"log"
	"net/http"

	"go-gather/http/routes"

	"github.com/gorilla/mux"
)

func main() {
	router := mux.NewRouter()

	// Add auth routes
	routes.AuthRoutes(router)

	// Start the HTTP server
	log.Println("Server running on port 3001")
	log.Fatal(http.ListenAndServe(":3001", router))
}
