package main

import (
	"gopkg.in/mgo.v2/bson"
	"net/http"
	"gopkg.in/mgo.v2"
	"github.com/labstack/echo"
	"log"
)

type poll struct {
	ID      bson.ObjectId `bson:"_id" json:"id"`
	Title   string `json:"title"`
	Options []string `json:"options"`
	Results map[string]int `json:"results,omitempty"`
}

func handlePollsGetAll(c echo.Context) error {
	log.Println("handlePollsGetAll")
	db := GetVar(c.Request(), "db").(*mgo.Database)
	polls := db.C("polls")
	q := polls.Find(nil)

	var result []*poll
	if err := q.All(&result); err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, result)
}

func handlePollsGet(c echo.Context) error {
	log.Println("handlePollsGet")
	db := GetVar(c.Request(), "db").(*mgo.Database)
	polls := db.C("polls")

	id := c.Param("id")
	log.Println("id:", id)
	q := polls.FindId(bson.ObjectIdHex(id))

	var result []*poll
	if err := q.All(&result); err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, result)
}

func handlePollsPost(c echo.Context) error {
	db := GetVar(c.Request(), "db").(*mgo.Database)
	polls := db.C("polls")

	var p poll
	if err := decodeBody(c.Request(), &p); err != nil {
		return c.NoContent(http.StatusBadRequest)
	}

	p.ID = bson.NewObjectId()
	if err := polls.Insert(p); err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}

	return c.Redirect(http.StatusCreated, "/polls" + p.ID.Hex())
}

func handlePollsDelete(c echo.Context) error {
	db := GetVar(c.Request(), "db").(*mgo.Database)
	polls := db.C("polls")

	id := c.Param("id")
	if err := polls.RemoveId(bson.ObjectIdHex(id)); err != nil {
		return c.JSON(http.StatusInternalServerError, "{}")
	}

	return c.NoContent(http.StatusOK)
}
