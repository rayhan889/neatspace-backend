package middlewares

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jwt"
	"github.com/rayhan889/neatspace/pkg/apputils"
)

func JWTMiddleware(secret []byte, alg jwa.SignatureAlgorithm) fiber.Handler {
	return func(c *fiber.Ctx) error {
		auth := c.Get("Authorization")
		if auth == "" {
			return fiber.NewError(fiber.StatusUnauthorized, "missing authorization header")
		}

		parts := strings.SplitN(auth, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid authorization header format")
		}

		tokenStr := strings.TrimSpace(parts[1])

		jwtGen := apputils.NewJWTGenerator(apputils.JWTConfig{
			SecretKey:  secret,
			SigninAlgo: alg,
		})

		claims, err := jwtGen.ParseAndValidate(c.Context(), tokenStr)
		if err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, fmt.Sprintf("invalid token :%v", err))
		}

		if t, ok := claims["type"]; ok {
			if tStr, ok := t.(string); ok && tStr != "access" {
				return fiber.NewError(fiber.StatusUnauthorized, "token is not an access token")
			}
		}

		c.Locals("jwt_claims", claims)
		log.Println("claims", claims)

		if sub, ok := claims[jwt.SubjectKey]; ok {
			c.Locals("user_id", fmt.Sprint(sub))
		}

		if sid, ok := claims["sid"]; ok {
			c.Locals("session_id", fmt.Sprint(sid))
		} else if sid2, ok := claims["SID"]; ok {
			c.Locals("session_id", fmt.Sprint(sid2))
		}
		if aud, ok := claims["aud"]; ok {
			c.Locals("audience", fmt.Sprint(aud))
		}

		c.Locals("jwt_raw", tokenStr)

		ctx := context.WithValue(c.Context(), apputils.JwtClaimsContextKey, claims)
		c.SetUserContext(ctx)

		return c.Next()
	}
}
