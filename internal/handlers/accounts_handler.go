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

	"github.com/brianvoe/gofakeit/v7"
	"github.com/brianvoe/gofakeit/v7/source"
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
	TokenExpiration   = 6 * time.Hour
	CANDIDATE_ROLE_ID = 1
	EmployerRoleId    = 2
	UniversityRoleId  = 3
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
			RoleId: CANDIDATE_ROLE_ID,
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

func (h *AccountsHandler) UserRole(ctx *gin.Context) {
	bearerToken := strings.TrimPrefix(ctx.GetHeader("Authorization"), "Bearer ")
	claims := ParseJWT(bearerToken)

	gormDB, _ := initializer.GetGorm()
	var identityType string
	errGetIdentity := gormDB.Model(&models.IdentityAccess{}).Select([]string{"type"}).Where("user_id = ?", claims.Id).First(&identityType).Error
	if errGetIdentity != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errGetIdentity.Error(),
			"message": "user doesn't have identity!",
		})

		ctx.Abort()
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    identityType,
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

func (h *AccountsHandler) UserAccountInfo(ctx *gin.Context) {
	bearerToken := strings.TrimPrefix(ctx.GetHeader("Authorization"), "Bearer ")
	claims := ParseJWT(bearerToken)

	gormDB, _ := initializer.GetGorm()
	var userAccountData map[string]interface{}
	errGetUser := gormDB.Model(&models.User{}).Select([]string{"fullname", "email"}).Where("id = ?", claims.Id).First(&userAccountData).Error

	if errGetUser != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errGetUser.Error(),
			"message": "user account record may not found",
		})

		ctx.Abort()
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    userAccountData,
	})
}

/* multi-purpose */
func (h *AccountsHandler) MakeAccount(ctx *gin.Context) {
	var AccountProps struct {
		Fullname     string `json:"fullname" binding:"required"`
		Email        string `json:"email" binding:"required"`
		Password     string `json:"password" binding:"required"`
		IdentityType string `json:"identity_type" binding:"required"`
	}

	if errBindBody := ctx.ShouldBindJSON(&AccountProps); errBindBody != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   errBindBody.Error(),
			"message": "check your JSON fields",
		})

		ctx.Abort()
		return
	}

	gormDB, errGorm := initializer.GetGorm()
	if errGorm != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errGorm.Error(),
			"message": "failed getting GORM database instance",
		})

		ctx.Abort()
		return
	}

	roles := []map[string]interface{}{}
	getRoles := gormDB.Model(&models.Role{}).Select([]string{"id", "name"}).Find(&roles)
	if getRoles.RowsAffected == 0 {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "no available roles",
			"message": "roles table is empty, create role first before creating account",
		})

		ctx.Abort()
		return
	}

	ch_userID := make(chan string)
	go GenUuid(AccountProps.Fullname, ch_userID)
	ch_userPassword := make(chan string)
	go HashPassword(AccountProps.Password, ch_userPassword)

	rolesMap := map[string]uint{}
	for _, role := range roles {
		rolesMap[role["name"].(string)] = role["id"].(uint)
	}

	if rolesMap[AccountProps.IdentityType] == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid identity_type",
			"message": "identity_type value should be the one of ['basic/candidate', 'basic/employer', 'basic/university', 'sdkdev/administrator']",
		})

		ctx.Abort()
		return
	}

	m_user := models.User{
		Id:        <-ch_userID,
		Fullname:  AccountProps.Fullname,
		Email:     AccountProps.Email,
		Password:  <-ch_userPassword,
		CreatedAt: time.Now(),
	}
	errMakeAccount := gormDB.Transaction(func(tx *gorm.DB) error {
		if errCreateUser := tx.Create(&m_user).Error; errCreateUser != nil {
			return errCreateUser
		}
		m_identity := models.IdentityAccess{
			UserId: m_user.Id,
			RoleId: rolesMap[AccountProps.IdentityType],
			Type:   strings.Split(AccountProps.IdentityType, "/")[1],
		}
		if errCreateIdentity := tx.Create(&m_identity).Error; errCreateIdentity != nil {
			return errCreateIdentity
		}
		return nil
	})

	if errMakeAccount != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errMakeAccount.Error(),
			"message": "failed creating account with provided data",
		})

		ctx.Abort()
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    "account created successfully",
	})
}

func (h *AccountsHandler) MakeRandomAccounts(ctx *gin.Context) {
	var RandomAccountOptions struct {
		Count        int    `json:"count" binding:"required"`
		IdentityType string `json:"identity_type" binding:"required"`
	}

	if errBind := ctx.ShouldBindJSON(&RandomAccountOptions); errBind != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   errBind.Error(),
			"message": "please double check your JSON fields",
		})

		ctx.Abort()
		return
	}

	IDGenerator := func(fullname string) string {
		namespace := uuid.Must(uuid.NewRandom())
		data := []byte(fullname)
		sha1ID := uuid.NewSHA1(namespace, data)
		return sha1ID.String()
	}

	PasswordGenerator := func(formattedName string) string {
		hashed, errHash := bcrypt.GenerateFromPassword([]byte(formattedName), bcrypt.DefaultCost)
		if errHash != nil {
			log.Println("fail hashing password")
			panic(errHash)
		}

		return string(hashed)
	}

	gormDB, errGorm := initializer.GetGorm()
	if errGorm != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errGorm.Error(),
			"message": "fail getting GORM connection instance",
		})

		ctx.Abort()
		return
	}

	switch RandomAccountOptions.IdentityType {
	case "basic/employer":
		ctx.JSON(http.StatusCreated, gin.H{
			"success": true,
			"data":    "will be available soon",
		})

		ctx.Abort()
		return

	case "basic/candidate":
		candidateFaker := gofakeit.NewFaker(source.NewCrypto(), true)

		m_users := make([]models.User, RandomAccountOptions.Count)
		m_identities := make([]models.IdentityAccess, RandomAccountOptions.Count)
		m_candidates := make([]models.Candidate, RandomAccountOptions.Count)
		for i := 0; i < RandomAccountOptions.Count; i++ {
			fullname := fmt.Sprintf("%s %s", candidateFaker.FirstName(), candidateFaker.LastName())
			m_users[i] = models.User{
				Id:        IDGenerator(fullname),
				Fullname:  fullname,
				Email:     candidateFaker.Email(),
				Password:  PasswordGenerator(strings.ToLower(strings.ReplaceAll(fullname, " ", "."))),
				CreatedAt: time.Now(),
			}

			m_candidates[i] = models.Candidate{
				Id:          IDGenerator(fullname),
				Expertise:   candidateFaker.JobTitle(),
				AboutMe:     candidateFaker.Sentence(32),
				DateOfBirth: gofakeit.PastDate(),
				CreatedAt:   time.Now(),
				User:        &m_users[i],
			}

			m_identities[i] = models.IdentityAccess{
				UserId: m_users[i].Id,
				RoleId: 1,
				Type:   "candidate",
			}
		}

		errCreateCandidates := gormDB.Transaction(func(tx *gorm.DB) error {
			errStoreCandidates := tx.CreateInBatches(&m_candidates, 50).Error
			if errStoreCandidates != nil {
				return errStoreCandidates
			}

			errStoreIdentities := tx.CreateInBatches(&m_identities, 50).Error
			if errStoreIdentities != nil {
				return errStoreIdentities
			}

			return nil
		})
		if errCreateCandidates != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   errCreateCandidates.Error(),
				"message": "fail creating random candidates",
			})

			ctx.Abort()
			return
		}

		ctx.JSON(http.StatusCreated, gin.H{
			"success": true,
			"data":    fmt.Sprintf("%d candidates created", len(m_candidates)),
		})

		ctx.Abort()
		return

	default:
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid identity_type",
			"message": "'identity_type' should be one of these option [basic/candidate, basic/employer, basic/university, sdkdev/administrator]",
		})

		ctx.Abort()
		return
	}
}
