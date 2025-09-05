package weatherservice

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log" 
	"net/http"
	"os"

	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

type UserData struct {
	Email    string   `json:"email"`
	Password string   `json:"password,omitempty"`
	Cities   []string `json:"cities"`
}

var DB *sql.DB

func InitPostgres() error {
	host := os.Getenv("POSTGRES_HOST")
	port := os.Getenv("POSTGRES_PORT")
	user := os.Getenv("POSTGRES_USER")
	password := os.Getenv("POSTGRES_PASSWORD")
	dbname := os.Getenv("POSTGRES_DB")

	if host == "" || port == "" || user == "" || dbname == "" {
		return fmt.Errorf("postgres environment variables are not set properly")
	}

	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	var err error
	DB, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		return fmt.Errorf("failed to open Postgres: %w", err)
	}

	if err = DB.Ping(); err != nil {
		return fmt.Errorf("failed to ping Postgres: %w", err)
	}

	_, err = DB.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			email VARCHAR(255) NOT NULL PRIMARY KEY,
			password VARCHAR(255) NOT NULL,
			cities TEXT[] DEFAULT '{}'
		);
	`)
	if err != nil {
		return fmt.Errorf("failed to create users table: %w", err)
	}

	return nil
}

func createUser(r *http.Request) error {
	var userData UserData
	if err := json.NewDecoder(r.Body).Decode(&userData); err != nil {
		log.Printf("createUser: decode error: %v", err) 
		return fmt.Errorf("createUser: decode error: %w", err)
	}

	safeToLog := struct {
		Email  string
		Cities []string
	}{
		Email:  userData.Email,
		Cities: userData.Cities,
	}
	log.Printf("createUser: received user data: %+v", safeToLog)

	if userData.Email == "" || userData.Password == "" {
		return errors.New("createUser: email and password are required")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(userData.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("createUser: password hashing error: %v", err) 
		return fmt.Errorf("createUser: password hashing error: %w", err)
	}

	if err := addCitiesToDB(userData.Cities); err != nil {
		log.Printf("createUser: addCitiesToDB error: %v", err) 
		return fmt.Errorf("createUser: addCitiesToDB error: %w", err)
	}

	_, err = DB.Exec(`
		INSERT INTO users (email, password, cities)
		VALUES ($1, $2, $3)
		ON CONFLICT (email) DO NOTHING;
	`, userData.Email, string(hash), pq.Array(userData.Cities))
	if err != nil {
		log.Printf("createUser: insert error: %v", err)
		return fmt.Errorf("createUser: insert error: %w", err)
	}

	log.Printf("createUser: user %s created (or already exists)", userData.Email) // CHANGED
	return nil
}

func changeUserData(r *http.Request) error {
	var req UserData
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("changeUserData: decode error: %v", err)
		return fmt.Errorf("changeUserData: decode error: %w", err)
	}
	log.Printf("changeUserData: received request for %s, cities=%v", req.Email, req.Cities) 

	if req.Email == "" || req.Password == "" {
		return errors.New("changeUserData: email and password are required")
	}

	var storedHash string
	err := DB.QueryRow("SELECT password FROM users WHERE email=$1", req.Email).Scan(&storedHash)
	if err == sql.ErrNoRows {
		log.Printf("changeUserData: user %s not found", req.Email)
		return errors.New("changeUserData: user not found")
	}
	if err != nil {
		log.Printf("changeUserData: select error: %v", err) 
		return fmt.Errorf("changeUserData: select error: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(req.Password)); err != nil {
		log.Printf("changeUserData: incorrect password for %s", req.Email) 
		return errors.New("changeUserData: incorrect password")
	}

	if err := addCitiesToDB(req.Cities); err != nil {
		log.Printf("changeUserData: addCitiesToDB error: %v", err) 
		return fmt.Errorf("changeUserData: addCitiesToDB error: %w", err)
	}

	_, err = DB.Exec("UPDATE users SET cities = $1 WHERE email = $2", pq.Array(req.Cities), req.Email)
	if err != nil {
		log.Printf("changeUserData: update error: %v", err) 
		return fmt.Errorf("changeUserData: update error: %w", err)
	}

	log.Printf("changeUserData: user %s cities updated", req.Email) 
	return nil
}

func getUserData(r *http.Request) (UserData, error) {
	var req UserData
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("getUserData: decode error: %v", err) 
		return UserData{}, fmt.Errorf("getUserData: decode error: %w", err)
	}
	log.Printf("getUserData: request for %s", req.Email) 

	if req.Email == "" || req.Password == "" {
		return UserData{}, errors.New("getUserData: email and password are required")
	}

	var storedHash string
	var cities []string
	err := DB.QueryRow("SELECT password, cities FROM users WHERE email=$1", req.Email).Scan(&storedHash, pq.Array(&cities))
	if err == sql.ErrNoRows {
		log.Printf("getUserData: user %s not found", req.Email) 
		return UserData{}, errors.New("getUserData: user not found")
	}
	if err != nil {
		log.Printf("getUserData: select error: %v", err) 
		return UserData{}, fmt.Errorf("getUserData: select error: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(req.Password)); err != nil {
		log.Printf("getUserData: incorrect password for %s", req.Email) 
		return UserData{}, errors.New("getUserData: incorrect password")
	}

	log.Printf("getUserData: success for %s, cities=%v", req.Email, cities) 
	return UserData{
		Email:  req.Email,
		Cities: cities,
	}, nil
}

func deleteUser(r *http.Request) error {
	var req UserData
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("deleteUser: decode error: %v", err) 
		return fmt.Errorf("deleteUser: decode error: %w", err)
	}
	log.Printf("deleteUser: request for %s", req.Email) 

	if req.Email == "" || req.Password == "" {
		return errors.New("deleteUser: email and password are required")
	}

	var storedHash string
	err := DB.QueryRow("SELECT password FROM users WHERE email=$1", req.Email).Scan(&storedHash)
	if err == sql.ErrNoRows {
		log.Printf("deleteUser: user %s not found", req.Email) 
		return errors.New("deleteUser: user not found")
	}
	if err != nil {
		log.Printf("deleteUser: select error: %v", err) 
		return fmt.Errorf("deleteUser: select error: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(req.Password)); err != nil {
		log.Printf("deleteUser: incorrect password for %s", req.Email) 
		return errors.New("deleteUser: incorrect password")
	}

	_, err = DB.Exec("DELETE FROM users WHERE email=$1", req.Email)
	if err != nil {
		log.Printf("deleteUser: delete error: %v", err) 
		return fmt.Errorf("deleteUser: delete error: %w", err)
	}

	log.Printf("deleteUser: user %s deleted", req.Email) 
	return nil
}
