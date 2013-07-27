package cfg

import (
	"database/sql"
	"flag"
	"github.com/golang/glog"
	_ "github.com/mattn/go-sqlite3"
)

var configSqlite = flag.String("config.sqlite", "config.db", "path to config database")

var db *sql.DB

func init() {
	var err error
	if configSqlite != nil && *configSqlite != "" {
		db, err = sql.Open("sqlite3", *configSqlite)
		if err != nil {
			glog.Fatalf("unable to open config database %q", *configSqlite)
		}
	}
}

func Int(key string) (int, error) {
	var val int
	row := db.QueryRow("SELECT val FROM C WHERE key = ?;", key)
	err := row.Scan(&val)
	return val, err
}

func String(key string) (string, error) {
	var val string
	row := db.QueryRow("SELECT val FROM C WHERE key = ?;", key)
	err := row.Scan(&val)
	return val, err
}

func MustInt(key string) int {
	val, err := Int(key)
	if err != nil {
		glog.Fatalf("Unable to retrieve value for key %q", key)
	}
	return val
}

func MustString(key string) string {
	val, err := String(key)
	if err != nil {
		glog.Fatalf("Unable to retrieve value for key %q", key)
	}
	return val
}
