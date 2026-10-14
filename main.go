package main

import (
	"log"
	"net/http"
)

func main() {
	client, err := NewEmailClient()
	if err != nil {
		log.Fatal(err)
	}
	http.HandleFunc("/signup", func(w http.ResponseWriter, r *http.Request) { handleSignup(client, w, r) })
	log.Println("health verification service listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
