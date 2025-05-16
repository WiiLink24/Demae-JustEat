package justeat

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/WiiLink24/DemaeJustEat/demae"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func (j *JEClient) getPayPalToken(storeId string, basketId string) (*BrainTreeConfig, error) {
	payload := url.Values{
		"serviceType": []string{"delivery"},
		"checkoutId":  []string{basketId},
	}

	resp, err := j.httpGet(fmt.Sprintf("%s/payment/%s/partners/%s/options?%s", j.KongAPIURL, strings.ToLower(string(j.Country)), storeId, payload.Encode()))
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)

	var payment PaymentTypes
	err = json.Unmarshal(data, &payment)
	if err != nil {
		return nil, err
	}

	for _, paymentType := range payment.AvailablePaymentTypes {
		if paymentType.PaymentType != "paypal_braintree" || paymentType.Status != "available" {
			continue
		}

		tokenData, err := base64.StdEncoding.DecodeString(paymentType.AdditionalData.ClientKey)
		if err != nil {
			return nil, err
		}

		var config BrainTreeConfig
		err = json.Unmarshal(tokenData, &config)
		if err != nil {
			return nil, err
		}

		return &config, nil
	}

	return nil, PaypalUnavailable
}

func (j *JEClient) getOrderID(basketID string, amount int, currency string) (string, error) {
	payload := map[string]any{
		"currency":  currency,
		"total":     amount,
		"returnUrl": fmt.Sprintf("%s/notify/checkout_payment", j.CheckoutURL),
		"methods": []map[string]any{
			{
				"type":   "paypal_braintree",
				"amount": amount,
			},
		},
	}

	resp, err := j.httpPost(fmt.Sprintf("%s/checkout/%s/%s/payments", j.KongAPIURL, strings.ToLower(string(j.Country)), basketID), payload)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var res map[string]any
	err = json.Unmarshal(data, &res)
	if err != nil {
		return "", err
	}

	return res["status"].(map[string]any)["native"].(map[string]any)["identifier"].(string), nil
}

func (j *JEClient) makePaypalURL(config *BrainTreeConfig, brainTree BrainTreeCreatePaypal) (*BrainTreePaymentResourceHead, error) {
	headers := map[string]string{
		"User-Agent":   "braintree/android/4.44.0",
		"Content-Type": "application/json",
	}

	resp, err := j.BrainTreePOST(fmt.Sprintf("%s/v1/paypal_hermes/create_payment_resource", config.ClientAPIUrl), brainTree, headers)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	fmt.Println(string(data))
	if resp.StatusCode != http.StatusCreated {
		return nil, PaypalURLFailed
	}

	var resource BrainTreePaymentResourceHead
	err = json.Unmarshal(data, &resource)
	return &resource, err
}

func (j *JEClient) sendPaypalMetadata() (*PaypalMetadata, error) {
	meta := PaypalMetadata{
		AppGUID:             "ae402f35-6b59-4f64-b5ba-6b26565ad795",
		AppID:               "JUST-EAT-APP",
		AndroidID:           demae.IDGenerator(16, "0123456789abcdef"),
		AppVersion:          "11.0.0.1610004768",
		AppFirstInstallTime: demae.RandIntWRange(1704277731000, 1745762494502),
		AppLastUpdateTime:   demae.RandIntWRange(1704277731000, 1745762494502),
		ConfURL:             "https://www.paypalobjects.com/rdaAssets/magnes/magnes_android_rec.json",
		CompVersion:         "5.5.0.release",
		DeviceModel:         "SM-F946B",
		DeviceName:          "Samsung SM-F946B",
		GSFID:               demae.IDGenerator(16, "0123456789abcdef"),
		IsEmulator:          false,
		EF:                  "00000",
		IsRooted:            false,
		RF:                  "0000101",
		OSType:              "Android",
		OSVersion:           "11",
		PayloadType:         "full",
		SMSEnabled:          true,
		MagnesGUID: struct {
			ID        string `json:"id"`
			CreatedAt int    `json:"created_at"`
		}{
			ID:        demae.UUID(),
			CreatedAt: demae.RandIntWRange(1704277731000, 1745762494502),
		},
		MagnesSource:      12,
		SourceAppVersion:  "11.0.0.1610004768",
		TotalStorageSpace: demae.RandIntWRange(50000000000, 59999999999),
		T:                 false,
		PairingID:         demae.IDGenerator(32, "0123456789abcdef"),
		ConnType:          "WIFI",
		ConfVersion:       "5.0",
		DMO:               true,
		DCID:              demae.IDGenerator(32, "0123456789abcdef"),
		DeviceUptime:      demae.RandIntWRange(400000, 700000),
		IpAddrs:           "192.168.1.100",
		IpAddresses:       []string{"192.168.1.100"},
		LocaleCountry:     string(j.Country),
		LocaleLang:        strings.Split(languageCodes[j.Country], "-")[0],
		PhoneType:         "gsm",
		RiskCompSessionID: demae.UUID(),
		Roaming:           false,
		SimOperatorName:   "",
		Timestamp:         int(time.Now().Unix() * 1000),
		// Doesn't matter
		TZName:          "Central European Standard Time",
		DS:              false,
		TZ:              7200000,
		NetworkOperator: "",
		ProxySetting:    "",
		MGID:            demae.IDGenerator(32, "0123456789abcdef"),
		PL:              "000100",
		SR: struct {
			AC bool `json:"ac"`
			GY bool `json:"gy"`
			MG bool `json:"mg"`
		}{
			AC: true,
			GY: true,
			MG: true,
		},
	}

	headers := map[string]string{
		// TODO: Cycle through device models
		"device_model":                  "SM-F946B",
		"app_version":                   "11.0.0.1610004768",
		"X-PAYPAL-RESPONSE-DATA-FORMAT": "NV",
		"os_version":                    "11",
		"os_type":                       "Android",
		"X-PAYPAL-REQUEST-DATA-FORMAT":  "NV",
		"X-PAYPAL-SERVICE-VERSION":      "1.0.0",
		"app_id":                        "JUST-EAT-APP",
		"comp_version":                  "5.5.0.release",
		"Content-Type":                  "application/x-www-form-urlencoded",
		"User-Agent":                    "Dalvik/2.1.0 (Linux; U; Android 11; SM-F946B Build/RQ3A.211001.001)",
	}

	// We have to encode as x-www-form-urlencoded.
	// This means first encoding PaypalMeta to a json string.
	metaStr, err := json.Marshal(meta)
	if err != nil {
		return nil, err
	}

	values := url.Values{}
	values.Set("additionalData", string(metaStr))
	values.Set("appGuid", meta.AppGUID)
	values.Set("libraryVersion", "Dyson/5.5.0.RELEASE (ANDROID 11)")

	resp, err := j.PayPalPOST("https://c.paypal.com/r/v1/device/client-metadata", values, headers)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	fmt.Println(string(data))
	return &meta, nil
}

func MakePaypalReturnURLFirst(head BrainTreePaymentResourceHead) string {
	return fmt.Sprintf("customer-details-oneapp.braintree://onetouch/v1/success?token=%s", head.PaymentResource.PaymentToken)
}

func MakePaypalReturnURL(token string, payerID string) string {
	return fmt.Sprintf("customer-details-oneapp.braintree://onetouch/v1/success?token=%s&PayerID=%s", token, payerID)
}

func (j *JEClient) GetPaypalNonce(config BrainTreeConfig, meta PaypalMetadata, fingerPrint string, returnURL string) (string, string, string, error) {
	header := map[string]string{
		"User-Agent":   "braintree/android/4.44.0",
		"Content-Type": "application/json",
	}

	payload := map[string]any{
		"_meta": map[string]any{
			"platform":    "android",
			"sessionId":   demae.IDGenerator(32, "0123456789abcdef"),
			"source":      "paypal-browser",
			"integration": "custom",
		},
		"paypalAccount": map[string]any{
			"correlationId": meta.PairingID,
			"intent":        "authorize",
			"options": map[string]any{
				"validate": false,
			},
			"client": map[string]string{},
			"response": map[string]string{
				"webURL": returnURL,
			},
			"response_type": "web",
		},
		"authorizationFingerprint": fingerPrint,
	}

	resp, err := j.BrainTreePOST(fmt.Sprintf("%s/v1/payment_methods/paypal_accounts", config.ClientAPIUrl), payload, header)
	if err != nil {
		return "", "", "", err
	}

	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", "", err
	}

	var m map[string]any
	err = json.Unmarshal(data, &m)
	if err != nil {
		return "", "", "", err
	}

	nonce := m["paypalAccounts"].([]any)[0].(map[string]any)["nonce"].(string)
	email := m["paypalAccounts"].([]any)[0].(map[string]any)["details"].(map[string]any)["payerInfo"].(map[string]any)["email"].(string)
	payerID := m["paypalAccounts"].([]any)[0].(map[string]any)["details"].(map[string]any)["payerInfo"].(map[string]any)["payerId"].(string)

	return nonce, email, payerID, nil
}

func (j *JEClient) SendPayment(meta PaypalMetadata, nonce string, email string, payerID string, orderID string) error {
	// Really weird, the payload requires this field to be JSON encoded, then gets encoded again.
	correlationId, _ := json.Marshal(map[string]any{"correlation_id": meta.PairingID})

	payload := map[string]any{
		"paymentMethod": "paypal_braintree",
		"identifier":    orderID,
		"additionalData": map[string]any{
			"payerEmail": email,
			"deviceData": string(correlationId),
			"payerId":    payerID,
		},
		"paymentToken": nonce,
	}

	resp, err := j.httpPost(fmt.Sprintf("%s/payment/%s/authorize", j.KongAPIURL, strings.ToLower(string(j.Country))), payload)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	fmt.Println(string(data))
	return nil
}
