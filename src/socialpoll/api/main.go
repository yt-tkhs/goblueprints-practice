package main

import (
	"net/http"
	"gopkg.in/mgo.v2"
	"flag"
	"log"
	"github.com/tylerb/graceful"
	"time"
)

var (
	addr = flag.String("addr", ":8080", "Address of Endpoint")
	mongo = flag.String("mongo", "localhost", "Address of MongoDB")
)

func main() {
	flag.Parse()

	log.Println("Connecting to MongoDB: ", *mongo)
	db ,err := mgo.Dial(*mongo)
	if err != nil {
		log.Fatalln("Failed to connect to MongoDB:", err)
	}
	defer db.Close()

	mux := http.NewServeMux()
	mux.HandleFunc("/polls/", withCORS(withVars(withDatabase(db, withAPIKey(handlePolls)))))

	log.Println("Starting Web Server:", *addr)
	graceful.Run(*addr, 1 * time.Second, mux)
	log.Println("Stopping Web Server...")
}

func withAPIKey(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("withAPIKey")
		if !isValidAPIKey(r.URL.Query().Get("key")) {
			respondErr(w, r, http.StatusUnauthorized, "Invalid API key.")
			return
		}
		fn(w, r)
	}
}

func isValidAPIKey(key string) bool {
	return key == "abc123"
}

func withDatabase(db *mgo.Session, fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("withDatabase: Start")
		thisDb := db.Copy()
		defer func() {
			log.Println("withDatabase: Close")
			thisDb.Close()
		}()
		SetVar(r, "db", thisDb.DB("ballots"))
		fn(w, r)
	}
}

func withVars(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("withVars: Start")
		OpenVars(r)
		defer func() {
			log.Println("withVars: Close")
			CloseVars(r)
		}()
		fn(w, r)
	}
}

func withCORS(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Expose-Headers", "Location")
		fn(w, r)
	}
}