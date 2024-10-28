package middlewares

import (
	"errors"
	"fmt"
	"future-interns-backend/internal/handlers"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/spf13/viper"
)

func AuthorizationWithBearer() gin.HandlerFunc {
	return func(context *gin.Context) {
		var secretKey = []byte(viper.GetString("authorization.jwt.secretKey"))
		bearerToken := context.GetHeader("Authorization")
		if !strings.HasPrefix(bearerToken, "Bearer ") {
			context.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "no token was provided",
				"message": "provide Authorization with Bearer token, kids",
			})
			context.Abort()
			return
		}
		bearerToken = strings.TrimPrefix(bearerToken, "Bearer ")
		tokenClaims := &handlers.TokenClaims{}
		_, errParse := jwt.ParseWithClaims(bearerToken, tokenClaims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return secretKey, nil
		})

		if errParse != nil {
			var (
				InvalidSignature = errors.Is(errParse, jwt.ErrTokenSignatureInvalid)
				TokenExpired     = errors.Is(errParse, jwt.ErrTokenExpired)
			)
			switch {
			case InvalidSignature:
				context.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"error":   errParse.Error(),
					"message": "invalid secret key, double check the secret key",
				})
				context.Abort()
				return
			case TokenExpired:
				context.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"error":   errParse.Error(),
					"message": "provided token was expired",
				})
				context.Abort()
				return
			}
		}
		context.Next()
	}
}
