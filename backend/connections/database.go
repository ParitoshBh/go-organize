package connections

import (
	"database/sql"
	"go-organizer/backend/logger"
	"go-organizer/backend/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var (
	sDB *sql.DB
	gDB *gorm.DB
)

func InitDatabaseConnection() {
	_logger := logger.Logger

	goOrmDB, err := gorm.Open(sqlite.Open("server.db"), &gorm.Config{})
	if err != nil {
		_logger.Fatal(err.Error())
	}

	// Migrate the schema
	goOrmDB.AutoMigrate(&models.User{})

	// user := models.User{Username: "paritosh", Password: "password", FirstName: "Paritosh", LastName: "Bhatia"}
	// result := goOrmDB.Create(&user)
	// _logger.Info(result)

	sqlDB, err := goOrmDB.DB()
	if err != nil {
		_logger.Fatal(err.Error())
	}

	gDB = goOrmDB
	sDB = sqlDB
}

func GetSqlDBConnection() *sql.DB {
	return sDB
}

func GetGoOrmDBConnection() *gorm.DB {
	return gDB
}
