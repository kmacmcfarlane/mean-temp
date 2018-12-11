package main

import (
"fmt"
"io/ioutil"
"net/http"
"time"

"github.com/Jeffail/gabs"
	"errors"
)

type Coordinate struct {
	Latitude  float64
	Longitude float64
}

const weatherCityUrlTemplate string = "https://www.metaweather.com/api//location/search/?lattlong=%f,%f"
const weatherUrlTemplate string = "https://www.metaweather.com/api/location/%d/%d/%d/%d"
const cityUrls string = "https://public.opendatasoft.com/api/records/1.0/search/?dataset=1000-largest-us-cities-by-population-with-geographic-coordinates&facet=city&facet=state&sort=population&rows=100"

func main() {

	cityData, err := doGetRequest(cityUrls)
	if err != nil {
		panic(err)
	}

	cityDataParsed, _ := gabs.ParseJSON(cityData)

	cities, _ := cityDataParsed.Path("records").Children()
	cityCoordinates := [100]Coordinate{}
	temperatures := make([]float64, 0)
	for i, city := range cities {
		coord := city.Path("fields.coordinates").Data().([]interface{})
		cityCoordinates[i] = Coordinate{
			Latitude:  coord[0].(float64),
			Longitude: coord[1].(float64),
		}

		temp, err := getCurrentTemperatureForCoordinates(cityCoordinates[i])

		if nil != err {
			fmt.Printf("Error getting temperature data for city: %s\n", err)
			continue
		}

		temperatures = append(temperatures, temp)
	}

	// Compute mean
	var total float64
	for _, temp := range temperatures {
		total += temp
	}

	mean := total / float64(len(temperatures))

	fmt.Printf("~~~~~~~~\nmean: %f\n", mean)
}

func getCurrentTemperatureForCoordinates(coord Coordinate) (result float64, err error) {
	weatherCityData, err := doGetRequest(fmt.Sprintf(weatherCityUrlTemplate, coord.Latitude, coord.Longitude))
	if err != nil {
		panic(err)
	}

	weatherCitiesParsed, _ := gabs.ParseJSON(weatherCityData)
	weatherCityWoeids := weatherCitiesParsed.Path("woeid").Data().([]interface{})
	weatherURLFormatted := fmt.Sprintf(weatherUrlTemplate, int64(weatherCityWoeids[0].(float64)), time.Now().Year(),
		int(time.Now().Month()), time.Now().Day())

	weatherData, err := doGetRequest(weatherURLFormatted)

	if err != nil {
		return result, err
	}

	weatherDataParsed, _ := gabs.ParseJSON(weatherData)

	tempNode := weatherDataParsed.Path("the_temp").Data().([]interface{})[0]

	if tempNode == nil {
		err = errors.New("no temp data found")
		return result, err
	}

	result = tempNode.(float64)

	fmt.Printf("temp: %f\n", result)

	return result, err
}

func doGetRequest(url string) ([]byte, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}