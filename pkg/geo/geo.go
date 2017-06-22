package geo

// Repository provides access a geo store.
type Repository interface {
	GetCountry(string) []*Geo
	GetCounty(string) []*Geo
	GetDistrict(string, string) []*Geo
	GetStreet(string, string, string) []*Geo
}

// Geo ...
type Geo struct {
	Zip      string
	Country  string
	County   string
	District string
	Street   string
}
