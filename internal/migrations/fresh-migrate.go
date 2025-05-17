package migrations

import (
	"fmt"
	"time"

	initializer "database-migration-cli/init"
	"database-migration-cli/internal/models"
)

func FreshMigration() {
	gormDB, errGorm := initializer.GetGorm()
	if errGorm != nil {
		panic(errGorm)
	}

	modelsToMigrate := []interface{}{
		// master
		&models.User{},
		&models.Skill{},
		&models.Social{},
		&models.Address{},
		&models.Image{},
		&models.Document{},
		&models.IdentityAccess{},
		&models.Role{},
		&models.Permission{},
		&models.RolePermission{},
		&models.CacheSession{},
		&models.RequestLog{},
		// candidate
		&models.Candidate{},
		&models.Education{},
		&models.CandidateSkill{},
		&models.Experience{},
		&models.CandidateSocial{},
		&models.CandidateAddress{},
		// employers
		&models.Employer{},
		&models.Headquarter{},
		&models.OfficeImage{},
		&models.EmployerSocial{},
		&models.Vacancy{},
		&models.Pipeline{},
		&models.Assessment{},
		&models.Interview{},
		&models.Offering{},
		&models.AssessmentDocument{},
		&models.AssessmentAssignee{},
		&models.AssessmentAssigneeSubmission{},
	}

	errDropTables := gormDB.Migrator().DropTable(modelsToMigrate...)
	if errDropTables != nil {
		fmt.Println(errDropTables.Error())
		return
	}

	errMigrate := gormDB.AutoMigrate(modelsToMigrate...)
	if errMigrate != nil {
		fmt.Println(errMigrate.Error())
		return
	}

	defaultRoles := []models.Role{
		{
			Name:        "basic/candidate",
			Description: "A basic access as a candidate",
			CreatedAt:   time.Now(),
		},
		{
			Name:        "basic/employer",
			Description: "A basic access as an employer",
			CreatedAt:   time.Now(),
		},
		{
			Name:        "basic/university",
			Description: "A basic access as a university administrator",
			CreatedAt:   time.Now(),
		},
		{
			Name:        "sdkdev/administrator",
			Description: "A complete access to any resources for future-interns project",
			CreatedAt:   time.Now(),
		},
	}
	errCreate := gormDB.Create(&defaultRoles).Error
	if errCreate != nil {
		panic(errCreate)
	}

	fmt.Println("tables migrated successfully...")
}
