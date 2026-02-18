package main

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"fmt"
	"strings"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)

type Env struct {
    db *sql.DB
}

func connectDB() (*sql.DB, error) {
    mysqlURL := os.Getenv("SCALINGO_MYSQL_URL")
    
    var dsn string
    if mysqlURL == "" {
        // En local
        dsn = "root:password@tcp(127.0.0.1:3306)/ma_bdd"
    } else {
        // Transformation de l'URL Scalingo pour le driver Go
        // mysql://user:pass@host:port/db -> user:pass@tcp(host:port)/db
        temp := strings.TrimPrefix(mysqlURL, "mysql://")
        parts := strings.Split(temp, "@")
        credentials := parts[0]
        hostAndDb := strings.Split(parts[1], "/")
        host := hostAndDb[0]
        dbName := hostAndDb[1]
        
        // On ajoute ?parseTime=true pour gérer les dates correctement en Go
        // On ajoute ?tls=true car Scalingo l'exige souvent en prod
        dsn = fmt.Sprintf("%s@tcp(%s)/%s?parseTime=true&tls=true", credentials, host, dbName)
    }

    // 1. Ouvrir la connexion (ne pas utiliser defer ici !)
    db, err := sql.Open("mysql", dsn)
    if err != nil {
        return nil, err
    }

    // 2. Vérifier la connexion
    err = db.Ping()
    if err != nil {
        return nil, err
    }

    fmt.Println("Connecté avec succès à MySQL !")
    return db, nil
}

// Handle homepage -> index.html
func homeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		errorHandler(w, r, http.StatusNotFound)
		return
	}

	tmpl, err := template.ParseFiles("static/index.html")
	if err != nil {
		log.Println("Erreur template index.html:", err)
		http.Error(w, "Erreur serveur", http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, nil)
	if err != nil {
		log.Println("Erreur exécution template index.html:", err)
	}

	
}

// Handle error
func errorHandler(w http.ResponseWriter, r *http.Request, status int) {
	w.WriteHeader(status)
	tmpl, err := template.ParseFiles("template/error.html")
	if err != nil {
		log.Println("Erreur template error.html:", err)
		http.Error(w, http.StatusText(status), status)
		return
	}

	data := struct {
		Status  int
		Message string
	}{
		Status:  status,
		Message: http.StatusText(status),
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		log.Println("Erreur exécution template error.html:", err)
	}
}

func main() {
	// Routes
	http.HandleFunc("/", homeHandler)

	// Database connection
	db, err := connectDB()
    if err != nil {
        log.Fatal("Impossible de se connecter à la BDD :", err)
    }
	defer db.Close()

	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Serve direct requests to /style.css (allows index.html to reference "style.css")
	http.Handle("/style.css", fs)

	port := "3000"
	if os.Getenv("PORT") != "" {
		port = os.Getenv("PORT")
	}

	log.Println("Serveur démarré sur http://localhost:" + port)
	log.Fatal(http.ListenAndServe(":"+port, nil))

}
