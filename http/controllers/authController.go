package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go-gather/db"
	"go-gather/http/models"

	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret = []byte("your_jwt_secret_key") // TODO: move to env

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("HomeHandler Called!")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Hello, World!"})
}

func SignUp(w http.ResponseWriter, r *http.Request) {
	fmt.Println("SignUp Called!")
	if r.Method != http.MethodPost {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Invalid request method",
		})
		return
	}

	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Invalid request body",
		})
		return
	}

	// Initialize empty rooms array if not provided
	if user.Rooms == nil {
		user.Rooms = []string{}
	}

	if user.Email == "" || user.Password == "" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Email or password missing",
		})
		return
	}

	fmt.Println("User Req body: ", user)

	db := db.GetInstance()
	defer db.Close()

	if !user.CreateUser(db) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Failed to create user",
		})
		return
	}

	fmt.Println("User created successfully")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "User created successfully",
		"email":   user.Email,
	})
}

func SignIn(w http.ResponseWriter, r *http.Request) {
	fmt.Println("SignIn Called!")
	if r.Method != http.MethodPost {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Invalid request method",
		})
		return
	}

	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Invalid request body",
		})
		return
	}

	if user.Email == "" || user.Password == "" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Email or password missing",
		})
		return
	}

	fmt.Printf("Login Attempt: %+v\n", user)

	db := db.GetInstance()
	defer db.Close()

	if !user.Authenticate(db) {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email": user.Email,
		"exp":   time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}
	fmt.Println("Token Generated: ", tokenString)

	rooms := user.GetRoomsOfUser(db)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"token":   tokenString,
		"rooms":   rooms,
	})
}

func Authenticate(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Authentication successful",
		"token":   r.Header.Get("Authorization"),
	})
}
