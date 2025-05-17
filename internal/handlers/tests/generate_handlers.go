package tests

import (
	"fmt"
	initializer "future-interns-backend/init"
	"future-interns-backend/internal/constants"
	"future-interns-backend/internal/models"
	"log"
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/brianvoe/gofakeit/v7/source"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type GenerateHandler struct {
}

type CombinationValue struct {
	LineIndustry    string `json:"line_industry"`
	EmployeeType    string `json:"employee_type"`
	WorkArrangement string `json:"work_arrangement"`
}

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

func MakeUUID(useFor string) string {
	namespace := uuid.Must(uuid.NewRandom())
	data := []byte("vacancy")
	sha1ID := uuid.NewSHA1(namespace, data)

	return sha1ID.String()
}

func MakeJobDesc(faker *gofakeit.Faker) string {
	randomSenteces := []string{
		faker.Sentence(10),
		faker.Sentence(14),
		faker.Sentence(18),
	}

	return strings.Join(randomSenteces, ". ")
}

func MakeUnorderedList(listFor string, faker *gofakeit.Faker) string {
	listedSentences := []string{}
	for i := 0; i < 3; i++ {
		if listFor == "qualifications" {
			listedSentences = append(listedSentences, "- "+faker.BuzzWord()+" "+faker.Sentence(10))
		}
		if listFor == "responsibilities" {
			listedSentences = append(listedSentences, "* "+faker.BuzzWord()+" "+faker.Sentence(10))
		}
	}

	return strings.Join(listedSentences, "\n")
}
func StratifiedSampling(totalSampling int) []CombinationValue {
	var allCombinations []CombinationValue

	for _, LineIndustry := range constants.LineIndustries {
		for _, EmployeeType := range constants.EmployeeType {
			for _, WorkArrangement := range constants.WorkArrangements {
				allCombinations = append(allCombinations, CombinationValue{
					LineIndustry:    LineIndustry,
					EmployeeType:    EmployeeType,
					WorkArrangement: WorkArrangement,
				})
			}
		}
	}

	groupByLineIndustry := map[string][]CombinationValue{}
	for _, LineIndustry := range constants.LineIndustries {
		for _, Combinations := range allCombinations {
			if Combinations.LineIndustry == LineIndustry {
				groupByLineIndustry[LineIndustry] = append(groupByLineIndustry[LineIndustry], Combinations)
			}
		}
	}

	statifiedSamples := []CombinationValue{}
	for _, group := range groupByLineIndustry {
		rand.Seed(time.Now().UnixNano())
		rand.Shuffle(len(group), func(i, j int) {
			group[i], group[j] = group[j], group[i]
		})
		samplePerIndustry := math.Round(float64(totalSampling / len(constants.LineIndustries)))
		statifiedSamples = append(statifiedSamples, group[:int(samplePerIndustry)]...)
	}

	return statifiedSamples
}

func (handler *GenerateHandler) MakeFakeVacancies(ctx *gin.Context) {
	var options struct {
		Sampling          []CombinationValue `json:"sampling" bonding:"required"`
		Offset            int                `json:"offset" binding:"required"`
		TotalRawVacancies int                `json:"total_raw_vacancies" binding:"required"`
	}
	if errBind := ctx.ShouldBindJSON(&options); errBind != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   errBind.Error(),
			"message": "double check JSON fields",
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
	var employersID []string
	getEmployers := gormDB.Model(&models.Employer{}).Select("id").Find(&employersID)
	if getEmployers.RowsAffected == 0 {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "no employers data",
			"message": "before generating vacancies the employer data should be exist",
		})
		return
	}

	vacanciesFaker := gofakeit.NewFaker(source.NewCrypto(), true)

	rawVacancies := []Vacancies{}

	for i := 0; i < options.TotalRawVacancies; i++ {
		rawVacancies = append(rawVacancies, Vacancies{
			ID:             MakeUUID("vacancy"),
			Position:       vacanciesFaker.JobTitle(),
			Description:    MakeJobDesc(vacanciesFaker),
			Qualification:  MakeUnorderedList("qualifications", vacanciesFaker),
			Responsibility: MakeUnorderedList("responsibilities", vacanciesFaker),
			LineIndustry:   options.Sampling[options.Offset-1].LineIndustry,
			EmployeeType:   options.Sampling[options.Offset-1].EmployeeType,
			MinExperience: vacanciesFaker.RandomString([]string{
				"No experience required",
				"Less than 1 year",
				"1-2 years",
				"3-5 years",
				"6-10 years",
				"More than 10 years",
			}),
			Salary:          int64(vacanciesFaker.IntRange(1000000, 100000000)),
			WorkArrangement: options.Sampling[options.Offset-1].WorkArrangement,
			SLA:             168,
			IsInactive:      vacanciesFaker.RandomInt([]int{1, 0}),
			EmployerID:      employersID[rand.Intn(len(employersID))],
			CreatedAt:       time.Now(),
		})
	}
	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    rawVacancies,
	})
}

func (handler *GenerateHandler) MakeSampling(ctx *gin.Context) {
	countQuery, isSamplingExist := ctx.GetQuery("count")
	if !isSamplingExist {
		countQuery = "120"
	}
	samplingCount, errConv := strconv.Atoi(countQuery)
	if errConv != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   errConv.Error(),
			"message": "'count' should be a valid number",
		})
		return
	} else if samplingCount > 800 {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid sampling count",
			"message": "'count' should be less than 800, maximum combinations is 800",
		})
		return
	}

	sampling := StratifiedSampling(samplingCount)

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    sampling,
	})
}

func (handler *GenerateHandler) StoreRawVacancies(ctx *gin.Context) {
	var RequestBody []struct {
		ID              string    `json:"id"`
		Position        string    `json:"position"`
		Description     string    `json:"description"`
		Qualification   string    `json:"qualification"`
		Responsibility  string    `json:"responsibility"`
		LineIndustry    string    `json:"line_industry"`
		EmployeeType    string    `json:"employee_type"`
		MinExperience   string    `json:"min_experience"`
		Salary          uint      `json:"salary"`
		WorkArrangement string    `json:"work_arrangement"`
		SLA             int32     `json:"sla"`
		IsInactive      int       `json:"is_inactive"`
		EmployerID      string    `json:"employer_id"`
		CreatedAt       time.Time `json:"created_at"`
	}

	if errBind := ctx.ShouldBindJSON(&RequestBody); errBind != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   errBind.Error(),
			"message": "failed to bind collection of JSON fields",
		})

		ctx.Abort()
		return
	}

	log.Println(len(RequestBody))

	gormDB, errGorm := initializer.GetGorm()
	if errGorm != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errGorm.Error(),
			"message": "failed getting GORM instance",
		})

		ctx.Abort()
		return
	}

	m_vacancies := []models.Vacancy{}
	vacanciesID := []string{}
	for _, vacancy := range RequestBody {
		vacanciesID = append(vacanciesID, vacancy.ID)
		m_vacancies = append(m_vacancies, models.Vacancy{
			Id:              vacancy.ID,
			Position:        vacancy.Position,
			Description:     vacancy.Description,
			Qualification:   vacancy.Qualification,
			Responsibility:  vacancy.Responsibility,
			LineIndustry:    vacancy.LineIndustry,
			EmployeeType:    vacancy.EmployeeType,
			MinExperience:   vacancy.MinExperience,
			Salary:          vacancy.Salary,
			WorkArrangement: vacancy.WorkArrangement,
			SLA:             vacancy.SLA,
			IsInactive:      vacancy.IsInactive != 0,
			EmployerId:      vacancy.EmployerID,
			CreatedAt:       vacancy.CreatedAt,
		})
	}

	errStoreVacancies := gormDB.CreateInBatches(&m_vacancies, 100).Error
	if errStoreVacancies != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errStoreVacancies.Error(),
			"message": fmt.Sprintf("failed storing %v vacancies", len(m_vacancies)),
		})

		ctx.Abort()
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    vacanciesID,
	})
}

func (handler *GenerateHandler) DeleteFakeVacancies(ctx *gin.Context) {
	countQuery, _ := ctx.GetQuery("count")
	count, errConv := strconv.Atoi(countQuery)
	if errConv != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   errConv.Error(),
			"message": "'count' must be a valid number",
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

	sqlQuery := `
		WITH del_vacancies AS (
			SELECT TOP (?) *
			FROM
				vacancies
			ORDER BY
				created_at DESC
		)
		DELETE FROM del_vacancies
	`

	delete := gormDB.Exec(sqlQuery, count)
	if delete.Error != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   delete.Error.Error(),
			"message": "there was an issue with sql query",
		})
		return
	}

	log.Println("deleted vacancies: \t", delete.RowsAffected)
	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    delete.RowsAffected,
	})
}
