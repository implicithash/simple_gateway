package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Response struct
type Response struct {
	Description string `json:"description"`
	Message     string `json:"message"`
}

func handler(w http.ResponseWriter, r *http.Request) {
	data := &Response{Description: "Testing API", Message: "Hello world!"}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(data)
}

func main() {
	http.HandleFunc("/", handler)

	fmt.Println("starting server at :8001")
	http.ListenAndServe(":8001", nil)
}
