package main

import (
	"database/sql"
	"log"

	"github.com/jeffsupersixer/jeff_go_postgres/util"
	_ "github.com/lib/pq"

	"github.com/jeffsupersixer/jeff_go_postgres/api"
	db "github.com/jeffsupersixer/jeff_go_postgres/db/sqlc"
)

func main() {
	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal("cannot load configurations:", err)
	}
	conn, err := sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}

	store := db.NewStore(conn)
	server := api.NewServer(store)
	err = server.Start(config.ServerAddress)
	if err != nil {
		log.Fatal("cannot start the server:", err)
	}
}
