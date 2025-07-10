package middlewares

import (
	"errors"
	"fmt"
	initializer "go-write-behind-service/init"
	"go-write-behind-service/internal/models"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/spf13/viper"
)

type TokenClaims struct {
	Id string
	jwt.RegisteredClaims
}

func AuthorizationWithBearer() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var authenticated bool
		var identityAccess map[string]any
		var userID string

		var secretKey = []byte(viper.GetString("authorization.jwt.secretKey"))
		bearerToken := ctx.GetHeader("Authorization")
		if !strings.HasPrefix(bearerToken, "Bearer ") {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "tidak ada akses token yang disertakan",
				"message": "Mohon sertakan akses token untuk autentikasi pengguna",
			})
			ctx.Abort()
			return
		}
		bearerToken = strings.TrimPrefix(bearerToken, "Bearer ")
		tokenClaims := TokenClaims{}
		_, errParse := jwt.ParseWithClaims(bearerToken, &tokenClaims, func(token *jwt.Token) (any, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("signing method tidak dikenali: %v", token.Header["alg"])
			}
			return secretKey, nil
		})

		if errParse != nil {
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
					"message": "Secret key tidak valid, terdapat masalah pada secret key",
				})
				ctx.Abort()
				return
			case TokenExpired:
				ctx.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"error":   errParse.Error(),
					"message": "Akses token yang disertakan telah kadaluarsa",
				})
				ctx.Abort()
				return
			case TokenMalformed:
				ctx.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"error":   errParse.Error(),
					"message": "Akses token yang disertakan tidak valid, mohon periksa ulang",
				})
				ctx.Abort()
				return
			}
		}

		authenticated = true
		userID = tokenClaims.Id

		DB, errDB := initializer.GetMssqlDB()
		if errDB != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   errDB.Error(),
				"message": "Gagal memanggil GORM Instance pada middleware:AuthorizationWithBearer",
			})
			ctx.Abort()
			return
		}

		errIdentity := DB.Model(&models.IdentityAccess{}).Select([]string{
			"role_id",
			"type",
		}).
			Where("user_id = ?", tokenClaims.Id).
			First(&identityAccess).Error
		if errIdentity != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   errIdentity.Error(),
				"message": "Gagal melakukan cek terhadap identitas pengguna pada middleware:AuthorizationWithBearer",
			})
			ctx.Abort()
			return
		}

		ctx.Set("authenticated", authenticated)            // -> bool
		ctx.Set("identity-access", identityAccess["type"]) // -> string
		ctx.Set("token", bearerToken)                      // -> string
		ctx.Set("user-id", userID)                         // -> string

		ctx.Next()
	}
}
