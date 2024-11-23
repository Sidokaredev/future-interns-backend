package initializer

import (
	"future-interns-backend/internal/models"
	"log"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func Migrate(target string) {
	gormDB, err := GetGorm()
	if err != nil {
		panic(err)
	}
	switch target {
	case "basic":
		errDrop := gormDB.Migrator().DropTable(&models.User{}, &models.Image{}, &models.Document{}, &models.Address{}, &models.Social{}, &models.Role{}, &models.Permission{}, &models.RolePermission{}, &models.IdentityAccess{})
		if errDrop != nil {
			panic(errDrop)
		}

		errJoinRolePermission := gormDB.SetupJoinTable(&models.Role{}, "Permissions", &models.RolePermission{})
		if errJoinRolePermission != nil {
			panic(errJoinRolePermission)
		}

		errMigrate := gormDB.AutoMigrate(&models.User{}, &models.Image{}, &models.Document{}, &models.Address{}, &models.Social{}, &models.Role{}, &models.Permission{}, &models.RolePermission{}, &models.IdentityAccess{})
		if errMigrate != nil {
			panic(errMigrate)
		}

		roles := []models.Role{
			{Name: "basic/candidate", Description: "A basic access as a candidate"},
			{Name: "basic/employer", Description: "A basic access as an employer"},
			{Name: "basic/university", Description: "A basic access as a university administrator"},
			{Name: "sdkdev/administrator", Description: "A complete access to any resources for future-interns project"},
		}
		seedRoles := gormDB.Create(&roles)
		if seedRoles.Error != nil {
			log.Println(seedRoles.Error)
			panic(seedRoles.Error)
		}

		log.Println("creating sdkadmin account...")
		ADMINISTRATOR_PASS := "sdk.admin"
		hashed, errHash := bcrypt.GenerateFromPassword([]byte(ADMINISTRATOR_PASS), bcrypt.DefaultCost)
		if errHash != nil {
			log.Println(errHash.Error())
			panic(errHash)
		}

		namespace := uuid.Must(uuid.NewRandom())
		data := []byte("sdk-administrator")

		sha1ID := uuid.NewSHA1(namespace, data)

		administrator := models.User{
			Id:       sha1ID.String(),
			Fullname: "sdk-administrator",
			Email:    "sdk-admin@sidokaredev.com",
			Password: string(hashed),
		}
		errCreateAdministrator := gormDB.Transaction(func(tx *gorm.DB) error {
			errCreateUser := tx.Create(&administrator).Error
			if errCreateUser != nil {
				return errCreateUser
			}

			identityAccess := models.IdentityAccess{
				UserId: administrator.Id,
				RoleId: 4,
				Type:   "administrator",
			}
			errCreateIdentityAccess := tx.Create(&identityAccess).Error
			if errCreateIdentityAccess != nil {
				return errCreateIdentityAccess
			}

			return nil
		})

		if errCreateAdministrator != nil {
			log.Println(errCreateAdministrator.Error())
			panic(errCreateAdministrator)
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
		errDrop := gormDB.Migrator().DropTable(&models.User{}, &models.Image{}, &models.Document{}, &models.Address{}, &models.Social{}, &models.Role{}, &models.Permission{}, &models.RolePermission{}, &models.IdentityAccess{}, &models.Candidate{}, &models.Education{}, &models.Experience{}, &models.Skill{}, &models.CandidateSkill{}, &models.CandidateAddress{}, &models.CandidateSocial{}, &models.Employer{}, &models.Headquarter{}, &models.OfficeImage{}, &models.EmployerSocial{}, &models.Vacancy{}, &models.Pipeline{}, &models.Assessment{}, &models.AssessmentDocument{}, &models.AssessmentAssignee{}, &models.AssessmentAssigneeSubmission{}, &models.Interview{}, &models.Offering{})
		if errDrop != nil {
			panic(errDrop)
		}
	case "permissions":
		/* user creator */
		timeNow := time.Now()
		permissions := []models.Permission{
			/* user as candidate */
			{Name: "users.candidate.create", Resource: "users", Description: "Create a new user as a candidate.", CreatedAt: timeNow, UpdatedAt: &timeNow},
			{Name: "users.candidate.update", Resource: "users", Description: "Update user data as a candidate.", CreatedAt: timeNow, UpdatedAt: &timeNow},
			{Name: "users.candidate.get", Resource: "users", Description: "Get detailed user data as a candidate.", CreatedAt: timeNow, UpdatedAt: &timeNow},
			{Name: "users.candidate.list", Resource: "users", Description: "List all records of users as candidates.", CreatedAt: timeNow, UpdatedAt: &timeNow},
			{Name: "users.candidate.delete", Resource: "users", Description: "Delete the user's data as a candidate.", CreatedAt: timeNow, UpdatedAt: &timeNow},
			/* user as employer */
			{Name: "users.employer.create", Resource: "users", Description: "Create a new user as a employer.", CreatedAt: timeNow, UpdatedAt: &timeNow},
			{Name: "users.employer.update", Resource: "users", Description: "Update user data as a employer.", CreatedAt: timeNow, UpdatedAt: &timeNow},
			{Name: "users.employer.get", Resource: "users", Description: "Get detailed user data as a employer.", CreatedAt: timeNow, UpdatedAt: &timeNow},
			{Name: "users.employer.list", Resource: "users", Description: "List all records of users as employers.", CreatedAt: timeNow, UpdatedAt: &timeNow},
			{Name: "users.employer.delete", Resource: "users", Description: "Delete the user's data as a employer.", CreatedAt: timeNow, UpdatedAt: &timeNow},
			/* vacancies */
			{Name: "vacancies.sla.update", Resource: "vacancies", Description: "Update the SLA data in the vacancies table, where the SLA data is a numeric count of hours.", CreatedAt: time.Now()},
		}

		errCreatePermissions := gormDB.Create(&permissions).Error
		if errCreatePermissions != nil {
			panic(errCreatePermissions.Error())
		}

		log.Println("migrate --target=permissions completed")
	}
}
