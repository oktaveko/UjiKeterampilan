package configs

import (
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"os"
	"ujiketerampilan/models"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDatabase() {
	//loadenv()

	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_ROOT")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbUser, dbPassword, dbHost, dbPort, dbName)

	fmt.Println(dsn)
	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	initMigration()
}

func initMigration() {
	err := DB.AutoMigrate(&models.Book{}, &models.User{}, &models.Borrowing{})
	if err != nil {
		log.Fatalf("Error during database migration: %s", err)
	}
}

func loadenv() {
	err := godotenv.Load()
	if err != nil {
		panic("Failed load env file")
	}
}
