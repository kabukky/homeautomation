package weather

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/kabukky/homeautomation/utils"
)

type OpenWeatherMapResponse struct {
	Current OpenWeatherMapData   `json:"current"`
	Hourly  []OpenWeatherMapData `json:"hourly"`
}

type OpenWeatherMapData struct {
	Time                     int64                   `json:"dt"`
	Sunset                   int64                   `json:"sunset"`
	Sunrise                  int64                   `json:"sunrise"`
	TemperatureCelsius       float32                 `json:"temp"`
	PrecipitationProbability float32                 `json:"pop"`
	Rain                     OpenWeatherMapRain      `json:"rain"`
	Weather                  []OpenWeatherMapWeather `json:"weather"`
}

type OpenWeatherMapWeather struct {
	ID          int    `json:"id"`
	Main        string `json:"main"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
}

type OpenWeatherMapRain struct {
	OneHour float32 `json:"1h"`
}

type Data struct {
	Current  *Weather  `json:"current"`
	Forecast []Weather `json:"forecast"`
}

type Weather struct {
	Time                     time.Time  `json:"time"`
	Sunset                   *time.Time `json:"sunset,omitempty"`
	Sunrise                  *time.Time `json:"sunrise,omitempty"`
	TemperatureCelsius       float32    `json:"temperature_celsius"`
	PrecipitationProbability float32    `json:"precipitation_probability"`
	PrecipitationAmount      float32    `json:"precipitation_amount"`
	Icon                     string     `json:"icon"`
	OWMID                    int        `json:"openweathermap_id"`
}

type WeatherCacheEntry struct {
	Response    *Data
	NextRefresh time.Time
}

var (
	apiURL    = "https://api.openweathermap.org/data/2.5/onecall?lat=%v&lon=%v&exclude=minutely&units=metric&lang=de&appid=%s"
	httClient = http.Client{
		Timeout: 30 * time.Second,
	}
	// Cache
	weatherCache         *WeatherCacheEntry
	weatherCacheMutex    sync.RWMutex
	weatherCacheDuration = 10 * time.Minute
)

func GetCached(ctx context.Context) (*Data, error) {
	weatherCacheMutex.RLock()
	cached := weatherCache
	weatherCacheMutex.RUnlock()
	if cached == nil || time.Now().After(cached.NextRefresh) {
		response, err := Get(ctx)
		if err != nil {
			if cached != nil {
				log.Println("Could not get weather response. Using cache:", err)
				return cached.Response, nil
			}
			return nil, err
		}
		weatherCacheMutex.Lock()
		// Refresh every 10 minutes
		weatherCache = &WeatherCacheEntry{Response: response, NextRefresh: time.Now().Add(weatherCacheDuration)}
		cached = weatherCache
		weatherCacheMutex.Unlock()
	}
	return cached.Response, nil
}

func Get(ctx context.Context) (*Data, error) {
	resp, err := httClient.Get(fmt.Sprintf(apiURL, utils.WeatherLatitude, utils.WeatherLongitude, utils.OpenWeatherMapAPIKey))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var owpResp OpenWeatherMapResponse
	err = json.NewDecoder(resp.Body).Decode(&owpResp)
	if err != nil {
		return nil, err
	}
	var result Data
	result.Current = convertOpenWeatherMapData(&owpResp.Current)
	for _, hour := range owpResp.Hourly {
		result.Forecast = append(result.Forecast, *convertOpenWeatherMapData(&hour))
	}

	return &result, nil
}

func convertOpenWeatherMapData(owpData *OpenWeatherMapData) *Weather {
	icon := ""
	id := 0
	if len(owpData.Weather) > 0 {
		icon = owpData.Weather[0].Icon
		id = owpData.Weather[0].ID
	}
	var sunset *time.Time
	var sunrise *time.Time
	if owpData.Sunset != 0 {
		s := time.Unix(owpData.Sunset, 0)
		sunset = &s
	}
	if owpData.Sunrise != 0 {
		s := time.Unix(owpData.Sunrise, 0)
		sunrise = &s
	}
	return &Weather{
		Time:                     time.Unix(owpData.Time, 0),
		Sunset:                   sunset,
		Sunrise:                  sunrise,
		TemperatureCelsius:       owpData.TemperatureCelsius,
		PrecipitationProbability: owpData.PrecipitationProbability,
		PrecipitationAmount:      owpData.Rain.OneHour,
		Icon:                     icon,
		OWMID:                    id,
	}
}
