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
	"text/template"
)

type Config struct {
	MySql   MysqlConfig   `toml:"mysql"`
	Command CommandConfig `toml:"command"`
}

type MysqlConfig struct {
	Host     string
	Port     uint16
	Username string
	Password string
	Database string
	Table    string
	Params   []string
}

type CommandConfig struct {
	Template string
	Timeout  uint
	Workers  uint
	Shell    string

	compiledTemplate *template.Template
}

var configFilename = flag.String("c", "", "configuration filename")
var logFilename = flag.String("l", "", "log filename")

var config Config
var db *sql.DB
var threads sync.WaitGroup

func dsn(for_log bool) string {
	timeout := uint(60)
	if (2 * config.Command.Timeout) > timeout {
		timeout = 2 * config.Command.Timeout
	}

	password := config.MySql.Password
	if for_log {
		password = "********"
	}

	dsn :=
		fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8&timeout=%ds&wait_timeout=%d",
			config.MySql.Username,
			password,
			net.JoinHostPort(config.MySql.Host, strconv.Itoa(int(config.MySql.Port))),
			config.MySql.Database,
			timeout,
			timeout)

	return dsn
}

func main() {
	var err error

	flag.Parse()

	reopenLogger()

	if len(*configFilename) == 0 {
		fmt.Fprintf(os.Stderr, "Error: configuration filename is absent\n")
		logger.Fatalf("Error: configuration filename is absent\n")
	}

	if _, err := toml.DecodeFile(*configFilename, &config); err != nil {
		logger.Fatalf("Can't parse configuration file: %s\n", err.Error())
	}

	config.Command.compiledTemplate, err = template.New("command").Parse(config.Command.Template)
	if err != nil {
		logger.Fatalf("Can't compile command template: %s\n", err.Error())
	}

	logger.Printf("Open database connection with DSN: %s\n", dsn(true))

	db, err = sql.Open("mysql", dsn(false))
	if err != nil {
		logger.Fatalf("Can't connect to database server: %s\n", err.Error())
	}
	defer db.Close()

	go signalHandler()

	runJobQueue()
	runQueueReader()

	threads.Wait()
}
