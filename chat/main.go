package main

import (
	"net/http"
	"github.com/labstack/gommon/log"
)

func main() {

	// When you access http://localhost:8080/, following function is executed.
	http.HandleFunc("/", func(w http.ResponseWriter, r*http.Request) {
		w.Write([]byte(`
			<html>
			  <head>
			    <title>Chat</title>
			  </head>
			  <body>
			    Let's chat!
			  </body>
		`))
	})

	// Start Web Server
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("ListenAndServer: ", err)
	}
}