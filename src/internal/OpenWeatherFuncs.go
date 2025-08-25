package weatherservice

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
)

// Functions to interact with the OpenWeatherMap API

const (
	apiWeatherURL     = "https://pro.openweathermap.org/data/2.5/weather"
	apiCoordinatesURL = "http://api.openweathermap.org/geo/1.0/direct"
)

var (
	apiKey = os.Getenv("API_WEATHER_KEY	")

)

type weatherAPIResp struct {
	Dt   int64 `json:"dt"`
	Main struct {
		Temp      float64 `json:"temp"`
		FeelsLike float64 `json:"feels_like"`
		Pressure  int64   `json:"pressure"`
	} `json:"main"`
	Wind struct {
		Speed float64 `json:"speed"`
		Deg   int64   `json:"deg"`
	} `json:"wind"`
}

func GetCoordinates(cityName string) (CityType, error) {
	url := fmt.Sprintf("%s?q=%s&limit=1&appid=%s", apiCoordinatesURL, cityName, apiKey)
	resp, err := http.Get(url)
	if err != nil {
		return CityType{}, fmt.Errorf("GetCoordinates: request error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return CityType{}, errors.New("GetCoordinates: non-200 response from API")
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return CityType{}, fmt.Errorf("GetWeather: read body error: %w", err)
	}

	var city CityType
	if err := json.Unmarshal(data, &city); err != nil {
		return CityType{}, fmt.Errorf("GetCoordinates: decode error: %w", err)
	}

	return city, nil
}

func GetWeather(city CityType) (weatherAPIResp, error) {
	url := fmt.Sprintf("%s?lat=%f&lon=%f&appid=%s&units=metric", apiWeatherURL, city.Lat, city.Lon, apiKey)
	resp, err := http.Get(url)
	if err != nil {
		return weatherAPIResp{}, fmt.Errorf("GetWeather: request error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return weatherAPIResp{}, errors.New("GetWeather: non-200 response from API")
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return weatherAPIResp{}, fmt.Errorf("GetWeather: read body error: %w", err)
	}

	var weatherResp weatherAPIResp
	if err := json.Unmarshal(data, &weatherResp); err != nil {
		return weatherAPIResp{}, fmt.Errorf("GetWeather: decode error: %w", err)
	}

	return weatherResp, nil
}