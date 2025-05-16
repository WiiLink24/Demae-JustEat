package server

import (
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

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
		var claims struct {
			Username string `json:"preferred_username"`
			UserId   string `json:"user_id"`
			Email    string `json:"email"`
		}
		if err = idToken.Claims(&claims); err != nil {
			c.Redirect(http.StatusTemporaryRedirect, "/login")
			c.Abort()
			return
		}
		log.Println(claims.Username, claims.UserId)

		c.Set("username", claims.Username)
		c.Set("user_id", claims.UserId)
		c.Set("email", claims.Email)
		c.Next()
	}
}
