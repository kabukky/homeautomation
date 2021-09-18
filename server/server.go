package server

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/kabukky/homeautomation/utils"
)

func Start() {
	router := httprouter.New()
	router.ServeFiles("/dashboard/*filepath", http.Dir(utils.DirectoryDashboard))
	http.ListenAndServe(utils.HostAndPort, router)
}
