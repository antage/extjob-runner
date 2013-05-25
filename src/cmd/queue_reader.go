package main

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type queueReader struct {
	stm       *sql.Stmt
	lastId    int32
	forceNext chan bool
	quit      chan bool
}

func newQueueReader(quit chan bool) *queueReader {
	var err error

	instance := &queueReader{}
	instance.forceNext = make(chan bool, 1)
	instance.quit = quit
	instance.lastId = 0

	params_columns := strings.Join(config.MySql.Params, ", ")

	instance.stm, err = db.Prepare(fmt.Sprintf("SELECT id, %s FROM `%s` WHERE done = 0 AND id > ? ORDER BY `id` LIMIT 1", params_columns, config.MySql.Table))
	if err != nil {
		logger.Fatalf("Can't prepare SQL-statement: %s\n", err.Error())
	}

	return instance
}

func (instance *queueReader) processJob() {
	var job job
	columns := make([]interface{}, 0, len(config.MySql.Params)+1)
	columns = append(columns, &job.id)
	for _ = range config.MySql.Params {
		columns = append(columns, new([]byte))
	}
	if err := instance.stm.QueryRow(instance.lastId).Scan(columns...); err != nil {
		if err == sql.ErrNoRows {
			return
		} else {
			logger.Fatalf("Can't parse SQL-result: %s\n", err.Error())
		}
	}
	job.params = make([]string, len(config.MySql.Params))
	for i := range config.MySql.Params {
		b := *(columns[i+1].(*[]byte))
		if !config.MySql.compiledParamsFilter[i].Match(b) {
			logger.Printf("Job #%d SQL-data filtered: params #%d (0-based index) doesn't match filter\n", job.id, i)
			job.SetDone(JOB_FILTERED)
			return
		} else {
			job.params[i] = string(b)
		}

	}

	logger.Printf("Found new job #%d, params: %v\n", job.id, job.params)

	select {
	case processJob <- &job:
	case <-instance.quit:
		instance.quit <- true
		return
	}

	instance.lastId = job.id
	instance.forceNext <- true
}

func (instance *queueReader) loop() {
	defer instance.stm.Close()
	defer threads.Done()

	instance.forceNext <- true

loop:
	for {
		select {
		case <-instance.quit:
			break loop
		case <-instance.forceNext:
			instance.processJob()
		case <-time.After(5 * time.Second):
			instance.processJob()
		}
	}
}

func runQueueReader() {
	quit := make(chan bool, 1)
	quit_chans = append(quit_chans, quit)

	instance := newQueueReader(quit)
	go instance.loop()

	threads.Add(1)
}
