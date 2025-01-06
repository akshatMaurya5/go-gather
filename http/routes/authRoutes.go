package routes

import (
	controller "go-gather/http/controllers"

	"github.com/gorilla/mux"
)

func AuthRoutes(router *mux.Router) {

	router.HandleFunc("/", controller.HomeHandler)
	router.HandleFunc("/register", controller.SignUp)
	router.HandleFunc("/login", controller.SignIn)
	router.HandleFunc("/authenticate", controller.Authenticate)

}
