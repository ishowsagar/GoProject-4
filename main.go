package main

// imports
import (
	"fem/internal/app"
	"fem/internal/routes"
	"flag"
	"fmt"
	"net/http"
	"time"
)

// Main function where go application spins up
func main() {

	// fallback port if not specified
	var port int
	flag.IntVar(&port,"port",8080,"GO BACKEND SERVER!")
	flag.Parse() // execute it

	app,err := app.NewApplication() //! returns Logger's output

	//  if caught any error intiting app
	if err !=nil {
		panic(err) // exits the app 
	}

	// closing db connection
	defer app.DB.Close() //!defer the execution to the very end of the application

	// ? - otherwise successfully imported function and executed
	fmt.Println("app is running!")

	//! server management

	// ? - handles request on this path
	r := routes.SetupRoutes(app)	// needs to pass logger as it points to application struct
	// creating instance of a server
	server := &http.Server{
		Addr: fmt.Sprintf(":%d",port),
		IdleTimeout: time.Minute,
		Handler: r, //! now parent handler is set for all route req --> handled through chi routes
		ReadTimeout: 10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	app.Logger.Printf("App is running on port : %d\n",port)



	// * server listens for any incoming request
	err = server.ListenAndServe() // returns error if failed to listen for a sever
   // if caught error listening for a server
	if err !=nil {
		app.Logger.Fatal(err)
	} 

}

