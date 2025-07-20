package server

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/WiiLink24/DemaeJustEat/justeat"
	"github.com/gin-gonic/gin"
	"io"
	"net"
	"net/http"
	"slices"
	"strconv"
	"time"
)

func getGenericHeaders(c *gin.Context) (map[string]string, error) {
	deviceId := c.Query("device_id")
	if deviceId == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
		})
		return nil, fmt.Errorf("device_id is required")
	}

	country := c.Query("country")
	if country == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
		})
		return nil, fmt.Errorf("country is required")
	}

	countryConv := justeat.Country(country)
	authorization := fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", justeat.ClientNames[countryConv], justeat.ClientUUIDs[countryConv]))))

	return map[string]string{
		"User-Agent":                fmt.Sprintf("[JUST-EAT-APP/%s/Android - %s - 11 (API 30)]", justeat.ApplicationVersion, deviceId),
		"Application-Id":            justeat.ApplicationID,
		"Accept-Language":           justeat.LanguageCodes[countryConv],
		"Accept-Charset":            "utf-8",
		"Accept-Tenant":             country,
		"Accept":                    justeat.Accept,
		"Accept-Version":            justeat.AcceptVersion,
		"X-Jet-Application-Id":      justeat.JetApplicationID,
		"X-Jet-Application-Version": justeat.JetVersion,
		"Authorization":             authorization,
		"Content-Type":              "application/x-www-form-urlencoded",
	}, nil
}

func getLoginData(c *gin.Context) {
	country := c.Query("country")
	if country == "" {
		c.JSON(http.StatusBadRequest, gin.H{})
	}

	countryConv := justeat.Country(country)

	header, err := getGenericHeaders(c)
	if err != nil {
		return
	}

	payload := map[string]any{
		"grant_type": "password",
		"scope":      "openid mobile_scope offline_access",
		"tenant":     country,
		"client_id":  justeat.ClientNames[countryConv],
	}

	c.JSON(http.StatusOK, gin.H{
		"header":  header,
		"payload": payload,
		"url":     fmt.Sprintf("%s/connect/token", justeat.AuthenticationURLs[countryConv]),
	})
}

func get2FAData(c *gin.Context) {
	country := c.Query("country")
	if country == "" {
		c.JSON(http.StatusBadRequest, gin.H{})
	}

	countryConv := justeat.Country(country)
	header, err := getGenericHeaders(c)
	if err != nil {
		return
	}

	payload := map[string]any{
		"grant_type": "mfa_otp",
	}

	c.JSON(http.StatusOK, gin.H{
		"header":  header,
		"payload": payload,
		"url":     fmt.Sprintf("%s/connect/token", justeat.AuthenticationURLs[countryConv]),
	})
}

func saveUserData(c *gin.Context) {
	// First verify the Wii is linked to this account.
	wiis, _ := c.Get("wiis")
	wiiNoStr := c.PostForm("wii_no")

	if !slices.Contains(wiis.([]string), wiiNoStr) {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Wii Number not linked to account.",
		})
		return
	}

	// We can now link.
	auth := c.PostForm("eat_auth")
	refreshToken := c.PostForm("refresh_token")
	expireTime := c.PostForm("expire_time")
	deviceModel := c.PostForm("device_model")
	acr := c.PostForm("acr")

	intExpireTime, err := strconv.ParseInt(expireTime, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Expire time must be an integer",
		})
		return
	}

	expiresTimeObj := time.Unix(intExpireTime, 0).Unix()
	_, err = pool.Exec(ctx, justeat.InsertUser, auth, expiresTimeObj, refreshToken, acr, deviceModel)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	// Finally we have to send off to the Account Manager.
	type JustEatPayload struct {
		WiiNumber string `json:"wii_number"`
		Auth      string `json:"auth"`
	}

	socketPayload := JustEatPayload{
		WiiNumber: wiiNoStr,
		Auth:      c.GetHeader("Authorization"),
	}

	data, err := json.Marshal(socketPayload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	// Dial up.
	conn, err := net.Dial("unix", "/tmp/eater.sock")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	defer conn.Close()
	_, err = conn.Write(data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	// Read response
	resp := make([]byte, 1024)
	n, err := conn.Read(resp)
	if err != nil && !errors.Is(err, io.EOF) {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	var result map[string]any
	err = json.Unmarshal(resp[:n-1], &result)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	if !result["success"].(bool) {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": result["message"].(string),
		})
		return
	}

	// Linked!
	c.JSON(http.StatusOK, gin.H{
		"success": true,
	})
}
