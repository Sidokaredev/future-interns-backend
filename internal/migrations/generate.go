package migrations

// import (
// 	"context"
// 	initializer "database-migration-cli/init"
// 	"database-migration-cli/internal/models"
// 	"fmt"
// 	"log"
// 	"strings"
// 	"time"

// 	"github.com/brianvoe/gofakeit/v7"
// 	"github.com/brianvoe/gofakeit/v7/source"
// 	"github.com/google/uuid"
// 	"golang.org/x/crypto/bcrypt"
// )

// func GetUUID(name string) string {
// 	namespace := uuid.Must(uuid.NewRandom())
// 	data := []byte(name)

// 	sha1 := uuid.NewSHA1(namespace, data)
// 	return sha1.String()
// }

// func Hash(password string) string {
// 	hashed, errHash := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
// 	if errHash != nil {
// 		panic(errHash)
// 	}
// 	return string(hashed)
// }
// func GenerateUser() {
// 	fake1 := gofakeit.NewFaker(source.NewCrypto(), true)

// 	users := make([]models.User, 0, 5000)
// 	for i := 0; i < 5000; i++ {
// 		fullname := fake1.Name()
// 		email := fake1.Email()
// 		users = append(users, models.User{
// 			ID:        GetUUID(fullname),
// 			Fullname:  fullname,
// 			Email:     email,
// 			Password:  Hash(strings.Split(email, "@")[0]),
// 			CreatedAt: time.Now(),
// 		})
// 	}

// 	gormDB, errGorm := initializer.GetGorm()
// 	if errGorm != nil {
// 		panic(errGorm)
// 	}

// 	log.Println("start inserting users...")
// 	createUsers := gormDB.CreateInBatches(&users, 100)
// 	if createUsers.Error != nil {
// 		panic(createUsers)
// 	}
// 	log.Println("5000 users created!")
// }

// func GetUsers() {
// 	gormDB, errGorm := initializer.GetGorm()
// 	if errGorm != nil {
// 		panic(errGorm)
// 	}

// 	users := []models.Users{}
// 	start := time.Now()
// 	getUsers := gormDB.Model(&models.Users{}).Find(&users)
// 	if getUsers.RowsAffected == 0 {
// 		fmt.Println("no users found")
// 		return
// 	}
// 	// usersManipulate := []map[string]interface{}{}
// 	// for _, user := range users {
// 	// 	usersManipulate = append(usersManipulate, map[string]interface{}{
// 	// 		"id":    user.ID,
// 	// 		"email": user.Email,
// 	// 	})
// 	// }
// 	elapsed := time.Since(start)
// 	fmt.Printf("Query selesai dalam %v\n", elapsed)
// }

// func SetRKeyRedis() {
// 	// rdb := redis.NewClient(&redis.Options{
// 	// 	Addr:     "127.0.0.1:6379",
// 	// 	Password: "",
// 	// 	DB:       0,
// 	// 	Protocol: 3,
// 	// })

// 	// ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
// 	// defer cancel()

// 	rdb := initializer.GetRedis()
// 	ctx := context.Background()
// 	if err := rdb.Ping(ctx).Err(); err != nil {
// 		log.Fatalf("failed connect to redis: %v", err)
// 	} else {
// 		log.Println("redis connection established")
// 	}
// 	err := rdb.Set(ctx, "firsttime", "Fatkhur Rozak", 300*time.Second).Err()
// 	if err != nil {
// 		panic(err)
// 	}

// 	log.Println("key stored successfully")
// }
