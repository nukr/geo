package geo

// Repository provides access a geo store.
type Repository interface {
	GetCountryList(lang string) []string
	GetCityList(country string) []string
	GetAreaList(city string) []string
	GetGeo(city, area string) *Geo
}

// Geo ...
type Geo struct {
	Name       string   `json:"name"`
	Zip        int      `json:"zip"`
	StreetName []string `json:"street_name"`
}
