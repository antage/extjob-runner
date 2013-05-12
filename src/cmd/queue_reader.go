package main

import (
	"database/sql"
	"fmt"
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

	instance.stm, err = db.Prepare(fmt.Sprintf("SELECT id, path FROM `%s` WHERE done = 0 AND id > ? ORDER BY `id` LIMIT 1", config.MySql.Table))
	if err != nil {
		logger.Fatalf("Can't prepare SQL-statement: %s\n", err.Error())
	}

	return instance
}

func (instance *queueReader) processJob() {
	var job job
	var pathb []byte
	if err := instance.stm.QueryRow(instance.lastId).Scan(&job.id, &pathb); err != nil {
		if err == sql.ErrNoRows {
			return
		} else {
			logger.Fatalf("Can't parse SQL-result: %s\n", err.Error())
		}
	}
	job.path = string(pathb)

	logger.Printf("Found new job #%d, path: %s\n", job.id, job.path)

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
