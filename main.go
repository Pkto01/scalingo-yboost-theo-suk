package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"os"
	"github.com/xo/dburl"
	_ "github.com/go-sql-driver/mysql"
)

// 1. On utilise la structure Env pour "transporter" la connexion BDD
type Env struct {
	db *sql.DB
}

func connectDB() (*sql.DB, error) {
    url := os.Getenv("SCALINGO_MYSQL_URL")
    if url == "" {
        // En local, on utilise une URL au format standard
        url = "mysql://root:password@127.0.0.1:3306/ma_bdd"
    }

    // dburl analyse l'URL et s'occupe de la conversion proprement
    db, err := dburl.Open(url)
    if err != nil {
        return nil, err
    }

    if err := db.Ping(); err != nil {
        return nil, err
    }

    return db, nil
}

// 2. homeHandler devient une MÉTHODE de Env
func (app *Env) homeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		app.errorHandler(w, r, http.StatusNotFound)
		return
	}

	// Exemple d'utilisation de la BDD (si besoin) :
	// app.db.Query("SELECT ...")

	tmpl, err := template.ParseFiles("static/index.html")
	if err != nil {
		log.Println("Erreur template index.html:", err)
		http.Error(w, "Erreur serveur", http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}

// 3. errorHandler devient aussi une méthode de Env
func (app *Env) errorHandler(w http.ResponseWriter, r *http.Request, status int) {
	w.WriteHeader(status)
	// Attention : vérifie bien si ton dossier s'appelle "static" ou "template"
	tmpl, err := template.ParseFiles("static/error.html") 
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
	tmpl.Execute(w, data)
}

func main() {
	// A. Connexion BDD
	db, err := connectDB()
	if err != nil {
		log.Fatal("Impossible de se connecter à la BDD :", err)
	}
	defer db.Close()

	// B. On initialise notre structure avec la connexion
	app := &Env{db: db}

	// C. On utilise app.homeHandler au lieu de homeHandler
	http.HandleFunc("/", app.homeHandler)

	// Fichiers statiques
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	http.Handle("/style.css", fs)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	log.Println("Serveur démarré sur http://localhost:" + port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}