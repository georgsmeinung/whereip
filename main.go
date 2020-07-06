package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"sort"
	"strconv"
	"time"

	// Package gorilla/mux implements a request router and dispatcher
	// for matching incoming requests to their respective handler.
	"github.com/gorilla/mux"
	// GJSON is a Go package that provides a fast and simple way to get
	// values from a json document. It has features such as one line
	// retrieval, dot notation paths, iteration, and parsing json lines.
	"github.com/tidwall/gjson"
)

// ::: Default values :::
// ISO code use for unkown IP couuntry
// Default value: "XX"
var unkwonCountryCode string

// Exchange rate if not available
// Default value: 0
var unvailableRate float64

// Base currency for exchange query
// Default value: "USD"
var baseCurrency string

// Base latitude for distance calc
// Default value: -34.603333 (Buenos Aires)
var baseLat float64

// Base longitud for distance calc
// Default value: -58.381667 (Buenos Aires)
var baseLng float64

// ::: External Services URLs :::
// URL for IP Geoference formatted for sprintf parsing
// Default value: "https://api.ip2country.info/ip?%v"
var ipInfoURL string

// URL for Country Info formatted for sprintf parsing
// Default value: "https://restcountries.eu/rest/v2/alpha/%v"
var countryInfoURL string

// API Key for Currency Rate API
// Default value: "fadee53564a2f8e3e95d7d8cc47d4c64"
var currencyAPIKey string

// URL for Currency Rate formatted for sprintf parsing
// Default value: "http://data.fixer.io/api/latest?access_key=%v&symbols=%v,%v"
var currencyInfoURL string

// ::: JSON Paths :::
// JSON Path for ISO countryCode formated for gjson parsing
// Default value: countryCode
var jsonCountryCodePath string

// JSON Path for Country Name formated for gjson parsing
// Default value: "translations.es"
var jsonCountryNamePath string

// JSON Path for Currency Rate formated for sprintg and gjson parsing
// Default value: "rates.%v"
var jsonCurrencyRatePath string

// JSON Path for Currency Code formated for gjson parsing
// Default value: "currencies.0.code"
var jsonCurrencyCodePath string

// JSON Path for Languages formated for gjson parsing
// Default value: "languages.#.name"
var jsonLanguagesPath string

// JSON Path for Timezones formated for gjson parsing
// Default value: "timezones"
var jsonTimezonesPath string

// JSON Path for Country Latitude formated for gjson parsing
// Default value: "latlng.0"
var jsonLatPath string

// JSON Path for Country Longitude formated for gjson parsing
// Default value: "latlng.1"
var jsonLngPath string

// This function Initializes global vars extracting their values
// from OS environment variables and conveting the strings to matching types
func initEnv() {
	var err error
	unkwonCountryCode = os.Getenv("WHIP_UNKCURR")
	unvailableRate, err = strconv.ParseFloat(os.Getenv("WHIP_UNAVCUR"), 64)
	if err != nil {
		panic(err)
	}
	baseCurrency = os.Getenv("WHIP_BASECUR")
	baseLat, err = strconv.ParseFloat(os.Getenv("WHIP_BASELAT"), 64)
	if err != nil {
		panic(err)
	}
	baseLng, err = strconv.ParseFloat(os.Getenv("WHIP_BASELNG"), 64)
	if err != nil {
		panic(err)
	}
	ipInfoURL = os.Getenv("WHIP_INFOURL")
	countryInfoURL = os.Getenv("WHIP_CINFURL")
	currencyAPIKey = os.Getenv("WHIP_CUEXKEY")
	currencyInfoURL = os.Getenv("WHIP_CUEXURL")
	jsonCountryCodePath = os.Getenv("WHIP_CCODPAT")
	jsonCountryNamePath = os.Getenv("WHIP_CNAMPAT")
	jsonCurrencyRatePath = os.Getenv("WHIP_CRATPAT")
	jsonCurrencyCodePath = os.Getenv("WHIP_CUCDPAT")
	jsonLanguagesPath = os.Getenv("WHIP_LANGPAT")
	jsonTimezonesPath = os.Getenv("WHIP_TZONPAT")
	jsonLatPath = os.Getenv("WHIP_BLATPAT")
	jsonLngPath = os.Getenv("WHIP_BLNGPAT")
}

// Query output service format
type whereIPquery struct {
	From        string
	When        string
	CountryCode string
	CountryName string
	Languages   []string
	Timezones   []string
	Distante    int64
	Currency    string
	ExRate      float64
}

// Query stat single entry
type queriesCount struct {
	CountryCode string
	CountryName string
	Distance    int64
	Queries     int64
}

// Query stat summary
type statsSummary struct {
	FurthestDistance int64
	ClosestDistance  int64
	AverageDistance  int64
}

// Error Message format
type errorMessage struct {
	Message string
}

// Global Array with queries stats
// TODO: put queries in an external persistent repository
var queriesStats []queriesCount

// getISO31661a2 With a given IPv4 or IPv6 checks the
// georeference service and returns the ISO 3166-1 alpha 2 code
// for the country from which the query was made
// TODO: implement a cache for the query
func getISO31661a2(ip string) string {
	URL := fmt.Sprintf(ipInfoURL, string(ip))

	resp, err := http.Get(URL)
	if err != nil {
		log.Fatalln(err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	} else {
		ISOCode := gjson.GetBytes(body, jsonCountryCodePath)
		return ISOCode.String()
	}

	return unkwonCountryCode
}

// getCountryInfo with a given ISO 3166-1 alpha 2 country code
// checks the Country info service and returns a byte array
// with the resulting JSON
// TODO: implement a cache for the query
func getCountryInfo(isoCode string) []byte {
	var countryMap map[string]interface{}
	URL := fmt.Sprintf(countryInfoURL, isoCode)

	resp, err := http.Get(URL)
	if err != nil {
		log.Fatalln(err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	} else {
		err := json.Unmarshal(body, &countryMap)
		if err != nil {
			panic(err)
		}
	}

	return body
}

// getCurrencyRate with a given ISO 4217 currency code
// checks the Currency Exchange service and returns
// the exchange rate respect to the base currency
// TODO: implement a cache for the query
func getCurrencyRate(isoCode string) float64 {
	URL := fmt.Sprintf(currencyInfoURL, currencyAPIKey, baseCurrency, isoCode)

	resp, err := http.Get(URL)
	if err != nil {
		log.Fatalln(err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	} else {
		ISOCode := gjson.GetBytes(body, fmt.Sprintf(jsonCurrencyRatePath, baseCurrency))
		return ISOCode.Float()
	}

	return unvailableRate
}

// getCountryCurrency with a given array of byte containing
// the country info returns the ISO 4217 currency code.
// If the country has more than one currency, returns only
// the first in the JSON
func getCountryCurrency(infoMap []byte) string {
	currencyISOCode := gjson.GetBytes(infoMap, jsonCurrencyCodePath)
	return currencyISOCode.String()
}

// getCountryName with a given array of byte containing the
// country info returns the Country Name
func getCountryName(infoMap []byte) string {
	countryName := gjson.GetBytes(infoMap, jsonCountryNamePath)
	return countryName.String()
}

// getCountryLanguages with a given array of byte containing
// the country info returns an array of string listing the name
// of the official languages spoken in the country
func getCountryLanguages(infoMap []byte) []string {
	var langList []string
	countryLanguages := gjson.GetBytes(infoMap, jsonLanguagesPath)
	for _, name := range countryLanguages.Array() {
		langList = append(langList, name.String())
	}
	return langList
}

// getCountryTimeZones with a given array of byte containing
// the country info returns an array of string listing the time
// in each of the country time zones
func getCountryTimeZones(baseTime time.Time, infoMap []byte) []string {
	var TZList []string
	var TZTime string
	var TZDiffSeg int
	countryTimeZones := gjson.GetBytes(infoMap, jsonTimezonesPath)
	for _, TZName := range countryTimeZones.Array() {
		TZOffSet := TZName.String()
		if TZOffSet == "UTC" {
			TZDiffSeg = 0
		} else {
			TZHourDiff, err := strconv.Atoi(TZOffSet[4:5])
			if err != nil {
				panic(err)
			}
			TZMinDiff, err := strconv.Atoi(TZOffSet[7:8])
			if err != nil {
				panic(err)
			}
			TZDiffSeg = TZHourDiff * 60 * 60
			TZDiffSeg += TZMinDiff * 60
			if TZOffSet[3:3] == "-" {
				TZDiffSeg *= -1
			}
		}
		TZLocation := time.FixedZone(TZName.String(), TZDiffSeg)
		TZTime = baseTime.In(TZLocation).Format("2006-01-02T15:04:05-0700")
		TZList = append(TZList, TZTime)
	}
	return TZList
}

// hsin Implements the haversin(θ) function
// More info: https://en.wikipedia.org/wiki/Haversine_formula
func hsin(theta float64) float64 {
	return math.Pow(math.Sin(theta/2), 2)
}

// Distance calculates the distance (in meters) between two points of
// a given longitude and latitude relatively accurately (using a spherical
// approximation of the Earth) through the Haversin Distance Formula for
// great arc distance on a sphere with accuracy for small distances
// point coordinates are supplied in degrees and converted into rad.
// in the func distance returned is METERS!!!!!!
// More info: http://en.wikipedia.org/wiki/Haversine_formula
// Code source: https://gist.github.com/cdipaolo/d3f8db3848278b49db68
func Distance(lat1, lon1, lat2, lon2 float64) float64 {
	// convert to radians
	// must cast radius as float to multiply later
	var la1, lo1, la2, lo2, r float64
	la1 = lat1 * math.Pi / 180
	lo1 = lon1 * math.Pi / 180
	la2 = lat2 * math.Pi / 180
	lo2 = lon2 * math.Pi / 180

	r = 6378100 // Earth radius in METERS

	// calculate
	h := hsin(la2-la1) + math.Cos(la1)*math.Cos(la2)*hsin(lo2-lo1)

	return 2 * r * math.Asin(math.Sqrt(h))
}

// getCountryDistance with a given array of byte containing
// the country info returns the distance to the base location
// rounted to int. Aproximated distance is in kilometers.
func getCountryDistance(infoMap []byte) int64 {
	countryLat := gjson.GetBytes(infoMap, jsonLatPath).Float()
	countryLong := gjson.GetBytes(infoMap, jsonLngPath).Float()
	distanceKm := Distance(countryLat, countryLong, baseLat, baseLng) / 1000
	return int64(math.Round(distanceKm))
}

// ByDistance implements sort.Interface based on the Distance field.
type ByDistance []queriesCount

func (a ByDistance) Len() int           { return len(a) }
func (a ByDistance) Less(i, j int) bool { return a[i].Distance < a[j].Distance }
func (a ByDistance) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

// Find returns the smallest index i at which x == a[i],
// or len(a) if there is no such index.
// If not found, returns -1
func Find(a []queriesCount, x queriesCount) int {
	for i, n := range a {
		if x.CountryCode == n.CountryCode {
			return i
		}
	}
	return -1
}

// insertStat creates a stat record in the stats persistence
// TODO: insert stats by using an async media like message queues
func insertStat(q queriesCount) {
	searchIndex := Find(queriesStats, q)
	if searchIndex >= 0 {
		queriesStats[searchIndex].Queries += q.Queries
	} else {
		queriesStats = append(queriesStats, q)
	}
}

// getIPInfo is the Event Handler for the service at /whereip
// Returns a JSON with the info of the query
func getIPInfo(w http.ResponseWriter, r *http.Request) {
	fromIP := mux.Vars(r)["ip"]
	queryTimestamp := time.Now()
	countryISO31661a2 := getISO31661a2(fromIP)
	countryInfo := getCountryInfo(countryISO31661a2)
	countryName := getCountryName(countryInfo)
	countryLanguages := getCountryLanguages(countryInfo)
	countryTimeZones := getCountryTimeZones(queryTimestamp, countryInfo)
	countryDistance := getCountryDistance(countryInfo)
	currencyISO := getCountryCurrency(countryInfo)
	currencyRate := getCurrencyRate(currencyISO)

	thisQuery := whereIPquery{
		From:        fromIP,
		When:        queryTimestamp.Format("2006-01-02T15:04:05-0700"),
		CountryCode: countryISO31661a2,
		CountryName: countryName,
		Languages:   countryLanguages,
		Timezones:   countryTimeZones,
		Distante:    countryDistance,
		Currency:    currencyISO,
		ExRate:      currencyRate,
	}

	thisQueryCount := queriesCount{
		CountryCode: countryISO31661a2,
		CountryName: countryName,
		Distance:    countryDistance,
		Queries:     1,
	}

	insertStat(thisQueryCount)

	js, err := json.Marshal(thisQuery)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

// getStats is the Event Handler for the service at /stats
// Returns a JSON with the required challenge stats
func getStats(w http.ResponseWriter, r *http.Request) {
	if len(queriesStats) > 0 {
		sort.Sort(ByDistance(queriesStats))
		summary := statsSummary{
			FurthestDistance: 0,
			ClosestDistance:  0,
			AverageDistance:  0,
		}

		if len(queriesStats) > 0 {
			summary.ClosestDistance = queriesStats[0].Distance
			summary.FurthestDistance = queriesStats[len(queriesStats)-1].Distance
			var queriesSum int64
			var distanceSum int64
			for _, v := range queriesStats { //range returns both the index and value
				distanceSum += (v.Queries * v.Distance)
				queriesSum += v.Queries
			}
			summary.AverageDistance = int64(math.Round(float64(distanceSum / queriesSum)))
		}

		js, err := json.Marshal(summary)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(js)
	} else {
		thisError := errorMessage{
			Message: "There isn't any query recorded.",
		}
		js, err := json.Marshal(thisError)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(js)
	}
}

// getFullStats is the Event Handler for the service at /fullstats
// Returns a JSON with the complete array of stats
func getFullStats(w http.ResponseWriter, r *http.Request) {
	if len(queriesStats) > 0 {
		js, err := json.Marshal(queriesStats)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(js)
	} else {
		thisError := errorMessage{
			Message: "There isn't any query recorded.",
		}
		js, err := json.Marshal(thisError)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(js)
	}
}

// getFullStats is the Event Handler for the service at /fullstats
// Returns a JSON with the complete array of stats
func clearStats(w http.ResponseWriter, r *http.Request) {
	queriesStats = nil
	thisError := errorMessage{
		Message: "Stats cleared!",
	}
	js, err := json.Marshal(thisError)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

// indexPath is Event Handler for the service at /
// renders the default page with Where IP usage instructions
func indexPath(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `
		<!doctype html>
		<html lang="en">
			<head>
				<!-- Required meta tags -->
				<meta charset="utf-8">
				<meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no">
			
				<!-- Bootstrap CSS -->
				<link rel="stylesheet" href="https://stackpath.bootstrapcdn.com/bootstrap/4.3.1/css/bootstrap.min.css" integrity="sha384-ggOyR0iXCbMQv3Xipma34MD+dH/1fQ784/j6cY/iJTQUOhcWr7x9JvoRxT2MZw1T" crossorigin="anonymous">		
				<title>Where IP</title>
			</head>
			<body>
				<div class="jumbotron jumbotron-fluid">
					<div class="container">
						<h1 class="display-4">Welcome to Where IP!</h1>
						<p class="lead">This an app that gets the Country for a given IP and related information. It supports both IPv4 and IPv6</p>
						<p>You can use it as a REST service in a line like:</p>
						<div class="alert alert-light" role="alert">
							<samp>
								AU <span class="badge badge-success">GET</span> <a href="/whereip/1.1.1.1">localhost:3000/whereip/1.1.1.1</a><br/>
								US <span class="badge badge-success">GET</span> <a href="/whereip/2606:4700:4700::1111">localhost:3000/whereip/2606:4700:4700::1111</a><br/>
								BR <span class="badge badge-success">GET</span> <a href="/whereip/200.223.129.162">localhost:3000/whereip/200.223.129.162</a><br/>
								ES <span class="badge badge-success">GET</span> <a href="/whereip/195.53.69.132">localhost:3000/whereip/195.53.69.132</a><br/>
								IN <span class="badge badge-success">GET</span> <a href="/whereip/203.115.71.66">localhost:3000/whereip/203.115.71.66</a>
							</samp>
						</div>
						<p>You can get view and clear stats from here:</p>
						<div class="alert alert-light" role="alert">
							<samp>
								<span class="badge badge-success">GET</span> <a href="/stats/">localhost:3000/stats/</a><br/>
								<span class="badge badge-success">GET</span> <a href="/fullstats/">localhost:3000/fullstats/</a><br/>
								<span class="badge badge-success">GET</span> <a href="/clearstats/">localhost:3000/clearstats/</a>
							</samp>
						</div>
					</div>
				</div>
			</body>
		</html>
	`)
}

func main() {
	initEnv()
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", indexPath)
	router.HandleFunc("/whereip/{ip}", getIPInfo).Methods("GET")
	router.HandleFunc("/stats/", getStats).Methods("GET")
	router.HandleFunc("/fullstats/", getFullStats).Methods("GET")
	router.HandleFunc("/clearstats/", clearStats).Methods("GET")
	log.Fatal(http.ListenAndServe(":3000", router))
}
