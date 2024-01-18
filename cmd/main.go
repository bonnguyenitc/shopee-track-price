package main

import (
	"log"
	"net/http"
	"os"

	"github.com/bonnguyenitc/shopee-stracks/back-end-go/api"
	"github.com/bonnguyenitc/shopee-stracks/back-end-go/database"
	"github.com/bonnguyenitc/shopee-stracks/back-end-go/jobs"
	"github.com/bonnguyenitc/shopee-stracks/back-end-go/logs"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

func main() {
	logs.LogSetup()
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	// init logging
	dbName := os.Getenv("DB_NAME")
	// setup database
	err = database.NewMongoDB(dbName)
	if err != nil {
		logs.LogWarning(logrus.Fields{
			"data": err.Error(),
		}, "main run database.NewMongoDB")
		log.Fatal(err)
	}
	// setup index
	database.SetupIndexed()
	// init router
	router := mux.NewRouter()
	// setup api
	api.SetupRoutes(router)
	jobs.RunCronJobs()
	log.Fatal(http.ListenAndServe(":8000", router))
}
