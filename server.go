package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strings"

	_ "github.com/lib/pq"
)

const DB_URL = "postgres://lispberry:(iloveapp)@postgresql-dev-hl.default.svc.cluster.local:5432/shop_db?sslmode=disable"

func main() {
	http.HandleFunc("/", HelloServer)
	http.ListenAndServe(":8080", nil)
}

func HelloServer(w http.ResponseWriter, r *http.Request) {
	log.Print("Handling request")

	db, err := sql.Open("postgres", DB_URL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	rows, err := db.Query(`SELECT name FROM public."Brand"`)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var names []string
	for rows.Next() {
		name := ""
		rows.Scan(&name)

		names = append(names, name+"2")
	}

	fmt.Fprintf(w, "%s", strings.Join(names, "\n"))
}
