package main

import (
	"log"
	"net/http"
	"os"

	"github.com/bonnguyenitc/shopee-stracks/back-end-go/api"
	"github.com/bonnguyenitc/shopee-stracks/back-end-go/database"
	"github.com/bonnguyenitc/shopee-stracks/back-end-go/jobs"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	dbName := os.Getenv("DB_NAME")
	// setup database
	err = database.NewMongoDB(dbName)
	if err != nil {
		log.Fatal(err)
	}
	database.SetupIndexed()
	// init router
	router := mux.NewRouter()
	// setup api
	api.SetupRoutes(router)
	jobs.RunCronJobs()
	log.Fatal(http.ListenAndServe(":8000", router))
}
