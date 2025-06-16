package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"rtglabs-go/ent"
	"rtglabs-go/ent/session"
)

func AuthMiddleware(client *ent.Client) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			auth := c.Request().Header.Get("Authorization")
			if !strings.HasPrefix(auth, "Bearer ") {
				return echo.NewHTTPError(http.StatusUnauthorized, "Missing bearer token")
			}
			token := strings.TrimPrefix(auth, "Bearer ")

			sess, err := client.Session.
				Query().
				Where(session.TokenEQ(token)).
				WithUser().
				Only(c.Request().Context())

			if err != nil || sess.ExpiresAt.Before(time.Now()) {
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid or expired token")
			}

			c.Set("user", sess.Edges.User)
			c.Set("token", token)

			return next(c)
		}
	}
}
