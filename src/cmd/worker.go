package main

import (
	"bytes"
	"os/exec"
	"strings"
	"time"
)

type worker struct {
	quit chan bool
}

func newWorker() *worker {
	w := &worker{}
	w.quit = make(chan bool, 1)
	return w
}

func (instance *worker) run(queue *jobQueue, job *job) {
	select {
	case <-instance.quit:
		return
	default:
	}

	threads.Add(1)
	go func() {
		defer threads.Done()

		argv := strings.Split(job.path, " ")
		cmd := exec.Command(config.FFMpeg.Path, argv...)
		var output bytes.Buffer
		cmd.Stderr = &output
		cmd.Stdout = &output

		err := cmd.Start()
		if err != nil {
			logger.Printf("Can't start FFMpeg for job #%d: %s\n", job.id, err.Error())
			job.SetDone(3)
			queue.freeWorker <- instance
			return
		}

		process_complete := make(chan bool)
		go func() {
			cmd.Wait()
			process_complete <- true
		}()

		select {
		case <-process_complete:
			logger.Printf("FFMpeg process complete for job #%d, success = %v\n", job.id, cmd.ProcessState.Success())
			if cmd.ProcessState.Success() {
				job.SetDone(1)
			} else {
				logger.Println(output.String())
				job.SetDone(2)
			}
			queue.freeWorker <- instance
		case <-time.After(time.Duration(config.FFMpeg.Timeout) * time.Second):
			logger.Printf("FFMpeg process killed by timeout for job #%d\n", job.id)
			cmd.Process.Kill()
			job.SetDone(3)
			queue.freeWorker <- instance
		case <-instance.quit:
			cmd.Process.Kill()
			return
		}
	}()
}
