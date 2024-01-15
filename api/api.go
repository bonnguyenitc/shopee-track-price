package api

import (
	"github.com/gorilla/mux"
)

func SetupRoutes(router *mux.Router) {
	SetupProductsApiRoutes(router)
	SetupTrackingsApiRoutes(router)
	SetupUsersApiRoutes(router)
}
