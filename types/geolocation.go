package types

import (
	"fmt"
	"math"

	"github.com/Kameleoon/client-go/v3/utils"
)

const geolocationEventType = "geolocation"

type Geolocation struct {
	duplicationUnsafeSendableBase
	country    string
	region     string
	city       string
	postalCode string
	latitude   float64
	longitude  float64
}

// Creates new Geolocation data.
//
// Parameters:
//   - country: Country.
//   - args: Region, City, Postal code.
func NewGeolocation(country string, args ...string) *Geolocation {
	region, city, postalCode := extractNewGeolocationArgs(args)
	return &Geolocation{
		country:    country,
		region:     region,
		city:       city,
		postalCode: postalCode,
		latitude:   math.NaN(),
		longitude:  math.NaN(),
	}
}

// Creates new Geolocation data.
//
// Parameters:
//   - latitude: Latitude.
//   - longitude: Longitude.
//   - country: Country.
//   - args: Region, City, Postal code.
func NewGeolocationWithCoords(latitude float64, longitude float64, country string, args ...string) *Geolocation {
	region, city, postalCode := extractNewGeolocationArgs(args)
	return &Geolocation{
		country:    country,
		region:     region,
		city:       city,
		postalCode: postalCode,
		latitude:   latitude,
		longitude:  longitude,
	}
}

func (g Geolocation) String() string {
	return fmt.Sprintf(
		"Geolocation{country:'%v',region:'%v',city:'%v',postal_code:'%v',latitude:%v,longitude:%v}",
		g.country,
		g.region,
		g.city,
		g.postalCode,
		g.latitude,
		g.longitude,
	)
}

func extractNewGeolocationArgs(args []string) (string, string, string) {
	var region, city, postalCode string
	if len(args) >= 1 {
		region = args[0]
		if len(args) >= 2 {
			city = args[1]
			if len(args) >= 3 {
				postalCode = args[2]
			}
		}
	}
	return region, city, postalCode
}

func (g *Geolocation) dataRestriction() {
	// This method is required to separate external type `Data` from `BaseData` types
}

func (g *Geolocation) Country() string {
	return g.country
}

func (g *Geolocation) Region() string {
	return g.region
}

func (g *Geolocation) City() string {
	return g.city
}

func (g *Geolocation) PostalCode() string {
	return g.postalCode
}

func (g *Geolocation) Latitude() float64 {
	return g.latitude
}

func (g *Geolocation) Longitude() float64 {
	return g.longitude
}

func (g *Geolocation) QueryEncode() string {
	qb := utils.NewQueryBuilder()
	qb.Append(utils.QPEventType, geolocationEventType)
	qb.Append(utils.QPCountry, g.country)
	qb.Append(utils.QPRegion, g.region)
	qb.Append(utils.QPCity, g.city)
	qb.Append(utils.QPPostalCode, g.postalCode)
	if !(math.IsNaN(g.latitude) || math.IsNaN(g.longitude)) {
		qb.Append(utils.QPLatitude, fmt.Sprintf("%f", g.latitude))
		qb.Append(utils.QPLongitude, fmt.Sprintf("%f", g.longitude))
	}
	qb.Append(utils.QPNonce, g.Nonce())
	return qb.String()
}

func (g *Geolocation) DataType() DataType {
	return DataTypeGeolocation
}
