package weatherservice

import (
	"fmt"
	"net/http"
	"strings"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.URL.Path {

	case "/v1/createUser":
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if err := createUser(r); err != nil {
			http.Error(w, fmt.Sprintf("RegisterUser error: %v", err), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"message": "User registered successfully"}`))

	case "/v1/changeUserData":
		if r.Method == http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if err := changeUserData(r); err != nil {
			http.Error(w, fmt.Sprintf("changeUserData error: %v", err), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "User data updated successfully"}`))

	case "/v1/getUserData":
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		userData, err := getUserData(r)
		if err != nil {
			http.Error(w, fmt.Sprintf("getUserData error: %v", err), http.StatusBadRequest)
			return
		}
		response := fmt.Sprintf(`{"email": "%s", "cities": ["%s"]}`, userData.Email, strings.Join(userData.Cities, `","`))
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))	

	case "/v1/deleteUser":
		if r.Method != http.MethodDelete {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if err := deleteUser(r); err != nil {
			http.Error(w, fmt.Sprintf("deleteUser error: %v", err), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "User deleted successfully"}`))
		
	default:
		http.Error(w, "Not found", http.StatusNotFound)
	}
}