package weatherservice

import (
	"fmt"
	"net/http"
	"strings"
	"log"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	log.Printf("Handler: %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)

	switch r.URL.Path {

	case "/v1/createUser":
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			log.Printf("Handler: wrong method %s for %s", r.Method, r.URL.Path)
			return
		}
		if err := createUser(r); err != nil {
			log.Printf("Handler: createUser error: %v", err)
			http.Error(w, fmt.Sprintf("RegisterUser error: %v", err), http.StatusBadRequest)
			return
		}
		log.Printf("Handler: user created successfully")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"message": "User registered successfully"}`))

	case "/v1/changeUserData":
		if r.Method == http.MethodGet {
			log.Printf("Handler: wrong method %s for %s", r.Method, r.URL.Path)
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if err := changeUserData(r); err != nil {
			log.Printf("Handler: changeUserData error: %v", err)
           	http.Error(w, fmt.Sprintf("changeUserData error: %v", err), http.StatusBadRequest)
			return
		}
		log.Printf("Handler: user data updated successfully")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "User data updated successfully"}`))

	 case "/v1/getUserData":
        if r.Method != http.MethodPost {
            log.Printf("Handler: wrong method %s for %s", r.Method, r.URL.Path)
            http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
            return
        }
        userData, err := getUserData(r)
        if err != nil {
            log.Printf("Handler: getUserData error: %v", err)
            http.Error(w, fmt.Sprintf("getUserData error: %v", err), http.StatusBadRequest)
            return
        }
        log.Printf("Handler: user data fetched for %s", userData.Email)
        response := fmt.Sprintf(`{"email": "%s", "cities": ["%s"]}`, userData.Email, strings.Join(userData.Cities, `","`))
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(response))

    case "/v1/deleteUser":
        if r.Method != http.MethodDelete {
            log.Printf("Handler: wrong method %s for %s", r.Method, r.URL.Path)
            http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
            return
        }
        if err := deleteUser(r); err != nil {
            log.Printf("Handler: deleteUser error: %v", err)
            http.Error(w, fmt.Sprintf("deleteUser error: %v", err), http.StatusBadRequest)
            return
        }
        log.Printf("Handler: user deleted successfully")
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(`{"message": "User deleted successfully"}`))

    default:
        log.Printf("Handler: not found %s %s", r.Method, r.URL.Path)
        http.Error(w, "Not found", http.StatusNotFound)
    }
}