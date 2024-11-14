package middlewares

import (
	"errors"
	"fmt"
	"future-interns-backend/internal/handlers"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/spf13/viper"
)

func AuthorizationWithBearer() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var secretKey = []byte(viper.GetString("authorization.jwt.secretKey"))
		bearerToken := ctx.GetHeader("Authorization")
		if !strings.HasPrefix(bearerToken, "Bearer ") {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "no token was provided",
				"message": "provide Authorization with Bearer token, kids",
			})
			ctx.Abort()
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
			log.Println("ERROR PARSE \t:", errParse.Error())
			var (
				InvalidSignature = errors.Is(errParse, jwt.ErrTokenSignatureInvalid)
				TokenExpired     = errors.Is(errParse, jwt.ErrTokenExpired)
				TokenMalformed   = errors.Is(errParse, jwt.ErrTokenMalformed)
			)
			switch {
			case InvalidSignature:
				ctx.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"error":   errParse.Error(),
					"message": "invalid secret key, double check the secret key",
				})
				ctx.Abort()
				return
			case TokenExpired:
				ctx.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"error":   errParse.Error(),
					"message": "provided token was expired",
				})
				ctx.Abort()
				return
			case TokenMalformed:
				ctx.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"error":   errParse.Error(),
					"message": "invalid token value, double check your token",
				})
				ctx.Abort()
				return
			}
		}
		ctx.Next()
	}
}
