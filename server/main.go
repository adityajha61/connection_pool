package main

import (
	"fmt"
	"net/http"
	"time"
)

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("waiting for 2 seconds")
	time.Sleep(2*time.Second)
	fmt.Fprintf(w, "Hello, World!")
	fmt.Println("exiting now")
}

func main() {
	http.HandleFunc("/", handler)

	port := 8080
	fmt.Printf("Server is running on port %d...\n", port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
}
