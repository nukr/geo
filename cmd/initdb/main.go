package main

import (
	"database/sql"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
)

type XMLAddress struct {
	XMLName xml.Name `xml:"Zip32"`
	City    string   `xml:"City"`
	Area    string   `xml:"Area"`
	Road    string   `xml:"Road"`
	Zip     string   `xml:"Zip5"`
	Scope   string   `xml:"Scope"`
}

type XMLStruct struct {
	XMLName      xml.Name     `xml:"dataroot"`
	XMLAddresses []XMLAddress `xml:"Zip32"`
}

func main() {
	dsn := os.Getenv("POSTGRES_DSN")
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal(err)
	}
	stmt := `
	CREATE TABLE IF NOT EXISTS zip32 (
		zip varchar(64),
		country varchar(64),
		city varchar(64),
		area varchar(64),
		road varchar(64),
		scope varchar(64)
	);
	CREATE TABLE IF NOT EXISTS city_list (
		city varchar(64),
		lang varchar(64)
	);
	CREATE TABLE IF NOT EXISTS country_list (
		country varchar(64),
		lang varchar(64)
	);
	`
	_, err = db.Exec(stmt)
	if err != nil {
		log.Fatal(err)
	}
	start := time.Now()
	file, err := os.Open("./Xml_10706.xml")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	insertZip32(db, file)
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
				$1, $2 
			)
			`, z, f)
		}
	}
	filelist = []string{"zh-tw"}
	for _, f := range filelist {
		fd, err := os.Open(fmt.Sprintf("./city_list/%s.json", f))
		if err != nil {
			log.Fatal(err)
		}
		var zz []string
		bsFile, err := ioutil.ReadAll(fd)
		json.Unmarshal(bsFile, &zz)
		for _, z := range zz {
			_, err := db.Exec(`
			INSERT INTO city_list (
				city,
				lang
			) VALUES (
				$1, $2 
			)
			`, z, f)
			if err != nil {
				fmt.Println(err)
			}
		}
	}
	fmt.Println(time.Since(start))
}

func insertZip32(db *sql.DB, file io.Reader) {
	bsXML, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatal(err)
	}
	xs := XMLStruct{}
	xml.Unmarshal(bsXML, &xs)
	for _, d := range xs.XMLAddresses {
		db.Exec(`
		INSERT INTO zip32 (
			zip,
			country,
			city,
			area,
			road,
			scope
		) VALUES (
			$1, $2, $3, $4, $5, $6
		)
		`, d.Zip, "臺灣", d.City, d.Area, d.Road, d.Scope)
	}
}
