package main

import (
	"log"
	"net/http"

	"mini-imageflux-go/internal/server"
)

func main() {
	mux := http.NewServeMux()

	imageHandler := server.NewImageHandler()
	mux.HandleFunc("/image", imageHandler.HandleImage)

	log.Println("server started at http://localhost:8080")

	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}