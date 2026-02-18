package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/xo/dburl"
)

// 1. On utilise la structure Env pour "transporter" la connexion BDD
type Env struct {
	db *sql.DB
}

type Todo struct {
	ID        int
	Title     string
	Completed bool
}

func connectDB() (*sql.DB, error) {
	rawURL := os.Getenv("SCALINGO_MYSQL_URL")

	if rawURL == "" {
		// En local
		rawURL = "mysql://root:password@127.0.0.1:3306/ma_bdd"
	} else {
		// --- FIX SCALINGO ---
		// On enlève les paramètres existants (comme ?useSSL=true) qui font planter
		if pos := strings.Index(rawURL, "?"); pos != -1 {
			rawURL = rawURL[:pos]
		}
		// On ajoute les bons paramètres pour Go
		rawURL += "?tls=true&parseTime=true"
	}

	// dburl va maintenant parser une URL propre
	db, err := dburl.Open(rawURL)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

func (app *Env) initDB() error {
	// Requête SQL pour créer la table des tâches
	query := `
	CREATE TABLE IF NOT EXISTS todos (
		id INT AUTO_INCREMENT PRIMARY KEY,
		title VARCHAR(255) NOT NULL,
		completed BOOLEAN DEFAULT FALSE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`

	// Exécution de la requête
	_, err := app.db.Exec(query)
	if err != nil {
		return err
	}

	log.Println("Base de données initialisée (Table 'todos' prête)")
	return nil
}

func (app *Env) homeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		app.errorHandler(w, r, http.StatusNotFound)
		return
	}

	// 1. Chercher les tâches dans la BDD
	rows, err := app.db.Query("SELECT id, title, completed FROM todos ORDER BY id DESC")
	if err != nil {
		log.Println(err)
		http.Error(w, "Erreur BDD", 500)
		return
	}
	defer rows.Close()

	var todos []Todo
	for rows.Next() {
		var t Todo
		rows.Scan(&t.ID, &t.Title, &t.Completed)
		todos = append(todos, t)
	}

	// 2. Préparer les données pour le template
	data := struct {
		Todos []Todo
	}{
		Todos: todos,
	}

	// 3. Envoyer au HTML
	tmpl, _ := template.ParseFiles("static/index.html")
	tmpl.Execute(w, data)
}

func (app *Env) errorHandler(w http.ResponseWriter, r *http.Request, status int) {
	w.WriteHeader(status)
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
	// Connexion BDD
	db, err := connectDB()
	if err != nil {
		log.Fatal("Impossible de se connecter à la BDD :", err)
	}
	defer db.Close()

	// On initialise notre structure avec la connexion
	app := &Env{db: db}

	err = app.initDB()
	if err != nil {
		log.Fatal("Erreur lors de l'initialisation des tables :", err)
	}

	// On utilise app.homeHandler au lieu de homeHandler
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
