package server

import (
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/kabukky/homeautomation/utils"
)

func Start() {
	router := httprouter.New()
	// Dashboard
	router.ServeFiles("/dashboard/*filepath", http.Dir(utils.DirectoryDashboard))
	// Weather
	router.GET(utils.APIBasePath+"weather", getWeather)
	router.GET(utils.APIBasePath+"calendar", getCalendar)
	router.GET(utils.APIBasePath+"camera", getCamera)
	log.Println("Starting HTTP server on port", utils.HostAndPort)
	http.ListenAndServe(utils.HostAndPort, router)
}
