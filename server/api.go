package server

import (
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/kabukky/homeautomation/calendar"
	"github.com/kabukky/homeautomation/camera"
	"github.com/kabukky/homeautomation/weather"
)

func getWeather(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	weatherData, err := weather.GetCached(r.Context())
	if err != nil {
		respondWithError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	respondWithJSON(w, weatherData)
}

func getCalendar(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	events, err := calendar.GetEvents()
	if err != nil {
		respondWithError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	respondWithJSON(w, events)
}

func getCamera(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	times := camera.GetTimes()
	respondWithJSON(w, times)
}

func respondWithError(w http.ResponseWriter, errMsg string, code int) {
	w.WriteHeader(code)
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(map[string]string{"error": errMsg})
	if err != nil {
		http.Error(w, err.Error(), code)
		return
	}
}

func respondWithJSON(w http.ResponseWriter, resp interface{}) {
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(resp)
	if err != nil {
		respondWithError(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
