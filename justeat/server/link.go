package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"slices"
	"strconv"
	"time"

	"github.com/WiiLink24/DemaeJustEat/justeat"
	"github.com/WiiLink24/DemaeJustEat/logger"
	"github.com/WiiLink24/nwc24"
	"github.com/gin-gonic/gin"
)

func getLoginData(c *gin.Context) {
	country := c.Query("country")
	if country == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "country is required",
		})
		return
	}

	countryConv := justeat.Country(country)

	c.JSON(http.StatusOK, gin.H{
		"eater_url": justeat.BasketURLs[countryConv],
		"token_url": fmt.Sprintf("%s/connect/token", justeat.AuthenticationURLs[countryConv]),
	})
}

func saveUserData(c *gin.Context) {
	// First verify the Wii is linked to this account.
	_wiis, _ := c.Get("wiis")
	wiiNoStr := c.PostForm("wii_number")

	wiis := _wiis.([]Wii)
	if !slices.ContainsFunc(wiis, func(w Wii) bool {
		return w.WiiNumber == wiiNoStr
	}) {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Wii Number not linked to account.",
		})
	}

	// Convert Wii number to object.
	// Since the Wii is linked, we can skip any checks.
	wiiNoInt, _ := strconv.ParseUint(wiiNoStr, 10, 64)
	wiiNo := nwc24.LoadWiiNumber(wiiNoInt)

	// We can now link.
	auth := c.PostForm("eat_auth")
	refreshToken := c.PostForm("refresh_token")
	expireTime := c.PostForm("expire_time")
	deviceModel := c.PostForm("device_model")
	acr := c.PostForm("acr")
	email, _ := c.Get("email")

	intExpireTime, err := strconv.ParseInt(expireTime, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Expire time must be an integer",
		})
		return
	}

	expiresTimeObj := time.Unix(intExpireTime, 0).UTC()
	_, err = pool.Exec(ctx, justeat.InsertUser, auth, expiresTimeObj, refreshToken, acr, deviceModel, email, strconv.Itoa(int(wiiNo.GetHollywoodID())))
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

	defer func(conn net.Conn) {
		err = conn.Close()
		if err != nil {
			logger.Error("WEBSERVER", err.Error())
			return
		}
	}(conn)
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
