package main

import (
	"log"
	"net/http"
	"os"

	"github.com/bonnguyenitc/shopee-stracks/back-end-go/api"
	"github.com/bonnguyenitc/shopee-stracks/back-end-go/database"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/rs/cors"
	"github.com/sirupsen/logrus"
)

func main() {
	file, err := os.OpenFile("storage/info.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	logrus.SetOutput(file)
	logrus.SetFormatter(&logrus.JSONFormatter{})
	// load env
	err = godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	// init logging
	dbName := os.Getenv("DB_NAME")
	// setup database
	err = database.NewMongoDB(dbName)
	if err != nil {
		log.Fatal(err)
	}
	// setup index
	database.SetupIndexed()
	// init router
	router := mux.NewRouter()
	// setup api
	api.SetupRoutes(router)
	// setup CORS
	handler := cors.Default().Handler(router)
	// jobs.RunCronJobs()
	log.Fatal(http.ListenAndServe(":8000", handler))

}
