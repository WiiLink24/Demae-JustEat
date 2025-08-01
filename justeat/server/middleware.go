package server

import (
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gin-gonic/gin"
	"net/http"
)

type Claims struct {
	Email    string          `json:"email"`
	Username string          `json:"preferred_username"`
	Name     string          `json:"name"`
	UserId   string          `json:"sub"`
	Groups   []string        `json:"groups"`
	Wiis     []string        `json:"wiis"`
	WWFC     []string        `json:"wwfc"`
	Dominos  map[string]bool `json:"dominos"`
	JustEat  map[string]bool `json:"just_eat"`
}

func AuthenticationMiddleware(verifier *oidc.IDTokenVerifier) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, err := c.Cookie("token")
		if err != nil {
			c.Redirect(http.StatusTemporaryRedirect, "/login")
			c.Abort()
			return
		}

		// Verify the OpenID Connect idToken.
		idToken, err := verifier.Verify(ctx, tokenString)
		if err != nil {
			c.Redirect(http.StatusTemporaryRedirect, "/login")
			c.Abort()
			return
		}

		// Parse custom claims.
		var claims Claims
		if err = idToken.Claims(&claims); err != nil {
			c.Redirect(http.StatusTemporaryRedirect, "/login")
			c.Abort()
			return
		}

		c.Set("username", claims.Username)
		c.Set("user_id", claims.UserId)
		c.Set("email", claims.Email)
		c.Set("just_eat", claims.JustEat)
		c.Next()
	}
}

func AuthenticationLinkerMiddleware(verifier *oidc.IDTokenVerifier) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			// We can't redirect off an Unauthorized status code.
			c.Status(http.StatusBadRequest)
			c.Abort()
			return
		}

		// Verify the OpenID Connect idToken.
		idToken, err := verifier.Verify(ctx, tokenString)
		if err != nil {
			c.Status(http.StatusUnauthorized)
			c.Abort()
			return
		}

		// Parse custom claims if needed.
		var claims Claims
		if err = idToken.Claims(&claims); err != nil {
			c.Status(http.StatusInternalServerError)
			c.Abort()
			return
		}

		c.Set("uid", claims.UserId)
		c.Set("wiis", claims.Wiis)
		c.Set("email", claims.Email)
		c.Next()
	}
}
