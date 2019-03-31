package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"gopkg.in/yaml.v2"

	"github.com/gorilla/mux"
)

const port = "10443"

type conf struct {
	Apikey string `yaml:"apikey"`
}

type nasaReturnData struct {
	Links struct {
		Next string `json:"next"`
		Self string `json:"self"`
	} `json:"links"`
	Page struct {
		Size          int `json:"size"`
		TotalElements int `json:"total_elements"`
		TotalPages    int `json:"total_pages"`
		Number        int `json:"number"`
	} `json:"page"`
	NearEarthObjects []struct {
		Links struct {
			Self string `json:"self"`
		} `json:"links"`
		NeoReferenceID     string  `json:"neo_reference_id"`
		Name               string  `json:"name"`
		Designation        string  `json:"designation"`
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
		OrbitalData struct {
			OrbitID                   string `json:"orbit_id"`
			OrbitDeterminationDate    string `json:"orbit_determination_date"`
			FirstObservationDate      string `json:"first_observation_date"`
			LastObservationDate       string `json:"last_observation_date"`
			DataArcInDays             int    `json:"data_arc_in_days"`
			ObservationsUsed          int    `json:"observations_used"`
			OrbitUncertainty          string `json:"orbit_uncertainty"`
			MinimumOrbitIntersection  string `json:"minimum_orbit_intersection"`
			JupiterTisserandInvariant string `json:"jupiter_tisserand_invariant"`
			EpochOsculation           string `json:"epoch_osculation"`
			Eccentricity              string `json:"eccentricity"`
			SemiMajorAxis             string `json:"semi_major_axis"`
			Inclination               string `json:"inclination"`
			AscendingNodeLongitude    string `json:"ascending_node_longitude"`
			OrbitalPeriod             string `json:"orbital_period"`
			PerihelionDistance        string `json:"perihelion_distance"`
			PerihelionArgument        string `json:"perihelion_argument"`
			AphelionDistance          string `json:"aphelion_distance"`
			PerihelionTime            string `json:"perihelion_time"`
			MeanAnomaly               string `json:"mean_anomaly"`
			MeanMotion                string `json:"mean_motion"`
			Equinox                   string `json:"equinox"`
			OrbitClass                struct {
				OrbitClassType        string `json:"orbit_class_type"`
				OrbitClassDescription string `json:"orbit_class_description"`
				OrbitClassRange       string `json:"orbit_class_range"`
			} `json:"orbit_class"`
		} `json:"orbital_data"`
		IsSentryObject bool   `json:"is_sentry_object"`
		NameLimited    string `json:"name_limited,omitempty"`
		SentryData     string `json:"sentry_data,omitempty"`
	} `json:"near_earth_objects"`
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

func nasaNeoBrowse() {

	var c conf
	c.getConf()

	var apiKey = c.Apikey

	const port = "10443"
	const nasaURL string = "https://api.nasa.gov/neo/rest/v1/neo/browse?api_key="

	var apiURL = fmt.Sprintf(nasaURL + apiKey)

	req, err := http.NewRequest("GET", apiURL, nil)
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		log.Println("Error on response. \n[ERROR] -", err)
		panic(err)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Println("Error on ioutil. \n[ERROR] -", err)
		panic(err)
	}

	var nasaData = new(nasaReturnData)
	err = json.Unmarshal(body, &nasaData)
	if err != nil {
		log.Println("Error Unmarshaling JSON Data. \n[ERROR] -", err)
		panic(err)
	}

	for i, NearEarthObject := range nasaData.NearEarthObjects {
		if NearEarthObject.IsPotentiallyHazardousAsteroid {
			fmt.Println("Asteroid Name:", NearEarthObject.Name)
			fmt.Println("Designation:", NearEarthObject.Designation)
			fmt.Println("Asteroid Potentially Hazardous:", NearEarthObject.IsPotentiallyHazardousAsteroid)
			fmt.Println("Absolute Magnitude:", NearEarthObject.AbsoluteMagnitudeH)
			fmt.Println("Size in KM's:", NearEarthObject.EstimatedDiameter.Kilometers.EstimatedDiameterMax)
			for _, CloseApproachData := range nasaData.NearEarthObjects[i].CloseApproachData {
				fmt.Println("Asteroid Speed in Kilometers Per Hour:", CloseApproachData.RelativeVelocity.KilometersPerHour)
			}
			fmt.Println("More information at: " + NearEarthObject.NasaJplURL + "\n")
		}
	}

}

func main() {
	nasaNeoBrowse()

	router := mux.NewRouter()

	// Routes for api calls.
	router.HandleFunc("/api", HelloWorld).Methods("GET")

	// Serve files from this directory if no api routes are hit
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("www")))

	srv := &http.Server{
		Handler: router,
		Addr:    ":" + port,
		// Enforcement of timeouts
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	log.Println("Listening on port " + port + "...")
	log.Fatalf("Fatal server error: %v", srv.ListenAndServe())
}

// HelloWorld API Reponse
func HelloWorld(response http.ResponseWriter, request *http.Request) {
	response.Write([]byte("Hello World!"))
	return
}
