package mysql

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"

	"gl.eda1.ru/go/go-service-template/config"
	"gl.eda1.ru/go/go-service-template/pkg/tracer"
)

const (
	maxOpenConns = 15
	maxLifetime  = 4 * time.Hour
)

var conn *Connector

// Connector use singleton.GetDBConnector() for connector getting.
type Connector struct {
	masterMtx sync.Mutex
	master    *sqlx.DB

	slavesMutex sync.Mutex
	slaves      []*sqlx.DB

	maxSlavesCount int
	cfg            *config.MySQL
	traces         tracer.TraceProvider
}

func GetConnector(cfg *config.MySQL, traces tracer.TraceProvider) *Connector {
	if conn != nil {
		return conn
	}

	maxSlavesCount, err := strconv.Atoi(os.Getenv("DB_SLAVE_COUNT"))
	if err != nil {
		log.Error().Msgf("Error parse max slaves count from env, max slave count set as 1, error was: %s", err)

		maxSlavesCount = 1
	}

	conn = &Connector{
		maxSlavesCount: maxSlavesCount,
		slaves:         make([]*sqlx.DB, maxSlavesCount),
		cfg:            cfg,
		traces:         traces,
	}

	return conn
}

type connectorData struct {
	Name           string
	Password       string
	Host           string
	Port           string
	DBName         string
	Charset        string
	NativePassword string
}

func (c *Connector) GetMaster(ctx context.Context) (*sqlx.DB, error) {
	c.masterMtx.Lock()
	defer c.masterMtx.Unlock()

	if c.master == nil {
		db, err := findDBConnection(ctx, 0)
		if err != nil {
			return nil, err
		}

		c.master = db
	}

	if err := c.master.Ping(); err != nil {
		return nil, err
	}

	return c.master, nil
}

func (c *Connector) GetSlave(ctx context.Context) (*sqlx.DB, error) {
	rndSlave := rand.Intn(c.maxSlavesCount)

	conn, err := c.findSlaveConnection(ctx, rndSlave+1)
	if err == nil {
		if err = conn.Ping(); err == nil {
			return conn, nil
		}
	}

	// пробуем найти другой живой
	for i := 0; i < c.maxSlavesCount; i++ {
		if i == rndSlave {
			continue
		}

		conn, err = c.findSlaveConnection(ctx, i+1)
		// Ошибка формирования sql.DB
		if err != nil {
			continue
		}

		// Проверка соединения
		if err = conn.Ping(); err != nil {
			continue
		}

		return conn, nil
	}

	// все слейвы недоступны
	return nil, err
}

func (c *Connector) findSlaveConnection(ctx context.Context, num int) (*sqlx.DB, error) {
	c.slavesMutex.Lock()
	defer c.slavesMutex.Unlock()

	if c.slaves[num-1] != nil {
		return c.slaves[num-1], nil
	}

	db, err := findDBConnection(ctx, num)
	if err != nil {
		return nil, err
	}

	c.slaves[num-1] = db

	return db, nil
}

func findDBConnection(ctx context.Context, slaveDBNumber int) (*sqlx.DB, error) {
	data := setConnectionVal(new(connectorData), slaveDBNumber)
	conf := &mysql.Config{
		User:   os.Getenv(data.Name),
		Passwd: os.Getenv(data.Password),
		Net:    "tcp",
		Addr:   fmt.Sprintf("%s:%s", os.Getenv(data.Host), os.Getenv(data.Port)),
		DBName: os.Getenv(data.DBName),
		Params: map[string]string{
			"charset": os.Getenv(data.Charset),
		},
		ParseTime:            true,
		Loc:                  time.Local,
		Timeout:              time.Second,
		AllowNativePasswords: os.Getenv(data.NativePassword) == "1",
	}
	db, err := sqlx.Open("mysql", conf.FormatDSN())
	if err != nil {
		err = fmt.Errorf("failed to connect database: %w", err)
		slog.Default().Log(ctx, slog.LevelError, err.Error())
		return nil, err
	}

	// TODO: оттюнить под конкретные настройки базы, это важно! Возможно вынести настройки из констант в конфиг
	db.SetConnMaxLifetime(maxLifetime)
	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(maxOpenConns)

	err = db.Ping()
	if err != nil {
		err = fmt.Errorf("failed to connect database: %w", err)
		slog.Default().Log(ctx, slog.LevelError, err.Error())
		return nil, err
	}

	return db, nil
}

func setConnectionVal(data *connectorData, slaveDBNumber int) *connectorData {
	data.Name = "DB_USERNAME"
	data.Password = "DB_PASSWORD"
	data.Host = "DB_HOST"
	data.Port = "DB_PORT"
	data.DBName = "DB_NAME"
	data.Charset = "DB_CHARSET"
	data.NativePassword = "DB_USE_NATIVE_PASSWORD"

	// на репликах отличается только хост, остальные параметры идентичны
	if slaveDBNumber != 0 {
		data.Host = fmt.Sprintf("%s_%d", data.Host, slaveDBNumber)
	}

	return data
}

func (c *Connector) Close() {
	var err error
	if c.master != nil {
		err = c.master.Close()
		if err != nil {
			fmt.Printf("error closing master: %v", err)
		}
	}

	for i, slave := range c.slaves {
		if slave != nil {
			err = slave.Close()
			if err != nil {
				fmt.Printf("error closing slave #%d: %v", i, err)
			}
		}
	}
}

func (c *Connector) GetTimeout() time.Duration {
	return c.cfg.Timeout
}

func (c *Connector) GetRefreshInterval() time.Duration {
	return c.cfg.RefreshInterval
}
