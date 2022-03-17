package main

import (
	"database/sql"
	"log"

	"github.com/gyu-young-park/simplebank/api"
	db "github.com/gyu-young-park/simplebank/db/sqlc"
	"github.com/gyu-young-park/simplebank/util"
	_ "github.com/lib/pq"
)

func main() {
	log.Println("Application Start")
	config, err := util.LocalConfig(".")
	if err != nil {
		log.Fatal("cannot load config", err)
	}
	conn, err := sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}
	store := db.NewStore(conn)
	server, err := api.NewServer(config, store)
	if err != nil {
		log.Fatal("cannot create server:", err)
	}

	err = server.Start(config.ServerAddress)
	if err != nil {
		log.Fatal("cannot start server:", err)
	}
}
