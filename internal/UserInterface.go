package weatherservice

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
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
		return fmt.Errorf("createUser: decode error: %w", err)
	}
	fmt.Println("createUser: received user data:", userData)
	if userData.Email == "" || userData.Password == "" {
		return errors.New("createUser: email and password are required")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(userData.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("createUser: password hashing error: %w", err)
	}

	if err := addCitiesToDB(userData.Cities); err != nil {
		return fmt.Errorf("createUser: addCitiesToDB error: %w", err)
	}

	_, err = DB.Exec(`
		INSERT INTO users (email, password, cities)
		VALUES ($1, $2, $3)
		ON CONFLICT (email) DO NOTHING;
	`, userData.Email, string(hash), pq.Array(userData.Cities))
	if err != nil {
		return fmt.Errorf("createUser: insert error: %w", err)
	}

	return nil
}

func changeUserData(r *http.Request) error {
	var req UserData
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return fmt.Errorf("changeUserData: decode error: %w", err)
	}
	if req.Email == "" || req.Password == "" {
		return errors.New("changeUserData: email and password are required")
	}

	var storedHash string
	err := DB.QueryRow("SELECT password FROM users WHERE email=$1", req.Email).Scan(&storedHash)
	if err == sql.ErrNoRows {
		return errors.New("changeUserData: user not found")
	}
	if err != nil {
		return fmt.Errorf("changeUserData: select error: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(req.Password)); err != nil {
		return errors.New("changeUserData: incorrect password")
	}

	if err := addCitiesToDB(req.Cities); err != nil {
		return fmt.Errorf("changeUserData: addCitiesToDB error: %w", err)
	}

	_, err = DB.Exec("UPDATE users SET cities = $1 WHERE email = $2", pq.Array(req.Cities), req.Email)
	if err != nil {
		return fmt.Errorf("changeUserData: update error: %w", err)
	}

	return nil
}

func getUserData(r *http.Request) (UserData, error) {
	var req UserData
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return UserData{}, fmt.Errorf("getUserData: decode error: %w", err)
	}
	if req.Email == "" || req.Password == "" {
		return UserData{}, errors.New("getUserData: email and password are required")
	}

	var storedHash string
	var cities []string
	err := DB.QueryRow("SELECT password, cities FROM users WHERE email=$1", req.Email).Scan(&storedHash, pq.Array(&cities))
	if err == sql.ErrNoRows {
		return UserData{}, errors.New("getUserData: user not found")
	}
	if err != nil {
		return UserData{}, fmt.Errorf("getUserData: select error: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(req.Password)); err != nil {
		return UserData{}, errors.New("getUserData: incorrect password")
	}

	return UserData{
		Email:  req.Email,
		Cities: cities,
	}, nil
}

func deleteUser(r *http.Request) error {
	var req UserData
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return fmt.Errorf("deleteUser: decode error: %w", err)
	}
	if req.Email == "" || req.Password == "" {
		return errors.New("deleteUser: email and password are required")
	}

	var storedHash string
	err := DB.QueryRow("SELECT password FROM users WHERE email=$1", req.Email).Scan(&storedHash)
	if err == sql.ErrNoRows {
		return errors.New("deleteUser: user not found")
	}
	if err != nil {
		return fmt.Errorf("deleteUser: select error: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(req.Password)); err != nil {
		return errors.New("deleteUser: incorrect password")
	}

	_, err = DB.Exec("DELETE FROM users WHERE email=$1", req.Email)
	if err != nil {
		return fmt.Errorf("deleteUser: delete error: %w", err)
	}

	return nil
}
