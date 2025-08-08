package grubhub

import (
	"errors"
)

var (
	UnrecognizedCountry = errors.New("invalid area code passed (outside of the USA)")
	AddressNotFound     = errors.New("no address was labeled demae")
)

const (
	QueryUserData   = `SELECT authentication, expires_at, refresh_token, acr, device_model FROM users WHERE wii_id = $1`
	UpdateAuthToken = `UPDATE users SET authentication = $1, refresh_token = $2, expires_at = $3 WHERE wii_id = $4`
	InsertUser      = `INSERT INTO users (authentication, expires_at, refresh_token, acr, device_model, email, wii_id) VALUES ($1, $2, $3, $4, $5, $6, $7)`
)

// MaxNumberOfRestaurants is required due to Wii memory constraints.
const MaxNumberOfRestaurants = 15

type Country string

const (
	UnitedStates Country = "US"
	Invalid      Country = ""
)

var ClientNames = map[Country]string{
	UnitedStates: "consumer_android_gh",
	Invalid:      "consumer_android_gh",
}

var ClientUUIDs = map[Country]string{
	UnitedStates: "50158598-42d0-41e4-aaff-9c5419c82215",
}

var LanguageCodes = map[Country]string{
	UnitedStates: "en-US",
}

var KongAPIURLs = map[Country]string{
	UnitedStates: "https://api-third-party-gtm.grubhub.com",
}

var GlobalMenuCDNURLs = map[Country]string{
	UnitedStates: "https://assets.grubhub.com",
}

var CheckoutURLs = map[Country]string{
	UnitedStates: "https://grubhub.com",
}

var timeZones = map[Country]string{
	UnitedStates: "America/New_York",
}

var AuthenticationURLs = map[Country]string{
	UnitedStates: "https://auth.grubhub.com",
}
