package main

import (
	"fmt"
)

const (
	JOB_FILTERED = -1
	JOB_TODO     = 0
	JOB_DONE     = 1
	JOB_ERROR    = 2
	JOB_FATAL    = 3
)

type job struct {
	id     int32
	params []string
}

func (instance job) SetDone(rc int) {
	query := fmt.Sprintf("UPDATE `%s` SET done = ? WHERE id = ?", config.MySql.Table)
	_, err := db.Exec(query, rc, instance.id)
	if err != nil {
		logger.Printf("Can't update job status in database: %s\n", err.Error())
	}
}
