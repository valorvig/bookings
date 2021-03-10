package main

import (
	"encoding/gob"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/valorvig/bookings/internal/config"
	"github.com/valorvig/bookings/internal/driver"
	"github.com/valorvig/bookings/internal/handlers"
	"github.com/valorvig/bookings/internal/helpers"
	"github.com/valorvig/bookings/internal/models"
	"github.com/valorvig/bookings/internal/render"
)

const portNumber = ":8080"

var app config.AppConfig        // put it here to let "middleware" which is in the same package "main" can also access to this "app"
var session *scs.SessionManager // customized session
var infoLog *log.Logger
var errorLog *log.Logger

func main() {
	db, err := run()
	if err != nil {
		log.Fatal(err) // stop the application from running
	}
	defer db.SQL.Close()

	fmt.Println(fmt.Sprintf("starting application on port %s", portNumber))
	// _ = http.ListenAndServe(portNumber, nil)

	srv := &http.Server{
		Addr:    portNumber,
		Handler: routes(&app), // call handlers in routes
	}

	err = srv.ListenAndServe()
	log.Fatal(err)
}

/*
$go run cmd/web/*.go
*/

func run() (*driver.DB, error) { // add *driver.DB and return db, so that we can use db.SQL.defer() later in main()
	// cut and paste everything in the main() starting and above "render.NewTemplates(&app)" till the beginning

	// what am I going to put in the session - the code with "gob" looks a bit odd
	// We can store primitives, but we need to actually tell our application about more complex types, structures
	// that we've defined ourselves and we want to store in the reservation model.
	// we've stored the value in the session. // m.App.Session.Put(r.Context(), "reservation", reservation)
	gob.Register(models.Reservation{}) // we want to store in the reservation model - now we can store values in the session
	gob.Register(models.User{})
	gob.Register(models.Room{})
	gob.Register(models.Restriction{})

	// change this to true when in production
	app.InProduction = false // change one place but affect everywhere

	infoLog = log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime) // Ldate & Ltime are the local date and local time
	app.InfoLog = infoLog

	errorLog = log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
	app.ErrorLog = errorLog

	// session := scs.New() // using short declaration means that the above session and this session are two different variables
	session = scs.New()
	session.Lifetime = 24 * time.Hour
	// true - let cppkies persist after the browser is closed
	// false - the session will not persist the next time users open a window or fire up their web browser session cookie
	session.Cookie.Persist = true
	// what site this cookie applies to
	session.Cookie.SameSite = http.SameSiteLaxMode // use default - cookies were sent for all requests
	// cookie is encrypted and from only HTTPS
	session.Cookie.Secure = app.InProduction // false for development, truen for productiuon

	app.Session = session

	// connect to database
	log.Println("Connecting to database...")
	// we're going to use hard coding for the meantime
	db, err := driver.ConnectSQL("host=localhost port=5432 dbname=bookings user=postgres password=1234")
	if err != nil {
		log.Fatal("Cannot connect to database! Dying...")
	}
	// defer db.SQL.Close() // we can't use this since the code is now in run() not main(). We still don't want to close it even after running run()

	tc, err := render.CreateTemplateCache()
	if err != nil {
		log.Fatal("cannot create template cache")
		return nil, err
	}

	app.TemplateCache = tc
	app.UseCache = false

	// transfer main app and repo to other files
	// We can create a new instance, a new MySQL function. This is a reposaitory for a handler - holds a repository that's of type for Postgres, MySQL, Oracle, MongoDB, etc.
	repo := handlers.NewRepo(&app, db) // db is a pointer to a driver - that driver can now only handle postgres
	handlers.NewHandlers(repo)
	render.NewRenderer(&app)
	helpers.NewHelpers(&app)

	// http.HandleFunc("/", handlers.Repo.Home)
	// http.HandleFunc("/about", handlers.Repo.About)

	return db, nil
}
