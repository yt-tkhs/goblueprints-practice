package main

import (
	"net/http"
	"github.com/labstack/gommon/log"
	"sync"
	"text/template"
	"path/filepath"
)

type templateHandler struct {
	once        sync.Once
	fileName    string
	templ       *template.Template
}

// Process HTTP Request
func (t *templateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request)  {
	t.once.Do(func() {
		t.templ = template.Must(template.ParseFiles(filepath.Join("templates", t.fileName)))
	})
	t.templ.Execute(w, nil)
}

func main() {

	// When you access http://localhost:8080/, following function is executed.
	http.Handle("/", &templateHandler{fileName: "index.html"})

	// Start Web Server
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("ListenAndServer: ", err)
	}
}