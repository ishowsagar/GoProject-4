package store

import (
	"database/sql"
	"fmt"
	"io/fs"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/pressly/goose/v3"
) 

func Open() (*sql.DB, error) {
	db,err := sql.Open("pgx","host=localhost user=postgres password=postgres dbname=postgres port=5432 sslmode=disable")

	//  if caught any error while opening an connection
	if err != nil {
		return nil,fmt.Errorf("db : open %w", err)
	}
	fmt.Println("Connected to the Database...")
	return db,err

}

// migrations --> verison control for sql database changes like thingy

func Migratefs(db *sql.DB,migrationfs fs.FS,dir string) error {
	goose.SetBaseFS(migrationfs)

	defer func() {
		goose.SetBaseFS(nil)
	}()
	
	return Migrate(db,dir)

}
func Migrate(db *sql.DB,dir string)error{
	err := goose.SetDialect("postgres")

	// 
	if err != nil {
		return fmt.Errorf("migrate : %w ",err)
	}
	err = goose.Up(db,dir)
	if err != nil {
		return fmt.Errorf("goose up : %w ",err)
	}
	return nil

}