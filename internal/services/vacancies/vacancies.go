package vacancies

import (
	initializer "go-read-through-service/init"
	"go-read-through-service/internal/models"
	"log"

	"gorm.io/gorm"
)

type VacanciesArgs struct {
	Time         string // -> RFC3339 Format
	Location     string
	Keyword      string
	LineIndustry string
	EmployeeType string
}

func CountAndApplied(userID string, queries VacanciesArgs, count *int64, applied *[]string) error {
	DB, err := initializer.GetMssqlDB()
	if err != nil {
		return err
	}

	errTransacQuery := DB.Transaction(func(tx *gorm.DB) error {
		if userID == "" {
			applied = &[]string{}
		} else {
			var candidateID string
			fetchCandidate := tx.Model(&models.Candidate{}).Select("id").Where("user_id = ?", userID).First(&candidateID)
			if fetchCandidate.Error != nil {
				log.Printf("candidate id: %v", fetchCandidate.Error.Error())
			} else {
				fetchApplied := tx.Model(&models.Pipeline{}).Select("vacancy_id").Where("candidate_id = ?", candidateID).Find(applied)
				if fetchApplied.Error != nil {
					return fetchApplied.Error
				}
			}
		}

		fetchCount := tx.Model(&models.Vacancy{}).
			Joins("INNER JOIN employers ON employers.id = vacancies.employer_id").
			Where(`vacancies.is_inactive = ? AND
		    vacancies.created_at <= ? AND
		    employers.location LIKE ? AND
		    (vacancies.position LIKE ? OR
		    vacancies.description LIKE ? OR
		    vacancies.qualification LIKE ? OR
		    vacancies.responsibility LIKE ?) AND
		    vacancies.line_industry LIKE ? AND
		    vacancies.employee_type LIKE ?`,
				false,
				queries.Time,
				queries.Location,
				queries.Keyword,
				queries.Keyword,
				queries.Keyword,
				queries.Keyword,
				queries.LineIndustry,
				queries.EmployeeType,
			).
			Order("vacancies.created_at DESC").Count(count)

		if fetchCount.Error != nil {
			return fetchCount.Error
		}

		return nil
	})

	if errTransacQuery != nil {
		return errTransacQuery
	}
	return nil
}

// DESC: Jika userID adalah empty string [""], maka tidak ada nilai yang di assign ke 'candidateID'
func CountAndCandidateID(userID string, queries VacanciesArgs, count *int64, candidateID *string) error {
	DB, err := initializer.GetMssqlDB()
	if err != nil {
		return err
	}

	errQuery := DB.Transaction(func(tx *gorm.DB) error {
		if userID != "" {
			fetchCandidate := tx.Model(&models.Candidate{}).Select("id").Where("user_id = ?", userID).First(candidateID)
			if fetchCandidate.Error != nil {
				return fetchCandidate.Error
			}
		}

		fetchCount := tx.Model(&models.Vacancy{}).
			Joins("INNER JOIN employers ON employers.id = vacancies.employer_id").
			Where(`vacancies.is_inactive = ? AND
		    vacancies.created_at <= ? AND
		    employers.location LIKE ? AND
		    (vacancies.position LIKE ? OR
		    vacancies.description LIKE ? OR
		    vacancies.qualification LIKE ? OR
		    vacancies.responsibility LIKE ?) AND
		    vacancies.line_industry LIKE ? AND
		    vacancies.employee_type LIKE ?`,
				false,
				queries.Time,
				queries.Location,
				queries.Keyword,
				queries.Keyword,
				queries.Keyword,
				queries.Keyword,
				queries.LineIndustry,
				queries.EmployeeType,
			).
			Order("vacancies.created_at DESC").Count(count)

		if fetchCount.Error != nil {
			return fetchCount.Error
		}
		return nil
	})

	if errQuery != nil {
		return errQuery
	}

	return nil
}
