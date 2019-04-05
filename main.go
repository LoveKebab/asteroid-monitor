package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"gopkg.in/yaml.v2"
)

const port = ":10443" // add to config
const nasaURL = "https://api.nasa.gov/neo/rest/v1/neo/browse?api_key=%s"

// look into using viper

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
	NearEarthObjects []nearEarth `json:"near_earth_objects"`
}

type nearEarth struct {
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
	var apiKey = "DEMO_KEY" //c.Apikey
	var apiURL = fmt.Sprintf(nasaURL, apiKey)

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

func main() {
	var c conf
	c.getConf()
	nasaData, err := nasaNeoBrowse(c)
	if err != nil {
		panic(err)
	}

	objects := []nearEarth{}
	for i, NearEarthObject := range nasaData.NearEarthObjects {
		if NearEarthObject.IsPotentiallyHazardousAsteroid {
			objects = append(objects, NearEarthObject)
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
	templ := `
	<body>
		{{ range $index, $element := . }}
		<p>{{ $element.Name }}</p>
		<p>{{ $element.Designation }}</p>
		<p>{{ $element.IsPotentiallyHazardousAsteroid }}</p>
		<p>{{ $element.AbsoluteMagnitudeH }}</p>
		<br>
		{{ end }}
	</body>
	`
	renderer, _ := template.New("basic").Parse(templ)

	router := mux.NewRouter()
	router.PathPrefix("/").Handler(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		renderer.Execute(response, objects)
		return
	}))

	//router.HandleFunc("/api", HelloWorld).Methods("GET")
	// Serve files from this directory if no api routes are hit
	//router.PathPrefix("/").Handler(http.FileServer(http.Dir("www")))

	srv := &http.Server{
		Handler: router,
		Addr:    port,
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
