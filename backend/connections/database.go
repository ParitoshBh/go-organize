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

	// // create table for holding session data
	// goOrmDB.Exec(`CREATE TABLE sessions (
	// 	token TEXT PRIMARY KEY,
	// 	data BLOB NOT NULL,
	// 	expiry REAL NOT NULL
	// )`)
	// // add index on sessions table
	// goOrmDB.Exec("CREATE INDEX sessions_expiry_idx ON sessions(expiry)")

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
