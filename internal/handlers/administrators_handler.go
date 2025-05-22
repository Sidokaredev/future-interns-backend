package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	initializer "future-interns-backend/init"
	"future-interns-backend/internal/models"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/brianvoe/gofakeit/v7/source"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AdministratorHandlers struct {
}

type CreateEmployerUserJSON struct {
	Fullname string `json:"fullname" binding:"required"`
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (a *AdministratorHandlers) CreateEmployerUser(ctx *gin.Context) {
	/*
			  get key "permissions"
			  check the permission name before executing a job
		    if doesn't have required permissions, just throw an fail response
	*/
	permissions, _ := ctx.Get("permissions")
	if !permissions.((map[string]bool))["users.employer.create"] {
		ctx.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Lacks the [users.employer.create] permission",
			"message": "you are not allowed to access this resource",
		})

		ctx.Abort()
		return
	}

	user_data := CreateEmployerUserJSON{}
	if errBind := ctx.ShouldBindJSON(&user_data); errBind != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   errBind.Error(),
			"message": "double check your JSON fields, kid",
		})

		ctx.Abort()
		return
	}

	ch_hash_password := make(chan string, 1)
	ch_uuid := make(chan string, 1)
	go GenUuid(user_data.Fullname, ch_uuid)
	go HashPassword(user_data.Password, ch_hash_password)

	const EmployerRoleId = 2

	gormDB, _ := initializer.GetGorm()
	m_user := models.User{
		Id:       <-ch_uuid,
		Fullname: user_data.Fullname,
		Email:    user_data.Email,
		Password: <-ch_hash_password,
	}

	errCreateEmployerUser := gormDB.Transaction(func(tx *gorm.DB) error {
		errCreateUser := tx.Create(&m_user).Error
		if errCreateUser != nil {
			return errCreateUser
		}

		errAssignIdentity := tx.Create(&models.IdentityAccess{
			UserId: m_user.Id,
			RoleId: EmployerRoleId,
			Type:   "employer",
		}).Error
		if errAssignIdentity != nil {
			return errAssignIdentity
		}

		return nil
	})

	if errCreateEmployerUser != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errCreateEmployerUser.Error(),
			"message": "failed creating new employer user",
		})

		ctx.Abort()
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    "Employer user created successfully",
	})
}
func (a *AdministratorHandlers) UpdateEmployerUser(ctx *gin.Context) {

}

func (a *AdministratorHandlers) GetEmployerUserById(ctx *gin.Context) {

}

func (a *AdministratorHandlers) DeleteEmployerUserById(ctx *gin.Context) {

}

func (a *AdministratorHandlers) ListEmployerUsers(ctx *gin.Context) {

}

/* Skill */
func (a *AdministratorHandlers) CreateSkills(ctx *gin.Context) {
	var skillForm struct {
		Name string `form:"name" binding:"required"`
	}
	if errBind := ctx.ShouldBind(&skillForm); errBind != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   errBind.Error(),
			"message": "name for skill is required",
		})

		ctx.Abort()
		return
	}

	skillIcon, errSkillIcon := ctx.FormFile("skill_icon_image")
	if errSkillIcon != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   errSkillIcon.Error(),
			"message": "skill_icon_image is required",
		})

		ctx.Abort()
		return
	}

	m_image, errImage := ImageData(skillIcon)
	if errImage != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errImage.Error(),
			"message": "there was an error with your file image",
		})

		ctx.Abort()
		return
	}

	gormDB, _ := initializer.GetGorm()
	errStoreSkill := gormDB.Transaction(func(tx *gorm.DB) error {
		errCreateImage := tx.Create(&m_image).Error
		if errCreateImage != nil {
			return errCreateImage
		}

		m_skill := models.Skill{
			Name:             skillForm.Name,
			SkillIconImageId: m_image.ID,
			CreatedAt:        time.Now(),
			UpdatedAt:        nil,
		}
		errCreateSkill := tx.Create(&m_skill).Error
		if errCreateSkill != nil {
			return errCreateSkill
		}

		return nil
	})

	if errStoreSkill != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errStoreSkill.Error(),
			"message": "error database operation",
		})

		ctx.Abort()
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    "skill stored successfully",
	})
}

func (a *AdministratorHandlers) DeleteSkills(ctx *gin.Context) {
	skillId := ctx.Param("id")
	if _, errParse := strconv.Atoi(skillId); errParse != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   errParse.Error(),
			"message": "skill id must be a valid number",
		})

		ctx.Abort()
		return
	}

	gormDB, _ := initializer.GetGorm()
	errDeleteSkill := gormDB.Transaction(func(tx *gorm.DB) error {
		var imageID int
		errGetImageID := tx.Model(&models.Skill{}).
			Select("skill_icon_image_id").
			Where("id = ?", skillId).
			First(&imageID).Error
		if errGetImageID != nil {
			return errGetImageID
		}

		log.Println("image id \t:", imageID)

		deleteSkillRow := tx.Delete(&models.Skill{}, skillId).RowsAffected
		if deleteSkillRow == 0 {
			return fmt.Errorf("failed deleting skill with id %v", skillId)
		}

		deleteImageRow := tx.Delete(&models.Image{}, imageID).RowsAffected
		if deleteImageRow == 0 {
			return fmt.Errorf("failed deleting image with id %v", imageID)
		}

		return nil
	})

	if errDeleteSkill != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errDeleteSkill.Error(),
			"message": "error database operation",
		})

		ctx.Abort()
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    fmt.Sprintf("skill with id %v deleted successfully", skillId),
	})
}

func (a *AdministratorHandlers) CreateSocial(ctx *gin.Context) {
	var socialForm struct {
		Name string `form:"name" binding:"required"`
	}
	if errBind := ctx.ShouldBind(&socialForm); errBind != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   errBind.Error(),
			"message": "double check your form-data fields, kids",
		})

		ctx.Abort()
		return
	}

	socialIcon, errSocialIcon := ctx.FormFile("social_icon_image")
	if errSocialIcon != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   errSocialIcon.Error(),
			"message": "you have to provide a image for social icon [social_icon_image]",
		})

		ctx.Abort()
		return
	}

	m_image, errImage := ImageData(socialIcon)
	if errImage != nil {
		ctx.JSON(http.StatusUnprocessableEntity, gin.H{
			"success": false,
			"error":   errImage.Error(),
			"message": "your provided file is corupted or invalid",
		})

		ctx.Abort()
		return
	}

	gormDB, _ := initializer.GetGorm()
	errCreateSocial := gormDB.Transaction(func(tx *gorm.DB) error {
		errStoreImage := tx.Create(&m_image).Error
		if errStoreImage != nil {
			return errStoreImage
		}

		m_social := models.Social{
			Name:        socialForm.Name,
			IconImageId: int(m_image.ID),
			CreatedAt:   time.Now(),
		}
		errStoreSocial := tx.Create(&m_social).Error
		if errStoreSocial != nil {
			return errStoreSocial
		}
		return nil
	})

	if errCreateSocial != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errCreateSocial.Error(),
			"message": "database operation failed, try again later",
		})

		ctx.Abort()
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    fmt.Sprintf("%s social created successfully!", socialForm.Name),
	})
}

/* Generates Vacancy */
func (a *AdministratorHandlers) GenerateVacancies(ctx *gin.Context) {
	var VacanciesOption struct {
		TotalEmployer             int `json:"total_employer" binding:"required"`
		TotalVacanciesPerEmployer int `json:"total_vacancies_per_employer" binding:"required"`
	}

	if errBind := ctx.ShouldBindJSON(&VacanciesOption); errBind != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   errBind.Error(),
			"message": "check your json fields and make sure the 'total_employer' and 'total_vacancies_per_employer' included",
		})

		ctx.Abort()
		return
	}

	ch_users := make(chan []models.User)
	ch_employers := make(chan []models.Employer)
	ch_vacancies := make(chan []models.Vacancy)

	go func(users chan<- []models.User) {
		userFaker := gofakeit.NewFaker(source.NewCrypto(), true)
		usersData := []models.User{}
		for i := 0; i < VacanciesOption.TotalEmployer; i++ {
			fullname := userFaker.FirstName() + userFaker.LastName()
			email := userFaker.Email()

			namespace := uuid.Must(uuid.NewRandom())
			data := []byte(fullname)
			sha1ID := uuid.NewSHA1(namespace, data)
			ID := sha1ID.String()
			hashedPassword, errHash := bcrypt.GenerateFromPassword([]byte(strings.Split(email, "@")[0]), bcrypt.DefaultCost)
			if errHash != nil {
				panic(errHash)
			}
			usersData = append(usersData, models.User{
				Id:        ID,
				Fullname:  fullname,
				Email:     email,
				Password:  string(hashedPassword),
				CreatedAt: time.Now(),
			})
		}

		users <- usersData
		close(users)
	}(ch_users)

	go func(employers chan<- []models.Employer) {
		employerFaker := gofakeit.NewFaker(source.NewCrypto(), true)
		employersData := []models.Employer{}
		for i := 0; i < VacanciesOption.TotalEmployer; i++ {
			companyName := employerFaker.Company()
			namespace := uuid.Must(uuid.NewRandom())
			data := []byte(companyName)
			sha1ID := uuid.NewSHA1(namespace, data)
			ID := sha1ID.String()
			totalOfEmployee := employerFaker.RandomString([]string{"1 - 10",
				"11 - 50",
				"51 - 200",
				"201 - 500",
				"501 - 1000",
				"1001 - 5000",
				"5001+",
			})

			employersData = append(employersData, models.Employer{
				Id:              ID,
				Name:            companyName,
				LegalName:       "PT. " + companyName,
				Location:        employerFaker.City(),
				Founded:         uint(employerFaker.Int16()),
				Founder:         employerFaker.PetName(),
				TotalOfEmployee: totalOfEmployee,
				Description:     employerFaker.Sentence(15),
				Website:         employerFaker.URL(),
				CreatedAt:       time.Now(),
			})
		}

		employers <- employersData
		close(employers)
	}(ch_employers)

	go func(vacancies chan<- []models.Vacancy) {
		vacanciesFaker := gofakeit.NewFaker(source.NewCrypto(), true)
		vacanciesData := []models.Vacancy{}
		for i := 0; i < VacanciesOption.TotalEmployer*VacanciesOption.TotalVacanciesPerEmployer; i++ {
			namespace := uuid.Must(uuid.NewRandom())
			data := []byte("vacancy")
			sha1ID := uuid.NewSHA1(namespace, data)
			ID := sha1ID.String()

			vacanciesData = append(vacanciesData, models.Vacancy{
				Id:             ID,
				Position:       vacanciesFaker.JobTitle(),
				Description:    vacanciesFaker.Sentence(50),
				Qualification:  vacanciesFaker.Sentence(50),
				Responsibility: vacanciesFaker.Sentence(50),
				LineIndustry: vacanciesFaker.RandomString([]string{
					"IT and Technology",
					"Finance",
					"Construction and Real Estate",
					"Insurance",
					"Retail and E-commerce",
					"Entertainment and Media",
					"Transportation and Logistics",
					"Telecommunications",
					"Education",
					"Legal Services"}),
				EmployeeType: vacanciesFaker.RandomString([]string{
					"Full-time",
					"Part-time",
					"Contract",
					"Freelance",
					"Internship",
					"Temporary",
					"Volunteer",
					"Remote",
					"On-call",
					"Seasonal"}),
				MinExperience: vacanciesFaker.RandomString([]string{
					"No experience required",
					"Less than 1 year",
					"1-2 years",
					"3-5 years",
					"6-10 years",
					"More than 10 years",
				}),
				Salary: uint(vacanciesFaker.IntRange(1000000, 100000000)),
				WorkArrangement: vacanciesFaker.RandomString([]string{
					"On-site",
					"Remote",
					"Hybrid (On-site & Remote)",
					"Flexible Hours",
					"Shift-based",
					"Compressed Workweek",
					"Freelance/Project-based",
					"Rotational",
				}),
				SLA:        168,
				IsInactive: false,
				CreatedAt:  time.Now(),
			})
		}

		vacancies <- vacanciesData
		close(vacancies)
	}(ch_vacancies)

	gormDB, errGorm := initializer.GetGorm()
	if errGorm != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errGorm.Error(),
			"message": "failed connection with gorm.DB instance",
		})

		ctx.Abort()
		return
	}

	errGenVacancies := gormDB.Transaction(func(tx *gorm.DB) error {
		m_users := <-ch_users
		errStoreUsers := tx.CreateInBatches(&m_users, 10).Error
		if errStoreUsers != nil {
			return errStoreUsers
		}

		m_identityAccesses := []models.IdentityAccess{}
		m_employers := <-ch_employers
		for index, user := range m_users {
			m_identityAccesses = append(m_identityAccesses, models.IdentityAccess{
				UserId: user.Id,
				RoleId: 2,
				Type:   "employer",
			})
			m_employers[index].UserId = user.Id
		}

		errStoreIdentity := tx.CreateInBatches(&m_identityAccesses, 10).Error
		if errStoreIdentity != nil {
			return errStoreIdentity
		}
		errStoreEmployers := tx.CreateInBatches(&m_employers, 10).Error
		if errStoreEmployers != nil {
			return errStoreEmployers
		}

		m_vacancies := <-ch_vacancies
		for index, employer := range m_employers {
			startIndex := index * VacanciesOption.TotalVacanciesPerEmployer
			endIndex := startIndex + VacanciesOption.TotalVacanciesPerEmployer

			for i := startIndex; i < endIndex; i++ {
				m_vacancies[i].EmployerId = employer.Id
			}
		}

		errStoreVacancies := tx.CreateInBatches(&m_vacancies, VacanciesOption.TotalVacanciesPerEmployer).Error
		if errStoreVacancies != nil {
			return errStoreVacancies
		}

		return nil
	})

	if errGenVacancies != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errGenVacancies.Error(),
			"message": fmt.Sprintf("failed generating %v employer within %v vacancies on every employer", VacanciesOption.TotalEmployer, VacanciesOption.TotalVacanciesPerEmployer),
		})

		ctx.Abort()
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    fmt.Sprintf("%v employer within %v vacancies on every employer generated successfully", VacanciesOption.TotalEmployer, VacanciesOption.TotalVacanciesPerEmployer),
	})
}

func (a *AdministratorHandlers) DeleteGeneratedVacancies(ctx *gin.Context) {
	slaNumber, errConv := strconv.Atoi(ctx.Param("sla"))
	if errConv != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   errConv.Error(),
			"message": "param 'sla' must be a valid number",
		})

		ctx.Abort()
		return
	}

	gormDB, errGorm := initializer.GetGorm()
	if errGorm != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errGorm.Error(),
			"message": "failed establish connection with GORM instance",
		})

		ctx.Abort()
		return
	}

	var deletionStatus string
	errDeleteGeneratedVacancies := gormDB.Transaction(func(tx *gorm.DB) error {
		delVacancies := tx.Unscoped().Where("sla = ?", slaNumber).Delete(&models.Vacancy{})
		if delVacancies.RowsAffected == 0 {
			deletionStatus = "delete vacancies: no data deleted"
			return fmt.Errorf("%v rows affected, no vacancies were deleted", delVacancies.RowsAffected)
		}

		deletionStatus = fmt.Sprintf("%v vacancies", delVacancies.RowsAffected)

		delEmployer := tx.Unscoped().Where("background_profile_image_id IS NULL AND profile_image_id IS NULL").Delete(&models.Employer{})
		if delEmployer.RowsAffected == 0 {
			deletionStatus = "delete employer: no data deleted"
			return fmt.Errorf("%v rows affected, no employers were deleted", delEmployer.RowsAffected)
		}

		deletionStatus += fmt.Sprintf(", %v employers", delEmployer.RowsAffected)

		subQueryDeleteIdentity := tx.Unscoped().Model(&models.IdentityAccess{}).
			Select([]string{"identity_accesses.user_id"}).
			Joins(`
				LEFT JOIN employers ON employers.user_id = identity_accesses.user_id
				LEFT JOIN candidates ON candidates.user_id = identity_accesses.user_id
		`).
			Where("employers.user_id IS NULL AND candidates.user_id IS NULL")
		delIdentityAccesses := tx.Unscoped().Where("user_id IN (?)", subQueryDeleteIdentity).Delete(&models.IdentityAccess{})
		if delIdentityAccesses.RowsAffected == 0 {
			deletionStatus = "delete identity accesses: no data deleted"
			return fmt.Errorf("%v rows affected, no identities were deleted", delIdentityAccesses.RowsAffected)
		}

		deletionStatus += fmt.Sprintf(", %v identities", delIdentityAccesses.RowsAffected)

		subQueryDeleteUsers := tx.Unscoped().Model(&models.User{}).
			Select([]string{"users.id"}).
			Joins(`
				LEFT JOIN employers ON employers.user_id = users.id
				LEFT JOIN candidates ON candidates.user_id = users.id
		`).
			Where("employers.user_id IS NULL AND candidates.user_id IS NULL")
		delUsers := tx.Unscoped().Where("id IN (?)", subQueryDeleteUsers).Delete(&models.User{})
		if delUsers.RowsAffected == 0 {
			deletionStatus = "delete user: no data deleted"
			return fmt.Errorf("%v rows affected, no users were deleted", delUsers.RowsAffected)
		}

		deletionStatus += fmt.Sprintf(" and %v users has been deleted", delUsers.RowsAffected)
		return nil
	})

	if errDeleteGeneratedVacancies != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": true,
			"error":   errDeleteGeneratedVacancies.Error(),
			"message": deletionStatus,
		})

		ctx.Abort()
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    deletionStatus,
	})
}

func (h *AdministratorHandlers) GenerateRawVacancies(ctx *gin.Context) {
	type Vacancies struct {
		ID              string    `json:"id"`
		Position        string    `json:"position"`
		Description     string    `json:"description"`
		Qualification   string    `json:"qualification"`
		Responsibility  string    `json:"responsibility"`
		LineIndustry    string    `json:"line_industry"`
		EmployeeType    string    `json:"employee_type"`
		MinExperience   string    `json:"min_experience"`
		Salary          int64     `json:"salary"`
		WorkArrangement string    `json:"work_arrangement"`
		SLA             int       `json:"sla"`
		IsInactive      int       `json:"is_inactive"`
		EmployerID      string    `json:"employer_id"`
		CreatedAt       time.Time `json:"created_at"`
	}

	var countQuery string
	var countExist bool

	countQuery, countExist = ctx.GetQuery("count")
	if !countExist {
		countQuery = "5000"
	}
	count, errConv := strconv.Atoi(countQuery)
	if errConv != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   errConv.Error(),
			"message": "'count' should be a valid number",
		})

		ctx.Abort()
		return
	}

	gormDB, errGorm := initializer.GetGorm()
	if errGorm != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errGorm.Error(),
			"message": "fail getting GORM instance",
		})

		ctx.Abort()
		return
	}

	employersID := []string{}
	getEmployersID := gormDB.Raw(`
		SELECT TOP 50
			id
		FROM
			employers
		ORDER BY
			NEWID()
	`).Scan(&employersID)
	if getEmployersID.Error != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   getEmployersID.Error,
			"message": "failed while getting random employers ID",
		})

		ctx.Abort()
		return
	}

	vacanciesFaker := gofakeit.NewFaker(source.NewCrypto(), true)
	IDGenerator := func() string {
		namespace := uuid.Must(uuid.NewRandom())
		data := []byte("vacancy")
		sha1ID := uuid.NewSHA1(namespace, data)
		return sha1ID.String()
	}
	GenerateJobDescription := func() string {
		randomSenteces := []string{
			vacanciesFaker.Sentence(10),
			vacanciesFaker.Sentence(14),
			vacanciesFaker.Sentence(18),
		}

		return strings.Join(randomSenteces, ". ")
	}
	GenerateListedSentence := func(listFor string) string {
		listedSentences := []string{}
		for i := 0; i < 3; i++ {
			if listFor == "qualifications" {
				listedSentences = append(listedSentences, "- "+vacanciesFaker.BuzzWord()+" "+vacanciesFaker.Sentence(10))
			}
			if listFor == "responsibilities" {
				listedSentences = append(listedSentences, "* "+vacanciesFaker.BuzzWord()+" "+vacanciesFaker.Sentence(10))
			}
		}

		return strings.Join(listedSentences, "\n")
	}

	rawVacancies := []Vacancies{}
	for _, ID := range employersID {
		for i := 0; i < (count / len(employersID)); i++ {
			rawVacancies = append(rawVacancies, Vacancies{
				ID:             IDGenerator(),
				Position:       vacanciesFaker.JobTitle(),
				Description:    GenerateJobDescription(),
				Qualification:  GenerateListedSentence("qualifications"),
				Responsibility: GenerateListedSentence("responsibilities"),
				LineIndustry: vacanciesFaker.RandomString([]string{
					"IT and Technology",
					"Finance",
					"Construction and Real Estate",
					"Insurance",
					"Retail and E-commerce",
					"Entertainment and Media",
					"Transportation and Logistics",
					"Telecommunications",
					"Education",
					"Legal Services"}),
				EmployeeType: vacanciesFaker.RandomString([]string{
					"Full-time",
					"Part-time",
					"Contract",
					"Freelance",
					"Internship",
					"Temporary",
					"Volunteer",
					"Remote",
					"On-call",
					"Seasonal"}),
				MinExperience: vacanciesFaker.RandomString([]string{
					"No experience required",
					"Less than 1 year",
					"1-2 years",
					"3-5 years",
					"6-10 years",
					"More than 10 years",
				}),
				Salary: int64(vacanciesFaker.IntRange(1000000, 100000000)),
				WorkArrangement: vacanciesFaker.RandomString([]string{
					"On-site",
					"Remote",
					"Hybrid (On-site & Remote)",
					"Flexible Hours",
					"Shift-based",
					"Compressed Workweek",
					"Freelance/Project-based",
					"Rotational",
				}),
				SLA:        168,
				IsInactive: vacanciesFaker.RandomInt([]int{1, 0}),
				EmployerID: ID,
				CreatedAt:  time.Now(),
			})
		}
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    rawVacancies,
	})
}

// path: /test
func (h *AdministratorHandlers) CreateNewTestSession(ctx *gin.Context) {
	var TestProps struct {
		Label string `json:"label" binding:"required"`
	}

	if errBind := ctx.ShouldBindJSON(&TestProps); errBind != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   errBind.Error(),
			"message": "please check your JSON fields",
		})

		ctx.Abort()
		return
	}

	gormDB, errGorm := initializer.GetGorm()
	if errGorm != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errGorm.Error(),
			"message": "fail getting GORM instance",
		})

		ctx.Abort()
		return
	}

	m_cache_session := models.CacheSession{
		Label:     TestProps.Label,
		CreatedAt: time.Now(),
	}
	errCreateTest := gormDB.Create(&m_cache_session).Error
	if errCreateTest != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errCreateTest.Error(),
			"message": "fail creating new test",
		})

		ctx.Abort()
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data": gin.H{
			"id":      m_cache_session.ID,
			"message": "new test has been created",
		},
	})
}

func (h *AdministratorHandlers) GetRequestLogsByPattern(ctx *gin.Context) {
	sessionIDQuery := ctx.Param("sessionID")
	sessionID, errConv := strconv.Atoi(sessionIDQuery)
	if errConv != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   errConv.Error(),
			"message": "'sessionID' must be a valid number",
		})

		ctx.Abort()
		return
	}

	patternQuery, isPatterExist := ctx.GetQuery("pattern")
	if !isPatterExist || patternQuery == "" {
		patternQuery = "cache-aside"
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

	cacheLogs := []struct {
		ID                  uint      `json:"id"`
		CacheHit            int       `json:"cache_hit"`
		CacheMiss           int       `json:"cache_miss"`
		ResponseTime        uint64    `json:"response_time"`
		MemoryUsage         float64   `json:"memory_usage"`
		CPUUsage            float64   `json:"cpu_usage"`
		ResourceUtilization float64   `json:"resource_utilization"`
		CacheType           string    `json:"cache_type"`
		CreatedAt           time.Time `json:"created_at"`
		CacheSessionID      int       `json:"cache_session_id"`
	}{}
	getLogs := gormDB.Model(&models.RequestLog{}).Where("cache_session_id = ? AND cache_type = ?", sessionID, patternQuery).Order("created_at ASC").Limit(250).Find(&cacheLogs)
	if getLogs.RowsAffected == 0 {
		log.Printf("cache logs with cache_session_id = %d and type = %s does not exist", sessionID, patternQuery)
	}

	cacheHitCount := 0
	cacheMissCount := 0
	labels := []string{}
	responseTimeLogs := []uint64{}
	resourceUtilizationLogs := []float64{}
	for logIndex, log_ := range cacheLogs {
		if log_.CacheHit == 1 {
			cacheHitCount += 1
		} else if log_.CacheMiss == 1 {
			cacheMissCount += 1
		}

		labels = append(labels, fmt.Sprintf("request:%d", (logIndex+1)))
		responseTimeLogs = append(responseTimeLogs, log_.ResponseTime)
		resourceUtilizationLogs = append(resourceUtilizationLogs, log_.ResourceUtilization)
	}

	cacheStatus := map[string]interface{}{
		"datasets": []map[string]interface{}{
			{
				"type": "bar",
				"data": []map[string]interface{}{
					{"x": cacheHitCount, "y": "Cache Hit"},
					{"x": cacheMissCount, "y": "Cache Miss"},
				},
				"backgroundColor": []string{"#90caf9", "#ef9a9a"},
				"barThickness":    50,
			},
		},
	}

	resourceUtils := map[string]interface{}{
		"labels": labels,
		"datasets": []map[string]interface{}{
			{
				"label":       "Response Time",
				"data":        responseTimeLogs,
				"yAxisID":     "y",
				"borderColor": "#a5d6a7",
				"borderWidth": 2,
			},
			{
				"label":       "Resource Utilization",
				"data":        resourceUtilizationLogs,
				"yAxisID":     "y1",
				"borderColor": "#ffcc80",
				"borderWidth": 2,
			},
		},
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"chart": gin.H{
				"cache_status":   cacheStatus,
				"resource_utils": resourceUtils,
			},
			"logs": cacheLogs,
		},
	})
}

func (h *AdministratorHandlers) GetFullRequestLogsBySession(ctx *gin.Context) {
	sessionID, errConv := strconv.Atoi(ctx.Param("sessionID"))
	if errConv != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   errConv.Error(),
			"message": "'sessionID' must be a valid number",
		})
		return
	}

	gormDB, errGorm := initializer.GetGorm()
	if errGorm != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errGorm.Error(),
			"message": "fail getting GORM instance connection",
		})
		return
	}

	type RequestLogProps struct {
		ID                  uint      `json:"id" gorm:"column:id"`
		CacheHit            int       `json:"cache_hit" gorm:"column:cache_hit"`
		CacheMiss           int       `json:"cache_miss" gorm:"column:cache_miss"`
		ResponseTime        uint64    `json:"response_time" gorm:"column:response_time"`
		MemoryUsage         float64   `json:"memory_usage" gorm:"column:memory_usage"`
		CPUUsage            float64   `json:"cpu_usage" gorm:"column:cpu_usage"`
		ResourceUtilization float64   `json:"resource_utilization" gorm:"column:resource_utilization"`
		CacheType           string    `json:"cache_type" gorm:"column:cache_type"`
		CreatedAt           time.Time `json:"created_at" gorm:"column:created_at"`
		CacheSessionID      int       `json:"cache_session_id" gorm:"column:cache_session_id"`
	}

	var RequestLogs []RequestLogProps
	cacheTypes := []string{"no-cache", "no-cache", "cache-aside", "read-through", "write-through", "write-behind"}
	getRequestLogs := gormDB.Model(&models.RequestLog{}).Where("cache_session_id = ? AND cache_type IN (?)", sessionID, cacheTypes).Order("created_at ASC").Find(&RequestLogs)
	if getRequestLogs.Error != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   getRequestLogs.Error.Error(),
			"message": "there was an error with sql query",
		})
		return
	}
	if getRequestLogs.RowsAffected == 0 {
		RequestLogs = []RequestLogProps{}
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    RequestLogs,
	})
}

func (h *AdministratorHandlers) GetRequestLogsNoCache(ctx *gin.Context) {

}

func (h *AdministratorHandlers) CheckSessionTestStatus(ctx *gin.Context) {
	sessionIDParam := ctx.Param("sessionID")
	sessionID, errConv := strconv.Atoi(sessionIDParam)
	if errConv != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   errConv.Error(),
			"message": "'sessionID' param must be a valid number",
		})

		ctx.Abort()
		return
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

	type CacheSessionProps struct {
		ID    uint   `json:"id" gorm:"column:id"`
		Label string `json:"label" gorm:"column:label"`
		// Status    *string   `json:"status" gorm:"column:status"`
		CreatedAt time.Time `json:"created_at" gorm:"column:created_at"`
	}
	type SessionTestStatusProps struct {
		Detail       CacheSessionProps `json:"detail"`
		NoCache      int64             `json:"no_cache"`
		CacheAside   int64             `json:"cache_aside"`
		ReadThrough  int64             `json:"read_through"`
		WriteThrough int64             `json:"write_through"`
		WriteBehind  int64             `json:"write_behind"`
	}

	sessionTestStatus := SessionTestStatusProps{}
	errGetSessionDetail := gormDB.Transaction(func(tx *gorm.DB) error {
		cacheSessionInfo := CacheSessionProps{}
		errGetCacheSession := tx.Model(&models.CacheSession{}).Where("id = ?", sessionID).First(&cacheSessionInfo).Error
		if errGetCacheSession != nil {
			return errGetCacheSession
		}
		sessionTestStatus.Detail = cacheSessionInfo

		var noCacheLogs int64
		noCacheWriteInfo := tx.Model(&models.RequestLog{}).Where("cache_session_id = ? AND cache_type = ?", sessionID, "no-cache").Count(&noCacheLogs)
		if noCacheWriteInfo.Error != nil {
			return noCacheWriteInfo.Error
		}
		sessionTestStatus.NoCache = noCacheLogs

		var cacheAsideLogs int64
		cacheAsideInfo := tx.Model(&models.RequestLog{}).Where("cache_session_id = ? AND cache_type = ?", sessionID, "cache-aside").Count(&cacheAsideLogs)
		if cacheAsideInfo.Error != nil {
			return cacheAsideInfo.Error
		}
		sessionTestStatus.CacheAside = cacheAsideLogs

		var readThroughLogs int64
		readThroughInfo := tx.Model(&models.RequestLog{}).Where("cache_session_id = ? AND cache_type = ?", sessionID, "read-through").Count(&readThroughLogs)
		if readThroughInfo.Error != nil {
			return readThroughInfo.Error
		}
		sessionTestStatus.ReadThrough = readThroughLogs

		var writeThroughLogs int64
		writeThroughInfo := tx.Model(&models.RequestLog{}).Where("cache_session_id = ? AND cache_type = ?", sessionID, "write-through").Count(&writeThroughLogs)
		if writeThroughInfo.Error != nil {
			return writeThroughInfo.Error
		}
		sessionTestStatus.WriteThrough = writeThroughLogs

		var writeBehindLogs int64
		writeBehindInfo := tx.Model(&models.RequestLog{}).Where("cache_session_id = ? AND cache_type = ?", sessionID, "write-behind").Count(&writeBehindLogs)
		if writeBehindInfo.Error != nil {
			return writeBehindInfo.Error
		}
		sessionTestStatus.WriteBehind = writeBehindLogs

		return nil
	})

	if errGetSessionDetail != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errGetSessionDetail.Error(),
			"message": "data may not be available",
		})

		ctx.Abort()
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    sessionTestStatus,
	})
}

func (h *AdministratorHandlers) GenerateRandomToken(ctx *gin.Context) {
	countQuery, isCountExist := ctx.GetQuery("count")
	if !isCountExist {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "'count' search query missing",
			"message": "'count' is required, please include that query",
		})

		ctx.Abort()
		return
	}

	count, errConv := strconv.Atoi(countQuery)
	if errConv != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   errConv.Error(),
			"message": "'count' query value should be a valid number",
		})

		ctx.Abort()
		return
	}

	gormDB, errGorm := initializer.GetGorm()
	if errGorm != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errGorm.Error(),
			"message": "fail while getting GORM connection instance",
		})

		ctx.Abort()
		return
	}

	sql := `
		SELECT TOP (?)
			candidates.id,
			users.fullname,
			users.email
		FROM
			candidates
			INNER JOIN users ON users.id = candidates.user_id
		ORDER BY
			NEWID()
	`

	var candidates []struct {
		ID       string
		Fullname string
		Email    string
	}
	errGetRandomCandidateID := gormDB.Raw(sql, count).Scan(&candidates).Error
	if errGetRandomCandidateID != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errGetRandomCandidateID.Error(),
			"message": "there was something wrong with SQL Query",
		})

		ctx.Abort()
		return
	}

	type ResponseAuth struct {
		Success bool `json:"success"`
		Data    struct {
			UserID      string `json:"user_id"`
			AccessToken string `json:"access_token"`
			Role        struct {
				Description string `json:"description"`
				Name        string `json:"name"`
				Type        string `json:"type"`
			} `json:"role"`
		} `json:"data"`
	}

	var tokens []string
	for _, candidate := range candidates {
		requestBody := map[string]string{
			"email":    candidate.Email,
			"password": strings.ToLower(strings.ReplaceAll(candidate.Fullname, " ", ".")),
		}
		JSONBody, errMarshall := json.Marshal(requestBody)
		if errMarshall != nil {
			log.Println("MARSHAL BODY: ", errMarshall.Error())
			continue
		}

		response, errResp := http.Post("http://localhost:3000/api/v1/accounts/auth", "application/json", bytes.NewBuffer(JSONBody))
		if errResp != nil {
			log.Println("HTTP POST: ", errResp.Error())
			continue
		}
		defer response.Body.Close()

		if response.StatusCode != 200 {
			var failResponse struct {
				Success bool   `json:"success"`
				Error   string `json:"error"`
				Message string `json:"message"`
			}

			if errDecode := json.NewDecoder(response.Body).Decode(&failResponse); errDecode != nil {
				log.Println("DECODE BODY: ", errDecode.Error())
				continue
			}
			log.Println("FAIL REQUEST: ", failResponse)

			continue
		}
		var auth ResponseAuth
		if errDecode := json.NewDecoder(response.Body).Decode(&auth); errDecode != nil {
			log.Println("DECODE BODY: ", errDecode.Error())
			continue
		}

		tokens = append(tokens, auth.Data.AccessToken)
	}

	if len(tokens) == 0 {
		ctx.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "empty tokens returned",
			"message": "no tokens were generated",
		})

		ctx.Abort()
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    tokens,
	})
}

/* CACHE SESSION INFORMATION */
func (h *AdministratorHandlers) GetCacheSessionByID(ctx *gin.Context) {
	sessionIDParam := ctx.Param("sessionID")
	sessionID, errConv := strconv.Atoi(sessionIDParam)
	if errConv != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   errConv.Error(),
			"message": "'sessionID' param must be a valid number",
		})

		ctx.Abort()
		return
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

	cacheSessionDetail := map[string]interface{}{}
	errGetSessionDetail := gormDB.Transaction(func(tx *gorm.DB) error {
		cacheSession := map[string]interface{}{}
		errGetCacheSession := tx.Model(&models.CacheSession{}).Where("id = ?", sessionID).First(&cacheSession).Error
		if errGetCacheSession != nil {
			return errGetCacheSession
		}
		cacheSessionDetail["detail"] = cacheSession

		noCacheTestStatus := map[string]interface{}{}
		var noCacheWriteLogs int64
		noCacheWriteInfo := tx.Model(&models.RequestLog{}).Where("cache_session_id = ? AND cache_type = ?", sessionID, "no-cache-write").Count(&noCacheWriteLogs)
		if noCacheWriteInfo.Error != nil {
			return noCacheWriteInfo.Error
		}
		if noCacheWriteLogs < 200 {
			noCacheTestStatus["write"] = "not started"
		} else {
			noCacheTestStatus["write"] = "completed"
		}

		var noCacheReadLogs int64
		noCacheReadInfo := tx.Model(&models.RequestLog{}).Where("cache_session_id = ? AND cache_type = ?", sessionID, "no-cache-read").Count(&noCacheReadLogs)
		if noCacheReadInfo.Error != nil {
			return noCacheReadInfo.Error
		}
		if noCacheReadLogs < 200 {
			noCacheTestStatus["read"] = "not started"
		} else {
			noCacheTestStatus["read"] = "completed"
		}

		cacheSessionDetail["no_cache"] = noCacheTestStatus

		var cacheAsideLogs int64
		cacheAsideInfo := tx.Model(&models.RequestLog{}).Where("cache_session_id = ? AND cache_type = ?", sessionID, "cache-aside").Count(&cacheAsideLogs)
		if cacheAsideInfo.Error != nil {
			return cacheAsideInfo.Error
		}
		if cacheAsideLogs < 200 {
			cacheSessionDetail["cache_aside"] = "not started"
		} else {
			cacheSessionDetail["cache_aside"] = "completed"
		}

		var readThroughLogs int64
		readThroughInfo := tx.Model(&models.RequestLog{}).Where("cache_session_id = ? AND cache_type = ?", sessionID, "read-through").Count(&readThroughLogs)
		if readThroughInfo.Error != nil {
			return readThroughInfo.Error
		}
		if readThroughLogs < 200 {
			cacheSessionDetail["read_through"] = "not started"
		} else {
			cacheSessionDetail["read_through"] = "completed"
		}

		var writeThroughLogs int64
		writeThroughInfo := tx.Model(&models.RequestLog{}).Where("cache_session_id = ? AND cache_type = ?", sessionID, "write-through").Count(&writeThroughLogs)
		if writeThroughInfo.Error != nil {
			return writeThroughInfo.Error
		}
		if writeThroughLogs < 200 {
			cacheSessionDetail["write_through"] = "not started"
		} else {
			cacheSessionDetail["write_through"] = "completed"
		}

		var writeBehindLogs int64
		writeBehindInfo := tx.Model(&models.RequestLog{}).Where("cache_session_id = ? AND cache_type = ?", sessionID, "write-behind").Count(&writeBehindLogs)
		if writeBehindInfo.Error != nil {
			return writeBehindInfo.Error
		}
		if writeBehindLogs < 200 {
			cacheSessionDetail["write_behind"] = "not started"
		} else {
			cacheSessionDetail["write_behind"] = "completed"
		}

		return nil
	})

	if errGetSessionDetail != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errGetSessionDetail.Error(),
			"message": "data may not be available",
		})

		ctx.Abort()
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    cacheSessionDetail,
	})
}

func (h *AdministratorHandlers) DeleteLogs(ctx *gin.Context) {
	// delete if request not counted into 200
}

/* NO CACHE */
func (h *AdministratorHandlers) GetNoCacheLogs(ctx *gin.Context) {
	sessionIDQuery := ctx.Param("sessionID")
	sessionID, errConv := strconv.Atoi(sessionIDQuery)
	if errConv != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   errConv.Error(),
			"message": "'sessionID' must be a valid number",
		})

		ctx.Abort()
		return
	}
	// typeQuery, isTypeExist := ctx.GetQuery("type")
	// if !isTypeExist {
	// 	ctx.JSON(http.StatusBadRequest, gin.H{
	// 		"success": false,
	// 		"error":   "'type' query param is required",
	// 		"message": "please specify the logs type by adding 'type' query ['write' OR 'read']",
	// 	})

	// 	ctx.Abort()
	// 	return
	// }

	// if typeQuery == "write" {
	// 	typeQuery = "no-cache-write"
	// } else {
	// 	typeQuery = "no-cache-read"
	// }

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

	cacheLogs := []struct {
		ID                  uint      `json:"id"`
		CacheHit            int       `json:"cache_hit"`
		CacheMiss           int       `json:"cache_miss"`
		ResponseTime        uint64    `json:"response_time"`
		MemoryUsage         float64   `json:"memory_usage"`
		CPUUsage            float64   `json:"cpu_usage"`
		ResourceUtilization float64   `json:"resource_utilization"`
		CacheType           string    `json:"cache_type"`
		CreatedAt           time.Time `json:"created_at"`
		CacheSessionID      int       `json:"cache_session_id"`
	}{}
	getLogs := gormDB.Model(&models.RequestLog{}).Where("cache_session_id = ? AND cache_type = ?", sessionID, "no-cache").Order("created_at ASC").Limit(250).Find(&cacheLogs)
	if getLogs.RowsAffected == 0 {
		log.Printf("cache logs with cache_session_id = %d and type = %s does not exist", sessionID, "no-cache")
	}

	type Dataset struct {
		Label           string `json:"label"`
		Data            [4]int `json:"data"`
		BackgroundColor string `json:"backgroundColor"`
	}
	dataset1 := Dataset{
		Label:           "< 5%",
		Data:            [4]int{0, 0, 0, 0},
		BackgroundColor: "rgba(75, 192, 192, 0.2)",
	}
	dataset2 := Dataset{
		Label:           "5% - 10%",
		Data:            [4]int{0, 0, 0, 0},
		BackgroundColor: "rgba(153, 102, 255, 0.2)",
	}
	dataset3 := Dataset{
		Label:           "> 10%",
		Data:            [4]int{0, 0, 0, 0},
		BackgroundColor: "rgba(255, 159, 64, 0.2)",
	}

	for _, log_ := range cacheLogs {
		responseTimeCategories := map[string]struct {
			Min, Max uint64
		}{
			"< 500ms":         {0, 500},
			"500ms - 1000ms":  {500, 1000},
			"1000ms - 1500ms": {1000, 1500},
			"> 1500ms":        {1500, 60000},
		}

		resourceUtilizationCategories := map[string]struct {
			Min, Max float64
		}{
			"< 5%":     {0, 5},
			"5% - 10%": {5, 10},
			"> 10%":    {10, 100},
		}

		for resUtilLabel, resourceRange := range resourceUtilizationCategories {
			if log_.ResourceUtilization >= resourceRange.Min && log_.ResourceUtilization <= resourceRange.Max {
				for respTimeLabel, respRange := range responseTimeCategories {
					if log_.ResponseTime >= respRange.Min && log_.ResponseTime <= respRange.Max {
						switch resUtilLabel {
						case "< 5%":
							switch respTimeLabel {
							case "< 500ms":
								dataset1.Data[0] = dataset1.Data[0] + 1

							case "500ms - 1000ms":
								dataset1.Data[1] = dataset1.Data[1] + 1

							case "1000ms - 1500ms":
								dataset1.Data[2] = dataset1.Data[2] + 1

							case "> 1500ms":
								dataset1.Data[3] = dataset1.Data[3] + 1

							default:
								fmt.Printf("uknown label: %s", respTimeLabel)
							}

						case "5% - 10%":
							switch respTimeLabel {
							case "< 500ms":
								dataset2.Data[0] = dataset2.Data[0] + 1

							case "500ms - 1000ms":
								dataset2.Data[1] = dataset2.Data[1] + 1

							case "1000ms - 1500ms":
								dataset2.Data[2] = dataset2.Data[2] + 1

							case "> 1500ms":
								dataset2.Data[3] = dataset2.Data[3] + 1

							default:
								fmt.Printf("uknown label: %s", respTimeLabel)
							}

						case "> 10%":
							switch respTimeLabel {
							case "< 500ms":
								dataset3.Data[0] = dataset3.Data[0] + 1

							case "500ms - 1000ms":
								dataset3.Data[1] = dataset3.Data[1] + 1

							case "1000ms - 1500ms":
								dataset3.Data[2] = dataset3.Data[2] + 1

							case "> 1500ms":
								dataset3.Data[3] = dataset3.Data[3] + 1

							default:
								fmt.Printf("uknown label: %s", respTimeLabel)
							}

						default:
							fmt.Printf("unknown label: %s", resUtilLabel)
						}

						break // break rest looping
					}
				}

				break // the rest looping
			}
		}
	}

	dataChart := struct {
		Labels   []string  `json:"labels"`
		Datasets []Dataset `json:"datasets"`
	}{
		Labels:   []string{"< 500ms", "500ms - 1000ms", "1000ms - 1500ms", "> 1500ms"},
		Datasets: []Dataset{dataset1, dataset2, dataset3},
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"chart": dataChart,
			"logs":  cacheLogs,
		},
	})
}

/* CACHE ASIDE */

/* DASHBOARD */
func (h *AdministratorHandlers) GetPerformanceTestResults(ctx *gin.Context) {
	gormDB, errGorm := initializer.GetGorm()
	if errGorm != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errGorm.Error(),
			"message": "gorm: failed getting connection instance",
		})

		ctx.Abort()
		return
	}

	type SessionTestResults struct {
		ID    int    `gorm:"column:id" json:"id"`
		Label string `gorm:"column:label" json:"label"`
		// Status                 *string   `gorm:"column:status" json:"status"`
		CreatedAt              time.Time `gorm:"column:created_at" json:"created_at"`
		TotalOfCacheHit        int       `gorm:"column:total_of_cache_hit" json:"total_of_cache_hit"`
		NumberOfCacheHit       int       `gorm:"column:number_of_cache_hit" json:"number_of_cache_hit"`
		TotalOfCacheMiss       int       `gorm:"column:total_of_cache_miss" json:"total_of_cache_miss"`
		NumberOfCacheMiss      int       `gorm:"column:number_of_cache_miss" json:"number_of_cache_miss"`
		AVGResponseTime        int       `gorm:"column:avg_response_time" json:"avg_response_time"`
		AVGMemoryUsage         float64   `gorm:"column:avg_memory_usage" json:"avg_memory_usage"`
		AVGCPUUsage            float64   `gorm:"column:avg_cpu_usage" json:"avg_cpu_usage"`
		AVGResourceUtilization float64   `gorm:"column:avg_resource_utilization" json:"avg_resource_utilization"`
	}

	sessionsResult := []SessionTestResults{}
	getSessionsResult := gormDB.Model(&models.CacheSession{}).Select([]string{
		"cache_sessions.id",
		"cache_sessions.label",
		// "cache_sessions.status",
		"cache_sessions.created_at",
		"SUM(request_logs.cache_hit) AS total_of_cache_hit",
		"COUNT(request_logs.cache_hit) AS number_of_cache_hit",
		"SUM(request_logs.cache_miss) AS total_of_cache_miss",
		"COUNT(request_logs.cache_miss) AS number_of_cache_miss",
		"AVG(request_logs.response_time) AS avg_response_time",
		"AVG(request_logs.memory_usage) AS avg_memory_usage",
		"AVG(request_logs.cpu_usage) AS avg_cpu_usage",
		"AVG(request_logs.resource_utilization) AS avg_resource_utilization",
	}).Joins("LEFT JOIN request_logs ON request_logs.cache_session_id = cache_sessions.id").
		Group("cache_sessions.id, cache_sessions.label, cache_sessions.created_at").Find(&sessionsResult)

	if getSessionsResult.Error != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   getSessionsResult.Error.Error(),
			"message": "error sql server query",
		})

		ctx.Abort()
		return
	}

	for index, result := range sessionsResult {
		ptrResult := &sessionsResult[index]
		formattedAVGMem, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", result.AVGMemoryUsage), 64)
		formattedAVGCPU, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", result.AVGCPUUsage), 64)
		formattedAVGResource, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", result.AVGResourceUtilization), 64)

		ptrResult.AVGMemoryUsage = formattedAVGMem
		ptrResult.AVGCPUUsage = formattedAVGCPU
		ptrResult.AVGResourceUtilization = formattedAVGResource
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    sessionsResult,
	})
}
