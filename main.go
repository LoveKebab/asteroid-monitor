package main

import (
	"encoding/json"
	"fmt"
	"image/png"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"text/template"
	"time"

	"github.com/wcharczuk/go-chart"
	"gopkg.in/yaml.v2"
)

const nasaURL = "https://api.nasa.gov/neo/rest/v1/feed?start_date=%s&api_key=%s"

// look into using viper

type conf struct {
	Apikey string `yaml:"apikey"`
	Port   string `yaml:"port"`
}

type nasaReturnData struct {
	Links struct {
		Next string `json:"next"`
		Prev string `json:"prev"`
		Self string `json:"self"`
	} `json:"links"`
	ElementCount     int                          `json:"element_count"`
	NearEarthObjects map[string][]NearEarthObject `json:"near_earth_objects"`
}

// NearEarthObject is a struct used to parse the map in nasaReturnData for NearEarthObjects
type NearEarthObject struct {
	Links struct {
		Self string `json:"self"`
	} `json:"links"`
	ID                 string  `json:"id"`
	NeoReferenceID     string  `json:"neo_reference_id"`
	Name               string  `json:"name"`
	NasaJplURL         string  `json:"nasa_jpl_url"`
	AbsoluteMagnitudeH float64 `json:"absolute_magnitude_h"`
	EstimatedDiameter  struct {
		Kilometers struct {
			EstimatedDiameterMin float64 `json:"estimated_diameter_min"`
			EstimatedDiameterMax float64 `json:"estimated_diameter_max"`
		} `json:"kilometers"`
		Meters struct {
			EstimatedDiameterMin float64 `json:"estimated_diameter_min"`
			EstimatedDiameterMax float64 `json:"estimated_diameter_max"`
		} `json:"meters"`
		Miles struct {
			EstimatedDiameterMin float64 `json:"estimated_diameter_min"`
			EstimatedDiameterMax float64 `json:"estimated_diameter_max"`
		} `json:"miles"`
		Feet struct {
			EstimatedDiameterMin float64 `json:"estimated_diameter_min"`
			EstimatedDiameterMax float64 `json:"estimated_diameter_max"`
		} `json:"feet"`
	} `json:"estimated_diameter"`
	IsPotentiallyHazardousAsteroid bool `json:"is_potentially_hazardous_asteroid"`
	CloseApproachData              []struct {
		CloseApproachDate      string `json:"close_approach_date"`
		EpochDateCloseApproach int64  `json:"epoch_date_close_approach"`
		RelativeVelocity       struct {
			KilometersPerSecond string `json:"kilometers_per_second"`
			KilometersPerHour   string `json:"kilometers_per_hour"`
			MilesPerHour        string `json:"miles_per_hour"`
		} `json:"relative_velocity"`
		MissDistance struct {
			Astronomical string `json:"astronomical"`
			Lunar        string `json:"lunar"`
			Kilometers   string `json:"kilometers"`
			Miles        string `json:"miles"`
		} `json:"miss_distance"`
		OrbitingBody string `json:"orbiting_body"`
	} `json:"close_approach_data"`
	IsSentryObject bool `json:"is_sentry_object"`
}

func (c *conf) getConf() *conf {

	yamlFile, err := ioutil.ReadFile("conf.yaml")
	if err != nil {
		log.Printf("yamlFile.Get err %v ", err)
	}
	err = yaml.Unmarshal(yamlFile, c)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}

	return c
}

func nasaNeoBrowse(c conf) (*nasaReturnData, error) {
	apiKey := c.Apikey
	startDate := time.Now().Format("2006-01-02")
	var apiURL = fmt.Sprintf(nasaURL, startDate, apiKey)

	request, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, err
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	var nasaData nasaReturnData
	err = json.Unmarshal(body, &nasaData)
	return &nasaData, err
}

func drawChartWide(objects []NearEarthObject) (w io.Writer, err error) {
	graph := chart.Chart{
		Width: 1920, //this overrides the default.
		Series: []chart.Series{
			chart.ContinuousSeries{
				XValues: []float64{1.0, 2.0, 3.0, 4.0},
				YValues: []float64{1.0, 2.0, 3.0, 4.0},
			},
		},
	}

	graph.Render(chart.PNG, w)

	return w, err
}

func main() {
	var c conf
	c.getConf()
	nasaData, err := nasaNeoBrowse(c)

	port := c.Port
	if err != nil {
		panic(err)
	}

	objects := []NearEarthObject{}
	for _, NearEarthData := range nasaData.NearEarthObjects {
		for _, DeepNearEarthData := range NearEarthData {
			if DeepNearEarthData.IsPotentiallyHazardousAsteroid {
				objects = append(objects, DeepNearEarthData)
				fmt.Println("Asteroid Name:", DeepNearEarthData.Name)
				fmt.Println("Asteroid Potentially Hazardous:", DeepNearEarthData.IsPotentiallyHazardousAsteroid)
				fmt.Println("Absolute Magnitude:", DeepNearEarthData.AbsoluteMagnitudeH)
				fmt.Println("Size in KM's:", DeepNearEarthData.EstimatedDiameter.Kilometers.EstimatedDiameterMax)
				for _, approach := range DeepNearEarthData.CloseApproachData {
					fmt.Println("Close Approach Date:", approach.CloseApproachDate)
					fmt.Println("Total Projected Distance Asteroid Will Miss Earth:", approach.MissDistance.Kilometers, "KM's")
				}
				fmt.Println("More information at: " + DeepNearEarthData.NasaJplURL + "\n")
			}
		}
	}

	// below is the start of saving a chart as a file
	out, err := os.Create("./output.png")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	derp, err := drawChartWide(objects)

	err = png.Encode(out, derp)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	renderer := template.Must(template.ParseFiles("index.html"))

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		renderer.Execute(w, objects)
	})

	srv := &http.Server{
		Addr: port,
		// Enforcement of timeouts
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	log.Println("Listening on port " + port + "...")
	log.Fatalf("Fatal server error: %v", srv.ListenAndServe())
}
