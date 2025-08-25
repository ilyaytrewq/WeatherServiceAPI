package main 

import (
	"net/http"
	"fmt"
)

func main() {
	http.HandleFunc("/", Handlers.handler)
	fmt.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Printf("Server failed to start: %v\n", err)
	}
}