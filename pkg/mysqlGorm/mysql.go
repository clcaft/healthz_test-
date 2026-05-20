package mysqlGorm

import (
	"database/sql"

	"github.com/go-sql-driver/mysql"
)

func NewDB(conf *mysql.Config) (*sql.DB, error) {
	db, err := sql.Open("mysql", conf.FormatDSN())
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
