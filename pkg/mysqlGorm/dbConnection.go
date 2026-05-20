package mysqlGorm

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"gl.eda1.ru/go/go-service-template/config"
)

const (
	maxOpenConns = 15
	maxLifetime  = 4 * time.Hour
)

var conn *Connector

// Connector use singleton.GetDBConnector() for connector getting.
type Connector struct {
	masterMtx sync.Mutex
	master    *gorm.DB

	slavesMutex sync.Mutex
	slaves      []*gorm.DB

	maxSlavesCount int
	cfg            *config.MySQL
}

func GetConnector(cfg *config.MySQL) *Connector {
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
		slaves:         make([]*gorm.DB, maxSlavesCount),
		cfg:            cfg,
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

func (c *Connector) GetMaster(ctx context.Context) (*gorm.DB, error) {
	c.masterMtx.Lock()
	defer c.masterMtx.Unlock()

	if c.master == nil {
		db, err := findDBConnection(ctx, 0)
		if err != nil {
			return nil, err
		}

		c.master = db
	}

	return c.master, nil
}

func (c *Connector) GetSlave(ctx context.Context) (*gorm.DB, error) {
	rndSlave := rand.Intn(c.maxSlavesCount)

	conn, err := c.findSlaveConnection(ctx, rndSlave+1)
	if err == nil {
		return conn, nil
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

		return conn, nil
	}

	// все слейвы недоступны
	return nil, err
}

func (c *Connector) findSlaveConnection(ctx context.Context, num int) (*gorm.DB, error) {
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

func findDBConnection(ctx context.Context, slaveDBNumber int) (*gorm.DB, error) {
	data := setConnectionVal(new(connectorData), slaveDBNumber)

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local&timeout=1s",
		os.Getenv(data.Name),
		os.Getenv(data.Password),
		os.Getenv(data.Host),
		os.Getenv(data.Port),
		os.Getenv(data.DBName),
	)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
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
}

func (c *Connector) GetTimeout() time.Duration {
	return c.cfg.Timeout
}

func (c *Connector) GetRefreshInterval() time.Duration {
	return c.cfg.RefreshInterval
}
