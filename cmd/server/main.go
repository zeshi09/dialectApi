package main

import (
	"html/template"
	"log"
	"net/http"
	"path/filepath"
)

var templates *template.Template

func main() {
	var err error
	templates, err = template.ParseGlob(filepath.Join("templates", "*.html"))
	if err != nil {
		log.Fatalf("templates error: %v", err)
	}

	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/card", handleCard)
	http.HandleFunc("/detail", handleDetail)
	http.HandleFunc("/close", handleClose)

	log.Println("starting on localhost:9090")

	if err := http.ListenAndServe(":9090", nil); err != nil {
		log.Fatalf("starting server error: %v", err)
	}

}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	err := templates.ExecuteTemplate(w, "index.html", nil)
	if err != nil {
		http.Error(w, "500", http.StatusInternalServerError)
	}
}

func handleCard(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(`
		<div class="card">
			<h3>Карточка ` + id + `</h3>
			<p>Здесь подгружается контент с сервера 🎯</p>
		</div>
		`))
}

func handleDetail(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(`
				<div class="card-detail-overlay" id="detail-overlay">
			<button class="close-btn" onclick="closeCardSmoothly()">✖ Закрыть</button>
			<h2>Карточка ` + id + `</h2>
			<p>Подробное содержимое карточки ` + id + `</p>
		</div>
	`))
}

func handleClose(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(``))
}
