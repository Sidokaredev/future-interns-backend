package initializer

// func DatabaseInit() {
// 	gormDB, errDB := GetGorm()
// 	if errDB != nil {
// 		errMessage := fmt.Sprintf("Error getting GORM instance \t: %v", errDB.Error())
// 		fmt.Println(errMessage)
// 		return
// 	}

// 	listDroppedTable := []interface{}{
// 		"offerings",
// 		"interviews",
// 		"assessment_asignee_submissions",
// 		"assessment_assignees",
// 		"assessment_documents",
// 		"assessments",
// 		"pipelines",
// 		"candidate_socials",
// 		"candidate_skills",
// 		"candidate_addresses",
// 		"experiences",
// 		"educations",
// 		"candidates",
// 		"employer_socials",
// 		"office_images",
// 		"headquarters",
// 		"vacancies",
// 		"employers",
// 		"identity_accesses",
// 		"users",
// 		"role_permissions",
// 		"roles",
// 		"permissions",
// 		"addresses",
// 		"skills",
// 		"socials",
// 		"images",
// 		"documents",
// 	}
// 	errDropTables := gormDB.Migrator().DropTable(listDroppedTable...)
// 	if errDropTables != nil {
// 		fmt.Println(errDropTables.Error())
// 		return
// 	}

// 	migrateBasicTables := gormDB.AutoMigrate()
// }
