package models

import (
	"database/sql"
	"log"

	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Email    string   `json:"email"`
	Password string   `json:"password"`
	Rooms    []string `json:"rooms"`
}

func (u *User) CreateUser(db *sql.DB) bool {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Println("Password hashing failed", err)
		return false
	}

	query := `INSERT INTO users (email, password, rooms) VALUES ($1, $2, $3)`
	_, err = db.Exec(query, u.Email, hashedPassword, pq.Array(u.Rooms))
	if err != nil {
		log.Println("User creation failed", err)
		return false
	}

	log.Println("User created successfully")
	return true
}

func (u *User) Authenticate(db *sql.DB) bool {
	var hashedPassword string
	err := db.QueryRow("SELECT password FROM users WHERE email = $1", u.Email).Scan(&hashedPassword)
	if err != nil {
		log.Println("User not found", err)
		return false
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(u.Password))
	if err != nil {
		log.Println("Password does not match", err)
		return false
	}

	log.Println("User authenticated successfully!")
	return true
}
func (u *User) GetRoomsOfUser(db *sql.DB) []string {
	username := u.Email

	if username == "" {
		log.Println("Empty email")
		return []string{}
	}

	if db == nil {
		log.Println("Database connection is nil")
		return []string{}
	}

	log.Printf("Getting rooms for user %s", username)
	var rooms []string
	query := `SELECT rooms FROM users WHERE email=$1`
	err := db.QueryRow(query, username).Scan(pq.Array(&rooms))
	if err != nil {
		if err == sql.ErrConnDone {
			log.Printf("Database connection is closed for user %s", username)
		} else {
			log.Printf("Error querying rooms for user %s: %v", username, err)
		}
		return []string{}
	}

	return rooms
}
