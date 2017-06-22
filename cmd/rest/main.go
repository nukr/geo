package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	database "github.com/nukr/experiment/sqlite/pkg/database/sqlite3"
	"github.com/nukr/experiment/sqlite/pkg/geo"
)

func main() {
	r := mux.NewRouter()
	geoRepo := database.NewGeoRepository()
	r.Methods("OPTIONS").HandlerFunc(corsPreflight())
	r.HandleFunc("/", healthCheck())
	r.HandleFunc("/list", getCountry(geoRepo))
	r.HandleFunc("/list/{country}", getCounty(geoRepo))
	r.HandleFunc("/list/{country}/{county}", getDistrict(geoRepo))
	r.HandleFunc("/list/{country}/{county}/{district}", getStreet(geoRepo))
	r.HandleFunc("/healthz", healthCheck())
	http.ListenAndServe(":8888", r)
}

func corsPreflight() http.HandlerFunc {
	return func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		h := w.Header()
		h.Add("Access-Control-Allow-Origin", "*")
		h.Add("Access-Control-Allow-Methods", "GET, POST, PATCH, PUT, DELETE")
		h.Add("Access-Control-Allow-Headers", "Accept-Language, Content-Type")
		h.Add("Access-Control-Max-Age", "1728000")
		w.WriteHeader(200)
	}
}

func healthCheck() http.HandlerFunc {
	return func(
		w http.ResponseWriter,
		req *http.Request,
	) {
		h := w.Header()
		h.Add("Cache-Control", "no-cache, no-store, must-revalidate")
		h.Add("Pragma", "no-cache")
		h.Add("Expires", "0")
		fmt.Fprintf(w, "OK")
	}
}

func getCountry(geoRepo geo.Repository) http.HandlerFunc {
	return func(
		w http.ResponseWriter,
		req *http.Request,
	) {
		h := w.Header()
		h.Add("Content-Type", "application/json; charset=utf-8")
		h.Add("Access-Control-Allow-Origin", "*")
		acceptLang := strings.Split(strings.ToLower(req.Header.Get("Accept-Language")), ",")[0]
		if acceptLang != "zh-tw" {
			acceptLang = "en-us"
		}
		geos := geoRepo.GetCountry(acceptLang)
		var zz []string
		for _, geo := range geos {
			zz = append(zz, geo.Country)
		}
		data, _ := json.Marshal(zz)
		w.Write(data)
	}
}

func getCounty(geoRepo geo.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Add("Content-Type", "application/json; charset=utf-8")
		h.Add("Access-Control-Allow-Origin", "*")
		geos := geoRepo.GetCounty("台灣")
		var zz []string
		for _, geo := range geos {
			zz = append(zz, geo.County)
		}
		data, _ := json.Marshal(zz)
		w.Write(data)
	}
}

func getDistrict(geoRepo geo.Repository) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Add("Content-Type", "application/json; charset=utf-8")
		h.Add("Access-Control-Allow-Origin", "*")
		country := mux.Vars(r)["country"]
		county := mux.Vars(r)["county"]
		geos := geoRepo.GetDistrict(country, county)
		var zz []string
		for _, geo := range geos {
			zz = append(zz, geo.County)
		}
		data, _ := json.Marshal(zz)
		w.Write(data)
	}
}

func getStreet(geoRepo geo.Repository) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Add("Content-Type", "application/json; charset=utf-8")
		h.Add("Access-Control-Allow-Origin", "*")
		country := mux.Vars(r)["country"]
		county := mux.Vars(r)["county"]
		district := mux.Vars(r)["district"]
		geos := geoRepo.GetStreet(country, county, district)
		type Street struct {
			Name       string   `json:"name"`
			Zip        int      `json:"zip"`
			StreetName []string `json:"street_name"`
		}
		var zz []string
		for _, geo := range geos {
			zz = append(zz, geo.Street)
		}
		zip, _ := strconv.Atoi(geos[0].Zip[0:3])
		street := Street{
			Name:       geos[0].District,
			Zip:        zip,
			StreetName: zz,
		}
		data, _ := json.Marshal(street)
		w.Write(data)
	}
}
