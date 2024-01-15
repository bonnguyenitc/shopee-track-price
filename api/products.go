package api

import (
	"net/http"

	"github.com/gorilla/mux"
)

func getProductionsHandler(w http.ResponseWriter, r *http.Request) {
	//
}

func SetupProductsApiRoutes(router *mux.Router) {
	router.HandleFunc("/api/products", getProductionsHandler).Methods("POST")
}
