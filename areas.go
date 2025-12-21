package main

import (
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"

	"github.com/WiiLink24/DemaeJustEat/demae"
	"github.com/WiiLink24/DemaeJustEat/justeat"
)

// Supported countries
var (
	supportedCountries    = []string{"Austria", "Germany", "Ireland", "Italy", "Spain", "United Kingdom"}
	supportedCountriesMap = map[string]string{
		"AT": "Austria",
		"DE": "Germany",
		"IE": "Ireland",
		"IT": "Italy",
		"ES": "Spain",
		"GB": "United Kingdom",
	}
	supportedCountriesCodes = []justeat.Country{justeat.Austria, justeat.Germany, justeat.Ireland, justeat.Italy, justeat.Spain, justeat.UnitedKingdom}
	adminCodeNameCache      = map[string]string{}
)

func GetAdministrativeRegions(country justeat.Country) []demae.AreaNames {
	var areaNames []demae.AreaNames

	// Admin code for UK is GB.
	adminCode := string(country)
	if country == justeat.UnitedKingdom {
		adminCode = "GB"
	}

	for _, state := range geonameStates {
		if !strings.HasPrefix(state.Codes, adminCode) {
			continue
		}

		// Ex: GB.ENG (United Kingdom, England)
		if _, ok := adminCodeNameCache[state.Codes]; !ok {
			adminCodeNameCache[state.Codes] = state.Name
		}

		areaNames = append(areaNames, demae.AreaNames{
			AreaName: demae.CDATA{Value: state.Name},
			AreaCode: demae.CDATA{Value: state.Codes},
		})
	}

	return areaNames
}

func GetCitiesByAdminCode(stateCode, areaCode string) []demae.Area {
	var cities []demae.Area

	for _, city := range geonameCities {
		codes := strings.Split(stateCode, ".")
		if city.CountryCode != codes[0] || city.Admin1Code != codes[1] {
			continue
		}

		if *city.Population > 50000 {
			cities = append(cities, demae.Area{
				AreaName:   demae.CDATA{Value: city.Name},
				AreaCode:   demae.CDATA{Value: areaCode},
				IsNextArea: demae.CDATA{Value: 0},
				Display:    demae.CDATA{Value: 1},
				Kanji1:     demae.CDATA{Value: supportedCountriesMap[codes[0]]},
				Kanji2:     demae.CDATA{Value: city.Name},
				Kanji3:     demae.CDATA{Value: ""},
				Kanji4:     demae.CDATA{Value: ""},
			})
		}
	}

	return cities
}

func areaList(r *Response) {
	areaCode := r.request.URL.Query().Get("areaCode")

	// Nintendo, for whatever reason, require a separate "selectedArea" element
	// as a root node within output.
	// This violates about every XML specification in existence.
	// I am reasonably certain there was a mistake as their function to
	// interpret nodes at levels accepts a parent node, to which they seem to
	// have passed NULL instead of response.
	//
	// We are not going to bother spending time to deal with this.
	if r.request.URL.Query().Get("zipCode") != "" {
		version, apiStatus := GenerateVersionAndAPIStatus()
		r.ResponseFields = []any{
			demae.KVFieldWChildren{
				XMLName: xml.Name{Local: "response"},
				Value: []any{
					demae.KVFieldWChildren{
						XMLName: xml.Name{Local: "areaList"},
						Value: []any{
							demae.KVField{
								XMLName: xml.Name{Local: "segment"},
								Value:   "United States",
							},
							demae.KVFieldWChildren{
								XMLName: xml.Name{Local: "list"},
								Value: []any{
									demae.KVFieldWChildren{
										XMLName: xml.Name{Local: "areaPlace"},
										Value: []any{demae.AreaNames{
											AreaName: demae.CDATA{Value: "place name"},
											AreaCode: demae.CDATA{Value: 2},
										}},
									},
								},
							},
						},
					},
					demae.KVField{
						XMLName: xml.Name{Local: "areaCount"},
						Value:   "1",
					},
					version,
					apiStatus,
				},
			},
			demae.KVFieldWChildren{
				XMLName: xml.Name{Local: "selectedArea"},
				Value: []any{
					demae.KVField{
						XMLName: xml.Name{Local: "areaCode"},
						Value:   1,
					},
				},
			},
		}
		return
	}

	if areaCode == "0" {
		var countriesList []any
		for i, country := range supportedCountries {
			countriesList = append(countriesList, demae.KVFieldWChildren{
				XMLName: xml.Name{Local: fmt.Sprintf("place%d", i)},
				Value: []any{
					demae.KVField{
						XMLName: xml.Name{Local: "segment"},
						Value:   country,
					},
					demae.KVFieldWChildren{
						XMLName: xml.Name{Local: "list"},
						Value: []any{
							GetAdministrativeRegions(supportedCountriesCodes[i])[:],
						},
					},
				},
			})
		}
		r.AddKVWChildNode("areaList", countriesList)
		r.AddKVNode("areaCount", strconv.Itoa(len(countriesList)))
		return
	}

	newAreaCode := demae.IDGenerator(10, "0123456789")
	cities := GetCitiesByAdminCode(areaCode, newAreaCode)
	r.AddKVWChildNode("areaList", demae.KVFieldWChildren{
		XMLName: xml.Name{Local: "place"},
		Value: []any{
			demae.KVField{
				XMLName: xml.Name{Local: "container0"},
				Value:   "aaaaa",
			},
			demae.KVField{
				XMLName: xml.Name{Local: "segment"},
				Value:   adminCodeNameCache[areaCode],
			},
			demae.KVFieldWChildren{
				XMLName: xml.Name{Local: "list"},
				Value: []any{
					cities[:],
				},
			},
		},
	})
	r.AddKVNode("areaCount", strconv.FormatInt(int64(len(cities)), 10))
}
