package main

import (
	"database/sql"
	"flag"
	"fmt"
	"github.com/BurntSushi/toml"
	_ "github.com/go-sql-driver/mysql"
	"net"
	"os"
	"strconv"
	"sync"
)

type Config struct {
	MySql  MysqlConfig  `toml:"mysql"`
	FFMpeg FFMpegConfig `toml:"ffmpeg"`
}

type MysqlConfig struct {
	Host     string
	Port     uint16
	Username string
	Password string
	Database string
	Table    string
}

type FFMpegConfig struct {
	Path    string
	Timeout uint
	Workers uint
}

var configFilename = flag.String("c", "", "configuration filename")
var logFilename = flag.String("l", "", "log filename")

var config Config
var db *sql.DB
var threads sync.WaitGroup

func main() {
	var err error

	reopenLogger()

	flag.Parse()
	if len(*configFilename) == 0 {
		fmt.Fprintf(os.Stderr, "Error: configuration filename is absent\n")
		logger.Fatalf("Error: configuration filename is absent\n")
	}

	if _, err := toml.DecodeFile(*configFilename, &config); err != nil {
		logger.Fatalf("Can't parse configuration file: %s\n", err.Error())
	}

	timeout := uint(60)
	if (2 * config.FFMpeg.Timeout) > timeout {
		timeout = 2 * config.FFMpeg.Timeout
	}

	dsn :=
		fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8&timeout=%ds&wait_timeout=%d",
			config.MySql.Username,
			config.MySql.Password,
			net.JoinHostPort(config.MySql.Host, strconv.Itoa(int(config.MySql.Port))),
			config.MySql.Database,
			timeout,
			timeout)

	logger.Printf("Open database connection with DSN: %s\n", dsn)

	db, err = sql.Open("mysql", dsn)
	if err != nil {
		logger.Fatalf("Can't connect to database server: %s\n", err.Error())
	}
	defer db.Close()

	go signalHandler()

	runJobQueue()
	runQueueReader()

	threads.Wait()
}
