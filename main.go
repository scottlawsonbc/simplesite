package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

// To understand this code, refer to these two tutorials.
//     wiki website tutorial: https://golang.org/doc/articles/wiki/
//     sqlx tutorial: https://jmoiron.github.io/sqlx/

var db *sqlx.DB

type view struct {
	Header string
	Places []place
}

// Represents a row in the database. Change this to match your database columns.
type place struct {
	Country       string
	City          sql.NullString
	TelephoneCode int `db:"telcode"`
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

// Seed the database with some example data.
func seed() {
	schema := `CREATE TABLE place (
		country text,
		city text NULL,
		telcode integer);`
	db.MustExec(schema)
	cityState := `INSERT INTO place (country, telcode) VALUES (?, ?)`
	countryCity := `INSERT INTO place (country, city, telcode) VALUES (?, ?, ?)`
	db.MustExec(cityState, "Hong Kong", 852)
	db.MustExec(cityState, "Singapore", 65)
	db.MustExec(countryCity, "South Africa", "Johannesburg", 27)
}

// Execute query and return a list of places. Panic if error.
func query(q string) []place {
	places := []place{}
	rows, err := db.Queryx("SELECT * FROM place")
	check(err)
	for rows.Next() {
		var p place
		err = rows.StructScan(&p)
		check(err)
		places = append(places, p)
	}
	return places
}

// Render the html template with the data from database.
func render(wr io.Writer, data view) {
	f, err := ioutil.ReadFile("template.html")
	check(err)
	t, err := template.New("page").Parse(string(f))
	check(err)
	t.Option("missingkey=error")
	err = t.Execute(wr, data)
	check(err)
}

// Respond to HTTP request with rendered HTML file.
func show(w http.ResponseWriter, req *http.Request) {
	data := view{
		Header: "Page Header",
		Places: query("SELECT country FROM place;"),
	}
	render(w, data)
}

func main() {
	// Connect to an in-memory sqlite database.
	// Can replace this with mongodb or any other db.
	db = sqlx.MustConnect("sqlite3", ":memory:")
	// Seed the database with initial data.
	seed()
	// Start production ready webserver.
	http.HandleFunc("/data", show)
	fmt.Println("starting webserver at http://localhost:8090")
	fmt.Println("go to http://localhost:8090/data in your browser")
	err := http.ListenAndServe(":8090", nil)
	check(err)
}
