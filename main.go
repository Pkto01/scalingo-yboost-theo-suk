package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	// Récupération du port donné par Scalingo
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // fallback local
	}

	// Sert les fichiers du dossier "static"
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/", fs)

	log.Println("Server started on port", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
	