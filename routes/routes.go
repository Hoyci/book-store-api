package routes

import "github.com/gorilla/mux"

func Routes(router *mux.Router) {
	RegisterHealthcheckRoutes(router)
}
