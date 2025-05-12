package sqlite

import (
	"database/sql"
	"log"
	"time"

	"github.com/bayu-gara/parking-lot/pkg/config"

	//external
	_ "github.com/mattn/go-sqlite3"
)

func Init(cfg config.SQLiteConfig) (db *sql.DB, err error) {
	db, err = sql.Open("sqlite3", cfg.DBName+".db")
	if err != nil {
		log.Fatal(err)
	}

	db.SetMaxIdleConns(cfg.MaxIdleConnection)
	db.SetMaxOpenConns(cfg.MaxOpenConnection)
	db.SetConnMaxIdleTime(time.Duration(cfg.ConnectionMaxIdleTimeMinutes) * time.Minute)
	db.SetConnMaxLifetime(time.Duration(cfg.ConnectionMaxLifeTimeMinutes) * time.Minute)

	return db, err
}
