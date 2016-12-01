package mysqltest

import (
	"database/sql"
	"testing"

	_ "github.com/go-sql-driver/mysql"
)

func TestNewMysqld(t *testing.T) {
	t.Run("in cointainer", func(t *testing.T) {
		if !inDockerContainer() {
			t.SkipNow()
		}
		mysqld, err := NewMysqld(nil)
		if err != nil {
			t.Fatal(err)
		}
		db, err := sql.Open("mysql", mysqld.DSN())
		if err != nil {
			t.Fatal(err)
		}
		if err := db.Ping(); err != nil {
			t.Error("ping failed")
		}
	})
	t.Run("outside cointainer", func(t *testing.T) {
		if inDockerContainer() {
			t.SkipNow()
		}
		mysqld, err := NewMysqld(nil)
		if err != nil {
			t.Fatal(err)
		}
		db, err := sql.Open("mysql", mysqld.DSN())
		if err != nil {
			t.Fatal(err)
		}
		if err := db.Ping(); err != nil {
			t.Error("ping failed")
		}
	})
}
