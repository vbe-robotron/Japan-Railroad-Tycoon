package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/you/jrt/internal/api"
	"github.com/you/jrt/internal/hub"
	"github.com/you/jrt/internal/store"
)

func main() {
	port := flag.String("port", "8080", "listen port")
	dbPath := flag.String("db", "jrt.db", "SQLite database path")
	flag.Parse()

	st, err := store.NewSQLite(*dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer st.Close()

	h := hub.New(st)
	router := api.NewRouter(h, st)

	addr := ":" + *port
	log.Printf("JRT server listening on %s (db: %s)", addr, *dbPath)
	log.Fatal(http.ListenAndServe(addr, router))
}
