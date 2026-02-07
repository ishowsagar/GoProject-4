package app

import (
	"database/sql"
	"fem/internal/api"
	"fem/internal/store"
	"fem/migrations"
	"fmt"
	"log"
	"net/http"
	"os"
)

//  types declarement
type Application struct {
	Logger *log.Logger
	WorkoutHandler *api.WorkoutHandler
	DB *sql.DB
}

func NewApplication() (*Application,error) {

	// database connection
	pgDb,err := store.Open()
	if err != nil {
		return nil,err
	}

	// migration hookup with app
	err = store.Migratefs(pgDb,migrations.FS,".")
	if err != nil {
		panic(err)
	}


	logger := log.New(os.Stdout,"",log.Ldate | log.Ltime) 

	// ! Our store will go here

	// ! Our Handlers will go here
	workoutHandler := api.NewWorkoutHandler()

	app := &Application{  // taking instance of type struct
		Logger : logger,
		WorkoutHandler: workoutHandler,
		DB: pgDb,
	}
	
	return app,nil // as we had to return both things as we specified in the return type of the function

}

// application struct has this method on it
func (a *Application) HealthCheck(w http.ResponseWriter,req *http.Request) {
	fmt.Fprintf(w,"Status is available\n")
	//  w.Write([]byte("hello from Go developer!"))
}