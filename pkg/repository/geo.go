package repository

import (
	"database/sql"
	"encoding/json"
	"sort"
	"strconv"

	"github.com/pkg/errors"

	"github.com/nukr/geo/pkg/geo"
)

type geoRepo struct {
	countryList map[string][]string
	cityList    map[string][]string
	areaList    map[string][]string
	address     map[string]map[string]*geo.Geo
}

// NewGeoRepository 會初始化geoRepo
func NewGeoRepository(db *sql.DB) (geo.Repository, error) {
	g := geoRepo{
		countryList: make(map[string][]string),
		cityList:    make(map[string][]string),
		areaList:    make(map[string][]string),
		address:     make(map[string]map[string]*geo.Geo),
	}

	// 這邊將資料庫撈到的 country list 放進 geoRepo.countryList 這個 map 中，用 lang 當 key
	countryList, err := db.Query("SELECT country, lang from country_list")
	if err != nil {
		return nil, err
	}
	for countryList.Next() {
		var country, lang string
		err := countryList.Scan(&country, &lang)
		if err != nil {
			return nil, errors.Wrap(err, "countryList.Scan")
		}
		g.countryList[lang] = append(g.countryList[lang], country)
	}

	cityList, err := db.Query("SELECT city, lang from city_list")
	if err != nil {
		return nil, errors.Wrap(err, "db.Query cityList")
	}
	for cityList.Next() {
		var city, lang string
		err := cityList.Scan(&city, &lang)
		if err != nil {
			return nil, errors.Wrap(err, "cityList.Scan")
		}
		g.cityList[lang] = append(g.cityList[lang], city)
	}

	areaList, err := db.Query(`
	select distinct area, zip, city
	from (SELECT area, substr(zip, 1, 2) as zip, city from zip32) as zip32
	order by zip
	`)
	if err != nil {
		return nil, errors.Wrap(err, "db.Query areaList")
	}
	for areaList.Next() {
		var city, area, zip string
		err := areaList.Scan(&area, &zip, &city)
		if err != nil {
			return nil, errors.Wrap(err, "areaList.Scan")
		}
		g.areaList[city] = append(g.areaList[city], area)
	}

	address, err := db.Query(`
	select city, zip, area, jsonb_agg(road) as road from
	(select distinct road, area, city, substr(zip, 1, 3) as zip from zip32 order by zip) as zip32
	group by area, zip, city
	order by zip
	`)
	if err != nil {
		return nil, errors.Wrap(err, "db.Query address")
	}
	for address.Next() {
		var area, zip, city string
		var bsRoad []byte
		err := address.Scan(&city, &zip, &area, &bsRoad)
		if err != nil {
			return nil, errors.Wrap(err, "address.Scan")
		}
		iZip, err := strconv.Atoi(zip)
		if err != nil {
			return nil, errors.Wrap(err, "strconv.Atoi(zip)")
		}
		var road []string
		err = json.Unmarshal(bsRoad, &road)
		if err != nil {
			return nil, errors.Wrap(err, "json.Unmarshal(bsRoad, &road)")
		}
		sort.Strings(road)
		gg := &geo.Geo{
			Name:       area,
			Zip:        iZip,
			StreetName: road,
		}
		_, ok := g.address[city]
		if !ok {
			g.address[city] = make(map[string]*geo.Geo)
		}
		g.address[city][area] = gg
	}

	return &g, nil
}

func (g *geoRepo) GetCountryList(lang string) []string {
	return g.countryList[lang]
}

func (g *geoRepo) GetCityList(lang string) []string {
	return g.cityList[lang]
}

func (g *geoRepo) GetAreaList(city string) []string {
	return g.areaList[city]
}

func (g *geoRepo) GetGeo(city string, area string) *geo.Geo {
	return g.address[city][area]
}
