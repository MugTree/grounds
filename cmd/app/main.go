package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"
)

func main() {

	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("VT_DB_HOST"),
		os.Getenv("VT_DB_PORT"),
		os.Getenv("VT_DB_USER"),
		os.Getenv("VT_DB_PASSWORD"),
		os.Getenv("VT_DB_NAME"),
	)

	println(connStr)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("sql: error is on open - %v", err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatalf("sql: error is on ping - %v", err)
	}

	log.Println("Connected to Postgres")

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello from go"))
	})

	appPort := fmt.Sprintf(":%s", os.Getenv("APP_PORT"))

	log.Println("Server running on " + appPort)

	log.Fatal(http.ListenAndServe(appPort, nil))
}
