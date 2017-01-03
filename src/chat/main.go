package main

import (
	"net/http"
	"sync"
	"text/template"
	"path/filepath"
	"flag"
	"log"
	"trace"
	"os"
)

type templateHandler struct {
	once        sync.Once
	fileName    string
	templ       *template.Template
}

var address = flag.String("address", ":8080", "address of application")

// Process HTTP Request
func (t *templateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request)  {
	t.once.Do(func() {
		t.templ = template.Must(template.ParseFiles(filepath.Join("templates", t.fileName)))
	})
	t.templ.Execute(w, r)
}

func main() {
	flag.Parse()

	r := newRoom()
	r.tracer = trace.New(os.Stdout)

	// When you access http://localhost:8080/, following function is executed.
	http.Handle("/", &templateHandler{fileName: "index.html"})
	http.Handle("/room", r)

	go r.run()

	// Start Web Server
	log.Println("Starting web server (port:", *address, ")")
	if err := http.ListenAndServe(*address, nil); err != nil {
		log.Fatal("ListenAndServer: ", err)
	}
}