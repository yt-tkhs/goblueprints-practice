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

var address = flag.String("address", ":8080", "address of application")
var avatars Avatar = TryAvatars {
	UseFileSystemAvatar,
	UseAuthAvatar,
	UseGravatarAvatar,
}

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
			GoogleClientID,
			GoogleClientSecret,
			"http://localhost:8080/auth/callback/google"),
	)

	r := newRoom()
	r.tracer = trace.New(os.Stdout)

	http.Handle("/chat", MustAuth(&templateHandler{fileName: "index.html"}))
	http.Handle("/login", &templateHandler{fileName: "login.html"})
	http.Handle("/upload", &templateHandler{fileName: "upload.html"})
	http.HandleFunc("/uploader", uploaderHandler)
	http.HandleFunc("/auth/", loginHandler)
	http.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{
			Name: "auth",
			Value: "",
			Path: "/",
			MaxAge: -1,
		})
		w.Header()["Location"] = []string{"/chat"}
		w.WriteHeader(http.StatusTemporaryRedirect)
	})
	http.Handle("/avatars/", http.StripPrefix("/avatars/", http.FileServer(http.Dir("./avatars"))))

	http.Handle("/room", r)

	go r.run()

	// Start Web Server
	log.Println("Starting web server (port:", *address, ")")
	if err := http.ListenAndServe(*address, nil); err != nil {
		log.Fatal("ListenAndServer: ", err)
	}
}