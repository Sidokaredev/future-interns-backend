package handlers

import (
	"fmt"
	initializer "future-interns-backend/init"
	"future-interns-backend/internal/models"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/spf13/viper"
	"golang.org/x/crypto/bcrypt"
)

type AccountsHandler struct {
}

type Auth struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type Register struct {
	Fullname string `json:"fullname" binding:"required"`
	Auth
}

type TokenClaims struct {
	Id string
	jwt.RegisteredClaims
}

/* helpers */
func GenUuid(fullname string, ch_uuid chan string) {
	namespace := uuid.Must(uuid.NewRandom())
	data := []byte(fullname)

	sha1ID := uuid.NewSHA1(namespace, data)
	ch_uuid <- sha1ID.String()
	close(ch_uuid)
}

func HashPassword(password string, ch_hashedPassword chan string) {
	hashed, errHash := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if errHash != nil {
		log.Fatalf("failed to hash password: %v", errHash)
	}
	ch_hashedPassword <- string(hashed)
	close(ch_hashedPassword)
}

func GenerateToken(userId string, expires time.Duration) string {
	secretKey := viper.GetString("authorization.jwt.secretKey")
	claims := TokenClaims{
		Id: userId,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "barjakoub",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expires)),
		},
	}
	tokenizer := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, _ := tokenizer.SignedString([]byte(secretKey))

	return token
}

const (
	TokenExpiration = 6 * time.Hour
)

func (h *AccountsHandler) Auth(context *gin.Context) {
	var (
		auth   Auth
		m_user models.User
	)
	/* bind request */
	if errBind := context.ShouldBindJSON(&auth); errBind != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_info": errBind.Error(),
			"message":    "double check your JSON fields, kids",
		})
		return
	}
	gormDB, _ := initializer.GetGorm()
	/* query database */
	user := gormDB.Where("email = ?", auth.Email).First(&m_user)
	if user.Error != nil {
		message := fmt.Sprintf("account with email (%s) does not exist", auth.Email)
		context.JSON(http.StatusBadRequest, gin.H{
			"success":  false,
			"error":    user.Error.Error(),
			"messsage": message,
		})
		return
	}
	/* compare password */
	errComparePassword := bcrypt.CompareHashAndPassword([]byte(m_user.Password), []byte(auth.Password))
	if errComparePassword != nil {
		message := fmt.Sprintf("invalid password for (%s), double check please", auth.Email)
		context.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   errComparePassword.Error(),
			"message": message,
		})
		return
	}
	/* sign token */
	token := GenerateToken(m_user.Id, TokenExpiration)

	context.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"access_token": token,
			"user_id":      m_user.Id,
		},
	})
}

func (h *AccountsHandler) RegisterAccount(context *gin.Context) {
	/* channel */
	ch_uuid := make(chan string)
	ch_hashedPassword := make(chan string)

	var account Register
	if errBind := context.ShouldBindJSON(&account); errBind != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_info": errBind.Error(),
			"message":    "double check your JSON fields, kids",
		})
		return
	}
	/* Concurrent */
	go GenUuid(account.Fullname, ch_uuid)
	go HashPassword(account.Password, ch_hashedPassword)

	gormDB, _ := initializer.GetGorm()
	user := models.User{
		Id:       <-ch_uuid,
		Fullname: account.Fullname,
		Email:    account.Email,
		Password: <-ch_hashedPassword,
	}
	createUser := gormDB.Create(&user)
	/* Error Handling Database Operation */
	if createUser.Error != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   createUser.Error.Error(),
		})
		return
	}
	/* sign token */
	token := GenerateToken(user.Id, TokenExpiration)

	context.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"user_id":      user.Id,
			"access_token": token,
		},
	})
}
