package main

import (
	"fmt"
)

type job struct {
	id   int32
	path string
}

func (instance job) SetDone(rc int) {
	query := fmt.Sprintf("UPDATE `%s` SET done = ? WHERE id = ?", config.MySql.Table)
	_, err := db.Exec(query, rc, instance.id)
	if err != nil {
		logger.Printf("Can't update job status in database: %s\n", err.Error())
	}
}
