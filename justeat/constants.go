package justeat

import (
	"errors"
)

var (
	UnrecognizedCountry = errors.New("invalid area code passed (unrecognized country)")
	AddressNotFound     = errors.New("no address was labeled demae")
)

const (
	QueryUserData   = `SELECT authentication, expires_at, refresh_token, acr, device_model FROM users WHERE wii_id = $1`
	UpdateAuthToken = `UPDATE users SET authentication = $1, refresh_token = $2, expires_at = $3 WHERE wii_id = $4`
	UpdateBasket    = `UPDATE users SET basket = $1 WHERE wii_id = $2`
	InsertUser      = `INSERT INTO users (authentication, expires_at, refresh_token, acr, device_model, email, wii_id) VALUES ($1, $2, $3, $4, $5, $6, $7)`
)

// MaxNumberOfRestaurants is required due to Wii memory constraints.
const MaxNumberOfRestaurants = 15

// Country is one that is supported by Just Eat
type Country string

const (
	Austria       Country = "AT"
	Germany       Country = "DE"
	Ireland       Country = "IE"
	Italy         Country = "IT"
	Spain         Country = "ES"
	UnitedKingdom Country = "UK"
	Invalid       Country = ""
)

var ClientNames = map[Country]string{
	Austria:       "consumer_android_je",
	Germany:       "consumer_android_je",
	Ireland:       "consumer_android_je",
	Italy:         "consumer_android_je",
	Spain:         "consumer_android_je",
	UnitedKingdom: "consumer_android_je",
}

var ClientUUIDs = map[Country]string{
	Austria:       "50158598-42d0-41e4-aaff-9c5419c82215",
	Germany:       "50158598-42d0-41e4-aaff-9c5419c82215",
	Ireland:       "50158598-42d0-41e4-aaff-9c5419c82215",
	Italy:         "50158598-42d0-41e4-aaff-9c5419c82215",
	Spain:         "50158598-42d0-41e4-aaff-9c5419c82215",
	UnitedKingdom: "50158598-42d0-41e4-aaff-9c5419c82215",
}

var LanguageCodes = map[Country]string{
	Austria:       "de-AT",
	Germany:       "de-DE",
	Ireland:       "en-IE",
	Italy:         "it-IT",
	Spain:         "es-ES",
	UnitedKingdom: "en-GB",
}

var KongAPIURLs = map[Country]string{
	Austria:       "https://rest.api.eu-central-1.production.jet-external.com",
	Germany:       "https://rest.api.eu-central-1.production.jet-external.com",
	Ireland:       "https://i18n.api.just-eat.io",
	Italy:         "https://i18n.api.just-eat.io",
	Spain:         "https://i18n.api.just-eat.io",
	UnitedKingdom: "https://uk.api.just-eat.io",
}

var GlobalMenuCDNURLs = map[Country]string{
	Italy:         "https://menu-globalmenucdn.justeat-int.com",
	UnitedKingdom: "https://menu-globalmenucdn.je-apis.com",
}

var CheckoutURLs = map[Country]string{
	UnitedKingdom: "https://app-android.just-eat.co.uk",
}

var timeZones = map[Country]string{
	Austria:       "Europe/Vienna",
	Germany:       "Europe/Berlin",
	Ireland:       "Europe/Dublin",
	Italy:         "Europe/Rome",
	Spain:         "Europe/Madrid",
	UnitedKingdom: "Europe/London",
}

var AuthenticationURLs = map[Country]string{
	Austria:       "https://auth.lieferando.at",
	Germany:       "https://auth.lieferando.de",
	Ireland:       "https://auth.just-eat.ie",
	Italy:         "https://auth.justeat.it",
	Spain:         "https://auth.just-eat.es",
	UnitedKingdom: "https://auth.just-eat.co.uk",
}

var BasketURLs = map[Country]string{
	Austria:       "https://lieferando.at",
	Germany:       "https://lieferando.de",
	Ireland:       "https://just-eat.ie",
	Italy:         "https://justeat.it",
	Spain:         "https://just-eat.es",
	UnitedKingdom: "https://just-eat.co.uk",
}

var CurrencyISOCodes = map[Country]string{
	Austria:       "€",
	Germany:       "€",
	Ireland:       "€",
	Italy:         "€",
	Spain:         "€",
	UnitedKingdom: "£",
}
