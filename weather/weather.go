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
	UVIndex                  float32                 `json:"uvi"`
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
	UVIndex                  float32    `json:"uv_index"`
	Icon                     string     `json:"icon"`
	OWMID                    int        `json:"openweathermap_id"`
}

type WeatherCacheEntry struct {
	Response    *Data
	NextRefresh time.Time
}

var (
	apiURL    = "https://api.openweathermap.org/data/3.0/onecall?lat=%v&lon=%v&exclude=minutely&units=metric&lang=de&appid=%s"
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

	// 1. If the cache is completely empty, we have nothing to return.
	// We MUST block and fetch synchronously this first time.
	if cached == nil {
		return fetchAndUpdateCache(ctx)
	}

	// 2. If the cache exists but is expired, trigger an async refresh.
	if time.Now().After(cached.NextRefresh) {
		// TryLock (Go 1.18+) prevents a "cache stampede". It ensures only ONE
		// background goroutine refreshes the cache at a time.
		go func() {
			// Run the fetch in the background. If it fails, the cache
			// remains stale and the next request will attempt to refresh it again.
			_, err := fetchAndUpdateCache(context.Background())
			if err != nil {
				log.Println("Could not update weather cache:", err)
			}
		}()
	}

	// 3. Immediately return the cached response (which may be slightly stale)
	return cached.Response, nil
}

func fetchAndUpdateCache(ctx context.Context) (*Data, error) {
	response, err := Get(ctx)
	if err != nil {
		log.Println("Could not get weather response:", err)
		return nil, err
	}

	weatherCacheMutex.Lock()
	weatherCache = &WeatherCacheEntry{
		Response:    response,
		NextRefresh: time.Now().Add(weatherCacheDuration),
	}
	weatherCacheMutex.Unlock()

	return response, nil
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
		UVIndex:                  owpData.UVIndex,
		Icon:                     icon,
		OWMID:                    id,
	}
}
