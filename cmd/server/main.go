package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/you/jrt/internal/api"
	"github.com/you/jrt/internal/data"
	"github.com/you/jrt/internal/hub"
	"github.com/you/jrt/internal/store"
)

func main() {
	port := flag.String("port", "8080", "listen port")
	dbPath := flag.String("db", "jrt.db", "SQLite database path")
	dataDir := flag.String("data", "data", "game data directory")
	flag.Parse()

	st, err := store.NewSQLite(*dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer st.Close()

	ds, err := data.Load(*dataDir)
	if err != nil {
		log.Fatalf("データの読み込みに失敗しました: %v", err)
	}
	log.Printf("データを読み込みました: %d 都市 / %d 社", len(ds.Cities), len(ds.Companies))

	h := hub.New(st, ds)
	router := api.NewRouter(h, st)

	addr := ":" + *port
	log.Printf("JRT server listening on %s (db: %s)", addr, *dbPath)
	log.Fatal(http.ListenAndServe(addr, router))
}
