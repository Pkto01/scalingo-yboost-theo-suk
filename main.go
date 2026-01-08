package main

import (
	"net/http"
)

func main() {
	// Sert les fichiers du dossier "static"
	http.Handle("/", http.FileServer(http.Dir("static")))

	// Lance le serveur sur http://localhost:8080
	http.ListenAndServe(":8080", nil)
}
