// ClickHouseFuncs.go
package weatherservice

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"
	"sync"

	clickhouse "github.com/ClickHouse/clickhouse-go/v2"
)

var (
	ClickhouseConn clickhouse.Conn
	MapOfCities    map[string]CityType = make(map[string]CityType)
	mapMu          sync.RWMutex
)

type CityType struct {
	Name string  `json:"name"`
	Lat  float32 `json:"lat"`
	Lon  float32 `json:"lon"`
}

func InitClickhouse() error {
	host := os.Getenv("CLICKHOUSE_HOST")
	port := os.Getenv("CLICKHOUSE_PORT")
	user := os.Getenv("CLICKHOUSE_USER")
	password := os.Getenv("CLICKHOUSE_PASSWORD")
	database := os.Getenv("CLICKHOUSE_DB")

	log.Printf("InitClickhouse: env host=%s port=%s user=%s db=%s", host, port, user, database)

	if host == "" || port == "" || user == "" || password == "" || database == "" {
		return fmt.Errorf("ClickHouse environment variables are not set properly")
	}

	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{fmt.Sprintf("%s:%s", host, port)},
		Auth: clickhouse.Auth{
			Database: database,
			Username: user,
			Password: password,
		},
	})

	if err != nil {
		return fmt.Errorf("failed to connect to ClickHouse: %v", err)
	}

	ClickhouseConn = conn

	if err := createTables(); err != nil {
		return fmt.Errorf("failed to create tables: %v", err)
	}

	startPeriodicTask(30)
	log.Println("InitClickhouse: ready and periodic task started")

	return nil
}

func createTables() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	queries := []string{
		`CREATE TABLE IF NOT EXISTS weather_metrics ( 
                timestamp Datetime, 
                city String,
                temp Float32,
                app_temp Float32,
                pressure Int16,
                wind_speed Float32,
                wind_deg Int16
            ) ENGINE = MergeTree()
            PARTITION BY toYYYYMM(timestamp)
            ORDER BY timestamp`,

		`CREATE TABLE IF NOT EXISTS cities(
			city String,
			lat Float32,
			lon Float32
		) ENGINE = MergeTree()
		ORDER BY (city)`,
	}

	for _, query := range queries {
		if err := ClickhouseConn.Exec(ctx, query); err != nil {
			return fmt.Errorf("failed to create table: %v", err)
		}
	}

	rows, err := ClickhouseConn.Query(ctx, "SELECT city, lat, lon FROM cities")
	if err != nil {
		return fmt.Errorf("initClickHouse: select cities: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var city string
		var lat, lon float32

		if err := rows.Scan(&city, &lat, &lon); err != nil {
			log.Printf("initClickHouse: scan error: %v", err)
			continue
		}

		MapOfCities[city] = CityType{
			Name: city,
			Lat:  lat,
			Lon:  lon,
		}
	}

	log.Printf("initClickHouse: loaded %d cities from DB", len(MapOfCities))

	return nil
}

func addCitiesToDB(cities []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	batch, err := ClickhouseConn.PrepareBatch(ctx, "INSERT INTO cities (city, lat, lon)")
	if err != nil {
		return fmt.Errorf("addCitiesToDB: prepare batch: %w", err)
	}

	tmpMapOfCities := make(map[string]CityType)

	for _, cityName := range cities {
		_, ok := MapOfCities[cityName]
		if ok {
			continue
		}

		city, err := GetCoordinates(cityName)
		if err != nil {
			return fmt.Errorf("addCitiesToDB: get coordinates for city %s: %w", cityName, err)
		}
		tmpMapOfCities[cityName] = city

		if err := batch.Append(city.Name, city.Lat, city.Lon); err != nil {
			return fmt.Errorf("addCitiesToDB: append to batch: %w", err)
		}
	}

	if err := batch.Send(); err != nil {
		return fmt.Errorf("addCitiesToDB: send batch: %w", err)
	}

	for k, v := range tmpMapOfCities {
		MapOfCities[k] = v
	}
	log.Printf("addCitiesToDB: added %d cities to DB and map", len(tmpMapOfCities))

	return nil
}

func insertWeatherData(cities map[string]CityType) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	batch, err := ClickhouseConn.PrepareBatch(ctx, "INSERT INTO weather_metrics (timestamp, city, temp, app_temp, pressure, wind_speed, wind_deg)")
	if err != nil {
		return fmt.Errorf("insertWeatherResponses: prepare batch: %w", err)
	}

	for cityName, city := range cities {
		weatherResp, err := GetWeather(city)
		if err != nil {
			return fmt.Errorf("insertWeatherResponses: get weather for city %s: %w", cityName, err)
		}

		t := time.Unix(weatherResp.Dt, 0)

		if err := batch.Append(
			t,
			cityName,
			weatherResp.Main.Temp,
			weatherResp.Main.FeelsLike,
			weatherResp.Main.Pressure,
			weatherResp.Wind.Speed,
			weatherResp.Wind.Deg,
		); err != nil {
			return fmt.Errorf("insertWeatherResponses: append to batch: %w", err)
		}
	}

	if err := batch.Send(); err != nil {
		return fmt.Errorf("insertWeatherResponses: send batch: %w", err)
	}

	return nil
}

func startPeriodicTask(intervalSeconds int) {
	log.Println("start_periodic_task")

	go func() {
		ticker := time.NewTicker(time.Duration(intervalSeconds) * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			if err := insertWeatherData(MapOfCities); err != nil {
				log.Printf("Periodic task error: %v", err)
			} else {
				log.Println("Periodic task: Weather data inserted successfully")
			}
		}
	}()
}
