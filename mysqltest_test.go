package mysqltest

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func TestNewMysqld(t *testing.T) {
	t.Run("in cointainer", func(t *testing.T) {
		if !inDockerContainer() {
			t.SkipNow()
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
		mysqld, err := NewMysqld(ctx, "mysql:latest")
		if err != nil {
			t.Fatal(err)
		}
		defer cancel()
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
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
		mysqld, err := NewMysqld(ctx, "mysql:latest")
		if err != nil {
			t.Fatal(err)
		}
		defer cancel()
		db, err := sql.Open("mysql", mysqld.DSN())
		if err != nil {
			t.Fatal(err)
		}
		if err := db.Ping(); err != nil {
			t.Error("ping failed")
		}
	})
}
