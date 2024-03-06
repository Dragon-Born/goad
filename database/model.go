package database

import (
	log "github.com/sirupsen/logrus"
	"goad/pkg/yaml"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"strings"
)

var (
	DB     *gorm.DB
	OldDB  *gorm.DB
	Config *yaml.Config
)

const (
	CoinRate   = 1000000000.0
	EthDecimal = 10e18
)

func InitDB() (err error) {
	log.Info("Initializing database...")
	gCnf := &gorm.Config{
		Logger:                                   logger.Default.LogMode(logger.Silent),
		DisableForeignKeyConstraintWhenMigrating: true,
		PrepareStmt:                              true,
	}
	dbConfig := Config.Database.DBs[Config.Database.CurrentMode]
	if dbConfig.Type == "sqlite" {
		DB, err = gorm.Open(sqlite.Open(dbConfig.URI), gCnf)
	} else if dbConfig.Type == "mysql" {
		DB, err = gorm.Open(mysql.Open(dbConfig.URI), gCnf)
	} else {
		log.Fatalf("unknown database type: %v", dbConfig.Type)
		return
	}
	if err != nil {
		log.Fatal("failed to connect database: ", err)
		return
	}
	models := []any{
		&Wallet{},
	}
	err = DB.AutoMigrate(models...)
	if err != nil {
		log.Fatal("failed to migrate database: ", err)
		return
	}
	db, err := DB.DB()
	if err != nil {
		return err
	}
	db.SetMaxOpenConns(1000)
	db.SetMaxIdleConns(1000)
	log.Infof("%v Database has been initialized", strings.ToUpper(dbConfig.Type))
	return
}

func InitOldDB(path string) (err error) {
	gCnf := &gorm.Config{
		Logger:                                   logger.Default.LogMode(logger.Silent),
		DisableForeignKeyConstraintWhenMigrating: true,
		PrepareStmt:                              true,
	}
	OldDB, err = gorm.Open(sqlite.Open(path), gCnf)
	if err != nil {
		log.Fatal("failed to connect to old database: ", err)
		return
	}
	return
}
