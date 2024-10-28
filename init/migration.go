package initializer

import (
	"future-interns-backend/internal/models"
	"log"
)

func Migrate(target string) {
	gormDB, err := GetGorm()
	if err != nil {
		panic(err)
	}
	switch target {
	case "basic":
		errDrop := gormDB.Migrator().DropTable(&models.User{}, &models.Image{}, &models.Document{}, &models.Address{}, &models.Social{})
		if errDrop != nil {
			panic(errDrop)
		}

		errMigrate := gormDB.AutoMigrate(&models.User{}, &models.Image{}, &models.Document{}, &models.Address{}, &models.Social{})
		if errMigrate != nil {
			panic(errMigrate)
		}
		log.Println("migrate --target=basic completed")
	case "candidate":
		errDrop := gormDB.Migrator().DropTable(&models.Candidate{}, &models.Education{}, &models.Experience{}, &models.Skill{}, &models.CandidateSkill{}, &models.CandidateAddress{}, &models.CandidateSocial{})
		if errDrop != nil {
			panic(errDrop)
		}
		// setup jointable candidate skill
		errJoinCandidateSkill := gormDB.SetupJoinTable(&models.Candidate{}, "Skills", &models.CandidateSkill{})
		if errJoinCandidateSkill != nil {
			panic(errJoinCandidateSkill)
		}
		// setup jointable candidate address
		errJoinCandidateAddress := gormDB.SetupJoinTable(&models.Candidate{}, "Addresses", &models.CandidateAddress{})
		if errJoinCandidateAddress != nil {
			panic(errJoinCandidateAddress)
		}
		// setup jointable candidate social
		errJoinCandidateSocial := gormDB.SetupJoinTable(&models.Candidate{}, "Socials", &models.CandidateSocial{})
		if errJoinCandidateSocial != nil {
			panic(errJoinCandidateSocial)
		}
		errMigrate := gormDB.AutoMigrate(&models.Candidate{}, &models.Education{}, &models.Experience{}, &models.Skill{}, &models.CandidateSkill{}, &models.CandidateAddress{}, &models.CandidateSocial{})
		if errMigrate != nil {
			panic(errMigrate)
		}

		log.Println("migrate --target=candidate completed")
	case "employer":
		errDrop := gormDB.Migrator().DropTable(&models.Employer{}, &models.Headquarter{}, &models.OfficeImage{}, &models.EmployerSocial{}, &models.Vacancy{}, &models.Pipeline{}, &models.Assessment{}, &models.AssessmentDocument{}, &models.AssessmentAssignee{}, &models.AssessmentAssigneeSubmission{}, &models.Interview{}, &models.Offering{})
		if errDrop != nil {
			panic(errDrop)
		}
		// setup jointable employer social
		errJoinTable := gormDB.SetupJoinTable(&models.Employer{}, "Socials", &models.EmployerSocial{})
		if errJoinTable != nil {
			panic(errJoinTable)
		}
		errMigrate := gormDB.AutoMigrate(&models.Employer{}, &models.Headquarter{}, &models.OfficeImage{}, &models.EmployerSocial{}, &models.Vacancy{}, &models.Pipeline{}, &models.Assessment{}, &models.AssessmentDocument{}, &models.AssessmentAssignee{}, &models.AssessmentAssigneeSubmission{}, &models.Interview{}, &models.Offering{})
		if errMigrate != nil {
			panic(errMigrate)
		}

		log.Println("migrate --target=employer completed")
	case "dropAll":
		errDrop := gormDB.Migrator().DropTable(&models.Employer{}, &models.Headquarter{}, &models.OfficeImage{}, &models.EmployerSocial{}, &models.Vacancy{}, &models.Pipeline{}, &models.Assessment{}, &models.AssessmentDocument{}, &models.AssessmentAssignee{}, &models.AssessmentAssigneeSubmission{}, &models.Interview{}, &models.Offering{})
		if errDrop != nil {
			panic(errDrop)
		}
	}
}
