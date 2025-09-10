package weatherservice

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

const (
	apiWeatherURL     = "https://pro.openweathermap.org/data/2.5/weather"
	apiCoordinatesURL = "http://api.openweathermap.org/geo/1.0/direct"
)

var (
	apiKey = os.Getenv("API_WEATHER_KEY")
)

type weatherAPIResp struct {
	Dt   int64 `json:"dt"`
	Main struct {
		Temp      float32 `json:"temp"`
		FeelsLike float32 `json:"feels_like"`
		Pressure  int16   `json:"pressure"`
	} `json:"main"`
	Wind struct {
		Speed float32 `json:"speed"`
		Deg   int16   `json:"deg"`
	} `json:"wind"`
}

func GetCoordinates(cityName string) (CityType, error) {
	url := fmt.Sprintf("%s?q=%s&limit=1&appid=%s", apiCoordinatesURL, cityName, apiKey)
	log.Printf("GetCoordinates: URL=%s", url)

	resp, err := http.Get(url)
	if err != nil {
		log.Printf("GetCoordinates: request error: %v", err)
		return CityType{}, fmt.Errorf("GetCoordinates: request error: %w", err)
	}
	defer resp.Body.Close()

	log.Printf("GetCoordinates: status=%s", resp.Status)

	if resp.StatusCode != http.StatusOK {
		return CityType{}, errors.New("GetCoordinates: non-200 response from API")
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return CityType{}, fmt.Errorf("GetCoordinates: read body error: %w", err)
	}

	var cities []CityType
	if err := json.Unmarshal(data, &cities); err != nil {
		log.Printf("GetCoordinates: decode error: %v", err)
		return CityType{}, fmt.Errorf("GetCoordinates: decode error: %w", err)
	}
	if len(cities) == 0 {
		log.Printf("GetCoordinates: no results for city %s", cityName)
		return CityType{}, fmt.Errorf("GetCoordinates: no results for city %s", cityName)
	}

	if len(data) > 0 {
		sample := string(data)
		if len(sample) > 200 {
			sample = sample[:200] + "..."
		}
		log.Printf("GetCoordinates: response sample=%s", sample)
	}

	return cities[0], nil
}

func GetWeather(city CityType) (weatherAPIResp, error) {
	url := fmt.Sprintf("%s?lat=%f&lon=%f&appid=%s&units=metric", apiWeatherURL, city.Lat, city.Lon, apiKey)
	log.Printf("GetWeather: URL=%s", url)

	resp, err := http.Get(url)
	if err != nil {
		log.Printf("GetWeather: request error: %v", err)
		return weatherAPIResp{}, fmt.Errorf("GetWeather: request error: %w", err)
	}
	defer resp.Body.Close()

	log.Printf("GetWeather: status=%s", resp.Status)

	if resp.StatusCode != http.StatusOK {
		return weatherAPIResp{}, errors.New("GetWeather: non-200 response from API")
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return weatherAPIResp{}, fmt.Errorf("GetWeather: read body error: %w", err)
	}

	if len(data) > 0 {
		sample := string(data)
		if len(sample) > 200 {
			sample = sample[:200] + "..."
		}
		log.Printf("GetWeather: response sample=%s", sample)
	}

	var weatherResp weatherAPIResp
	if err := json.Unmarshal(data, &weatherResp); err != nil {
		log.Printf("GetWeather: decode error: %v", err)
		return weatherAPIResp{}, fmt.Errorf("GetWeather: decode error: %w", err)
	}

	return weatherResp, nil
}
