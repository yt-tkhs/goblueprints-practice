package main

import (
	"gopkg.in/mgo.v2"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"flag"
	"log"
	"net/http"
)

var (
	addr = flag.String("addr", ":8080", "Address of Endpoint")
	mongo = flag.String("mongo", "localhost", "Address of MongoDB")
)

type database struct {
	ses *mgo.Session
}

func main() {
	flag.Parse()

	log.Println("Connecting to MongoDB: ", *mongo)
	db ,err := mgo.Dial(*mongo)
	if err != nil {
		log.Fatalln("Failed to connect to MongoDB:", err)
	}
	defer db.Close()

	e := echo.New()
	e.Use(middleware.Recover())
	e.Use(middleware.Gzip())
	e.Use(middleware.Logger())
	e.Use(middleware.CORS()) // Cross Origin Resource Sharing

	v1 := e.Group("/api/v1")
	polls := v1.Group("/polls")

	ses := database{ses: db}
	polls.Use(withAPIKey)
	polls.Use(withVars)
	polls.Use(ses.withDatabase)

	polls.GET("", handlePollsGetAll)
	polls.GET("/:id", handlePollsGet)
	polls.POST("", handlePollsPost)
	polls.DELETE("/:id", handlePollsDelete)

	http.Handle("/", e)
	http.ListenAndServe(*addr, nil)
}

func withAPIKey(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		log.Println("withAPIKey")
		if isValidAPIKey(c.QueryParam("key")) {
			return next(c)
		}

		return c.NoContent(http.StatusUnauthorized)
	}
}

func (ses *database)withDatabase(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		log.Println("withDatabase: Start")
		thisDb := ses.ses.Copy()
		defer func() {
			log.Println("withDatabase: Close")
			thisDb.Close()
		}()
		SetVar(c.Request(), "db", thisDb.DB("ballots"))
		return next(c)
	}
}

func withVars(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		log.Println("withVars: Start")
		OpenVars(c.Request())
		defer func() {
			log.Println("withVars: Close")
			CloseVars(c.Request())
		}()
		return next(c)
	}
}

func isValidAPIKey(key string) bool {
	return key == "abc123"
}