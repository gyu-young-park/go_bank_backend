package db

import (
	"database/sql"
	"log"
	"os"
	"testing"

	"github.com/gyu-young-park/simplebank/util"
	_ "github.com/lib/pq"
)

var testQueries *Queries
var testDB *sql.DB

func TestMain(m *testing.M) {
	// setup
	config, err := util.LocalConfig("../..")
	if err != nil {
		log.Fatal("cannot load config:", err)
	}
	testDB, err = sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}
	testQueries = New(testDB)
	exitCode := m.Run()
	// teardown
	os.Exit(exitCode)
}
