package main

import (
	"database/sql"
	"flag"
	"fmt"

	_ "github.com/lib/pq"

	"addreality"
)

var (
	dsn = flag.String("dsn", "postgres://addreality:arbuz@localhost:5432/addreality?sslmode=disable", "set data source name")
)

func main() {
	db, err := sql.Open("postgres", *dsn)
	if err != nil {
		panic(err)
	}

	if err := db.Ping(); err != nil {
		panic(err)
	}

	builder, err := addreality.NewInsertBuilder(addreality.PgSQLDriver)
	if err != nil {
		panic(err)
	}

	if err := BulkDevice(db, builder); err != nil {
		panic(err)
	}
}

func BulkDevice(db *sql.DB, b addreality.InsertBuilder) error {
	rows := []struct {
		Name       string
		GroupID    int
		PlatformID int
	}{
		{Name: "device1", GroupID: 1, PlatformID: 5281},
		{Name: "device2", GroupID: 2, PlatformID: 5282},
		{Name: "device3", GroupID: 3, PlatformID: 5283},
	}

	for _, r := range rows {
		if err := b.Append(r.Name, r.GroupID, r.PlatformID); err != nil {
			return err
		}
	}

	batches, err := b.ToSQL()
	if err != nil {
		return err
	}

	for _, b := range batches {
		var query = fmt.Sprintf("INSERT INTO devices (name, group_id, platform_id) VALUES %s;", b.Query)
		_, err := db.Exec(query, b.Args...)
		if err != nil {
			return err
		}
	}

	return nil
}
