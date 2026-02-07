package store

import (
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v4/stdlib"
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