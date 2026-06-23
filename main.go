package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

type StationCoords struct {
	Geometry struct {
		Coordinates []float64	`json:"coordinates"`
	}				`json:"geometry"`
	
}

type StationHourlyLink struct {
	Properties struct {
		ForecastHourly string	`json:"forecastHourly"`
	}				`json:"properties"`

}

type HourlyForecast struct {
	StartTime string		`json:"startTime"`
	EndTime string			`json:"endTime"`
	Temperature int			`json:"temperature"` 
}

type StationHourlyForecast struct {
	Properties struct {
		Periods []HourlyForecast`json:"periods"`
	}				`json:"properties"`
}

func main() {
	// load .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("No .env file found")
	}

	// get redis creds
	redisHost := os.Getenv("REDIS_HOST") 
	redisPort := os.Getenv("REDIS_PORT")
	redisUser := os.Getenv("REDIS_USER")
	redisPass := os.Getenv("REDIS_PASS")
	redisDB := os.Getenv("REDIS_DB")

	// redis connection
	connectionString := fmt.Sprintf("redis://%s:%s@%s:%s/%s", redisUser, redisPass, redisHost, redisPort, redisDB)
	opt, err := redis.ParseURL(connectionString)
	if err != nil {
		log.Fatal(err)
	}

	client := redis.NewClient(opt)
	client.Ping(context.Background())

	router := gin.Default()

	router.GET("/hourlytemp", getHourlyTemp)


	router.Run(":8080")

}

func getHourlyTemp(context *gin.Context) {

	station := context.Query("station")

	// Get Coords For Given Station
	url := fmt.Sprintf("https://api.weather.gov/stations/%s", station)
	response, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	bytes, err := io.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}

	var StationCoords StationCoords
	err = json.Unmarshal(bytes, &StationCoords)

	// Get Hourly Link For Given Coords
	url = fmt.Sprintf("https://api.weather.gov/points/%f,%f",StationCoords.Geometry.Coordinates[1], StationCoords.Geometry.Coordinates[0])
	response, err = http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	bytes, err = io.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}

	var HourlyLink StationHourlyLink
	err = json.Unmarshal(bytes, &HourlyLink)

	// Get Hourly Forecast 
	response, err = http.Get(HourlyLink.Properties.ForecastHourly)
	if err != nil {
		log.Fatal(err)
	}
	bytes, err = io.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}
	
	var StationHourlyForecast StationHourlyForecast
	err = json.Unmarshal(bytes, &StationHourlyForecast)

	context.IndentedJSON(200, StationHourlyForecast.Properties.Periods[:12])
	
}

