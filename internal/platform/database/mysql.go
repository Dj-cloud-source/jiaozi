package database

import (
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"

	"jiaozi/internal/config"
)

func OpenMySQL(cfg config.MySQLConfig) (*sqlx.DB, error) {
	mysqlConfig := mysql.Config{
		User:      cfg.Username,
		Passwd:    cfg.Password,
		Net:       "tcp",
		Addr:      cfg.Host + ":" + cfg.Port,
		DBName:    cfg.Database,
		ParseTime: true,
		Loc:       time.FixedZone("Asia/Shanghai", 8*60*60),
	}

	db, err := sqlx.Connect("mysql", mysqlConfig.FormatDSN())
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(time.Hour)

	return db, nil
}
