package handlers

import (
	"errors"
	"fmt"
	initializer "future-interns-backend/init"
	"future-interns-backend/internal/models"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/spf13/viper"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
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

func GormErrType(err error) map[string]any {
	var (
		ErrDuplicatedKey        bool = errors.Is(err, gorm.ErrDuplicatedKey)
		ErrInvalidValueOfLength bool = errors.Is(err, gorm.ErrInvalidValueOfLength)
	)
	switch {
	case ErrDuplicatedKey:
		return map[string]any{
			"success": false,
			"error":   err.Error(),
			"message": "The data you provided must not be the same as the data already stored.",
		}
	case ErrInvalidValueOfLength:
		return map[string]any{
			"success": false,
			"error":   err.Error(),
			"message": "Some data exceeds the length limit of the database column.",
		}
	}
	return map[string]any{
		"success": false,
		"error":   err.Error(),
		"message": "record with your provided email already exists in database.",
	}
}

const (
	TokenExpiration  = 6 * time.Hour
	CandidateRoleId  = 1
	EmployerRoleId   = 2
	UniversityRoleId = 3
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
			"success": false,
			"error":   user.Error.Error(),
			"message": message,
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

	/*
		check identities table to determine role
		1. create sdkadmin user on migration -> have all access
	*/
	identityAccess := map[string]interface{}{}
	errGetUserRole := gormDB.Model(&models.IdentityAccess{}).
		Select([]string{"roles.name", "roles.description", "identity_accesses.type"}).
		Joins("INNER JOIN roles ON roles.id = identity_accesses.role_id").
		Where("user_id = ?", m_user.Id).
		First(&identityAccess).Error

	if errGetUserRole != nil {
		context.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   errGetUserRole.Error(),
			"message": fmt.Sprintf("unidentified user with email %s; your account has no access to any resources", auth.Email),
		})

		context.Abort()
		return
	}
	/* sign token */
	token := GenerateToken(m_user.Id, TokenExpiration)

	context.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"access_token": token,
			"user_id":      m_user.Id,
			"role":         identityAccess,
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
	errTx := gormDB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&user).Error; err != nil {
			return err
		}
		identityAccess := &models.IdentityAccess{
			UserId: user.Id,
			RoleId: CandidateRoleId,
			Type:   "candidate",
		}
		if err := tx.Create(&identityAccess).Error; err != nil {
			return err
		}

		return nil
	})
	/* Error Handling Database Operation */
	if errTx != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   errTx.Error(),
			"message": errTx.Error(),
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

func (h *AccountsHandler) UserInformation(context *gin.Context) {
	identities, _ := context.Get("identity-accesses")
	permissions, _ := context.Get("permissions")

	bearerToken := strings.TrimPrefix(context.GetHeader("Authorization"), "Bearer ")
	claims := ParseJWT(bearerToken)

	gormDB, _ := initializer.GetGorm()
	user_data := map[string]interface{}{}
	if errSearchUser := gormDB.Model(&models.User{}).
		Select([]string{
			"fullname",
			"email",
		}).
		Where("id = ?", claims.Id).
		First(&user_data).Error; errSearchUser != nil {
		message := fmt.Sprintf("user with id (%s) doesn't exist", claims.Id)
		context.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   errSearchUser.Error(),
			"message": message,
		})
		context.Abort()
		return
	}

	context.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"user":        user_data,
			"identity":    identities,
			"permissions": permissions,
		},
	})
}
