package app

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

//  types declarement
type Application struct {
	Logger *log.Logger
}

func NewApplication() (*Application,error) {

	logger := log.New(os.Stdout,"",log.Ldate | log.Ltime) 

	app := &Application{  // taking instance of type struct
		Logger : logger,
	}
	
	return app,nil // as we had to return both things as we specified in the return type of the function

}

// application struct has this method on it
func (a *Application) HealthCheck(w http.ResponseWriter,req *http.Request) {
	fmt.Fprintf(w,"Status is available\n")
	//  w.Write([]byte("hello from Go developer!"))
}