package database

import (
	"database/sql"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
	"github.com/nukr/experiment/sqlite/pkg/geo"
)

type geoRepository struct {
	db *sql.DB
}

// NewGeoRepository ...
func NewGeoRepository() geo.Repository {
	os.Remove("./foo.db")
	db, err := sql.Open("sqlite3", "./foo.db")
	if err != nil {
		log.Fatal(err)
	}
	insertData(db)
	return &geoRepository{db: db}
}

func insertData(db *sql.DB) {
	// create table
	{
		sqlStmt := `
		CREATE TABLE geo (
			zip varchar(64),
			country varchar(64),
			county varchar(64),
			district varchar(64),
			street varchar(64)
		);
		CREATE TABLE country_list (
			country varchar(64),
			lang varchar(64)
		);
		`
		db.Exec(sqlStmt)
	}
	file, err := os.Open("./Xml_10510.xml")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	bsXML, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatal(err)
	}

	type XMLAddress struct {
		Zip            string `xml:"欄位1"`
		Street         string `xml:"欄位2"`
		Range          string `xml:"欄位3"`
		CountyDistrict string `xml:"欄位4"`
	}
	type XMLStruct struct {
		XMLName      xml.Name     `xml:"dataroot"`
		XMLAddresses []XMLAddress `xml:"Xml_10510"`
	}
	xs := XMLStruct{}
	xml.Unmarshal(bsXML, &xs)
	for i, d := range xs.XMLAddresses {
		county := string([]rune(d.CountyDistrict)[0:3])
		district := string([]rune(d.CountyDistrict)[3:])
		db.Exec(`
		INSERT INTO geo (
			zip,
			country,
			county,
			district,
			street
		) VALUES (
			?, ?, ?, ?, ?
		)
		`, d.Zip, "台灣", county, district, d.Street)
		fmt.Println(i)
	}
	filelist := []string{"zh-tw", "en-us"}
	for _, f := range filelist {
		fd, err := os.Open(fmt.Sprintf("./country_list/%s.json", f))
		if err != nil {
			log.Fatal(err)
		}
		var zz []string
		bsFile, err := ioutil.ReadAll(fd)
		json.Unmarshal(bsFile, &zz)
		for _, z := range zz {
			db.Exec(`
			INSERT INTO country_list (
				country,
				lang
			) VALUES (
				?, ?
			)
			`, z, f)
		}
	}
}

func (g *geoRepository) GetCountry(lang string) []*geo.Geo {
	rows, err := g.db.Query(`
	SELECT country FROM country_list WHERE lang=?
	`, lang)
	if err != nil {
		log.Fatal(err)
	}
	geos := make([]*geo.Geo, 0, 100)
	for rows.Next() {
		row := &geo.Geo{}
		rows.Scan(&row.Country)
		geos = append(geos, row)
	}
	return geos
}

func (g *geoRepository) GetCounty(country string) []*geo.Geo {
	rows, err := g.db.Query(`
	SELECT DISTINCT county FROM geo WHERE country=?
	`, country)
	if err != nil {
		log.Fatal(err)
	}
	geos := make([]*geo.Geo, 0, 100)
	for rows.Next() {
		row := &geo.Geo{}
		rows.Scan(&row.County)
		geos = append(geos, row)
	}
	return geos
}

func (g *geoRepository) GetDistrict(country, county string) []*geo.Geo {
	rows, err := g.db.Query(`
	SELECT DISTINCT district FROM geo WHERE country=? AND county=?
	`, country, county)
	if err != nil {
		log.Fatal(err)
	}
	geos := make([]*geo.Geo, 0, 100)
	for rows.Next() {
		row := &geo.Geo{}
		rows.Scan(&row.County)
		geos = append(geos, row)
	}
	return geos
}

func (g *geoRepository) GetStreet(country, county, district string) []*geo.Geo {
	rows, err := g.db.Query(`
	SELECT district, street, zip FROM geo WHERE country=? AND county=? AND district=?
	`, country, county, district)
	if err != nil {
		log.Fatal(err)
	}
	geos := make([]*geo.Geo, 0, 100)
	for rows.Next() {
		row := &geo.Geo{}
		rows.Scan(&row.District, &row.Street, &row.Zip)
		geos = append(geos, row)
	}
	return geos
}
