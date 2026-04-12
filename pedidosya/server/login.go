package server

import (
	"crypto/rand"
	"encoding/base64"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"time"
)

func randString(nByte int) (string, error) {
	b := make([]byte, nByte)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func setCallbackCookie(w http.ResponseWriter, r *http.Request, name, value string) {
	c := &http.Cookie{
		Name:     name,
		Value:    value,
		MaxAge:   int(time.Hour.Seconds()),
		Secure:   r.TLS != nil,
		HttpOnly: true,
	}
	http.SetCookie(w, c)
}

func StartPanelHandler(c *gin.Context) {
	state, err := randString(16)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"Error": err.Error(),
		})
		return
	}

	setCallbackCookie(c.Writer, c.Request, "state", state)

	http.Redirect(c.Writer, c.Request, authConfig.OAuth2Config.AuthCodeURL(state), http.StatusFound)
}

func FinishPanelHandler(c *gin.Context) {
	state, err := c.Request.Cookie("state")
	if err != nil {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{
			"Error": "State cookie not found",
		})
		return
	}

	if c.Request.URL.Query().Get("state") != state.Value {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{
			"Error": "State did not match",
		})
		return
	}

	oauth2Token, err := authConfig.OAuth2Config.Exchange(c, c.Request.URL.Query().Get("code"))
	if err != nil {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{
			"Error": err.Error(),
		})
		return
	}

	rawIdtoken := oauth2Token.Extra("id_token").(string)
	if rawIdtoken == "" {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{
			"Error": "No id_token in response",
		})
		return
	}

	idToken, err := verifier.Verify(c, rawIdtoken)
	if err != nil {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{
			"Error": "Failed to verify id_token: " + err.Error(),
		})
		return
	}

	c.SetCookie("token", rawIdtoken, idToken.Expiry.Second(), "", "", false, true)

	//redirect to admin page
	c.Redirect(http.StatusMovedPermanently, "/pay")

}

func LoginPage(c *gin.Context) {
	c.HTML(http.StatusOK, "login.html", nil)
}
