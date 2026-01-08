package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
// Routes
	http.HandleFunc("/", homeHandler)

	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	port := "3000"
	if os.Getenv("PORT") != "" {
		port = os.Getenv("PORT")
	}

	log.Println("Serveur démarré sur http://localhost:" + port)
	log.Fatal(http.ListenAndServe(":"+port, nil))

}