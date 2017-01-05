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
	"github.com/stretchr/gomniauth"
	"github.com/stretchr/gomniauth/providers/google"
	"github.com/stretchr/objx"
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

	data := map[string]interface{} {
		"Host": r.Host,
	}
	if authCookie, err := r.Cookie("auth"); err == nil {
		data["UserData"] = objx.MustFromBase64(authCookie.Value)
	}

	t.templ.Execute(w, data)
}

func main() {
	flag.Parse()

	gomniauth.SetSecurityKey("test-security-key")
	gomniauth.WithProviders(
		google.New(
			"686923727320-dm8goejvsbmh0pmujt4j5nc0q0t6228c.apps.googleusercontent.com",
			"J7WJX81SJ7muppIQzYdvEY7S",
			"http://localhost:8080/auth/callback/google"),
	)

	r := newRoom()
	r.tracer = trace.New(os.Stdout)

	// When you access http://localhost:8080/, following function is executed.
	http.Handle("/chat", MustAuth(&templateHandler{fileName: "index.html"}))
	http.Handle("/login", &templateHandler{fileName: "login.html"})
	http.HandleFunc("/auth/", loginHandler)
	http.Handle("/room", r)

	go r.run()

	// Start Web Server
	log.Println("Starting web server (port:", *address, ")")
	if err := http.ListenAndServe(*address, nil); err != nil {
		log.Fatal("ListenAndServer: ", err)
	}
}