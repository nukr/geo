package main

import (
	"crypto/tls"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/nukr/geo/pkg/geo"
	"github.com/nukr/geo/pkg/repository"
	"golang.org/x/crypto/acme/autocert"
)

func main() {
	dsn := os.Getenv("POSTGRES_DSN")
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal(err)
	}
	r := mux.NewRouter()
	geoRepo, err := repository.NewGeoRepository(db)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	r.Methods("OPTIONS").HandlerFunc(corsPreflight())
	r.HandleFunc("/", healthCheck())
	r.HandleFunc("/list", getCountry(geoRepo))
	r.HandleFunc("/list/{country}", getCounty(geoRepo))
	r.HandleFunc("/list/{country}/{county}", getDistrict(geoRepo))
	r.HandleFunc("/list/{country}/{county}/{district}", getStreet(geoRepo))
	r.HandleFunc("/healthz", healthCheck())
	domain := os.Getenv("DOMAIN")
	if domain != "" {
		fmt.Printf("listening on port 443 with autocert %s\n", domain)
		m := &autocert.Manager{
			Cache:      autocert.DirCache("secret-dir"),
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist(domain),
		}
		go http.ListenAndServe(":http", m.HTTPHandler(nil))
		s := &http.Server{
			Addr:      ":https",
			TLSConfig: &tls.Config{GetCertificate: m.GetCertificate},
			Handler:   r,
		}
		s.ListenAndServeTLS("", "")
	} else {
		fmt.Println("listening on port 8888")
		http.ListenAndServe(":8888", r)
	}
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
		countryList := geoRepo.GetCountryList(acceptLang)
		data, _ := json.Marshal(countryList)
		w.Write(data)
	}
}

func getCounty(geoRepo geo.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Add("Content-Type", "application/json; charset=utf-8")
		h.Add("Access-Control-Allow-Origin", "*")
		cityList := geoRepo.GetCityList("zh-tw")
		data, _ := json.Marshal(cityList)
		w.Write(data)
	}
}

func getDistrict(geoRepo geo.Repository) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Add("Content-Type", "application/json; charset=utf-8")
		h.Add("Access-Control-Allow-Origin", "*")
		county := mux.Vars(r)["county"]
		areaList := geoRepo.GetAreaList(county)
		data, _ := json.Marshal(areaList)
		w.Write(data)
	}
}

func getStreet(geoRepo geo.Repository) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Add("Content-Type", "application/json; charset=utf-8")
		h.Add("Access-Control-Allow-Origin", "*")
		county := mux.Vars(r)["county"]
		district := mux.Vars(r)["district"]
		g := geoRepo.GetGeo(county, district)
		data, _ := json.Marshal(g)
		w.Write(data)
	}
}
