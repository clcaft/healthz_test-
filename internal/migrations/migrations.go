package migrations

import (
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"gl.eda1.ru/go/go-service-template/config"
)

type ExampleCounter struct {
	ID    int64 `gorm:"column:id;primaryKey"`
	Value int64 `gorm:"column:value;not null;default:0"`
}

func (ExampleCounter) TableName() string {
	return "example_counters"
}

func Run(cfg config.MySQL) error {
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=True&loc=Local",
		cfg.Username,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.DBName,
		cfg.Charset,
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("gorm open mysql: %w", err)
	}

	if err = db.AutoMigrate(&ExampleCounter{}); err != nil {
		return fmt.Errorf("auto migrate: %w", err)
	}

	return nil
}
