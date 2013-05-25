package main

import (
	"bytes"
	"os/exec"
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

		var command bytes.Buffer
		err := config.Command.compiledTemplate.Execute(&command, job.params)

		logger.Printf("Start command: %s\n", command.String())
		cmd := exec.Command(config.Command.Shell, "-c", command.String())
		var output bytes.Buffer
		cmd.Stderr = &output
		cmd.Stdout = &output

		err = cmd.Start()
		if err != nil {
			logger.Printf("Can't start command for job #%d: %s\n", job.id, err.Error())
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
			logger.Printf("Command process complete for job #%d, success = %v\n", job.id, cmd.ProcessState.Success())
			if cmd.ProcessState.Success() {
				job.SetDone(1)
			} else {
				logger.Println(output.String())
				job.SetDone(2)
			}
			queue.freeWorker <- instance
		case <-time.After(time.Duration(config.Command.Timeout) * time.Second):
			logger.Printf("Command process killed by timeout for job #%d\n", job.id)
			cmd.Process.Kill()
			job.SetDone(3)
			queue.freeWorker <- instance
		case <-instance.quit:
			cmd.Process.Kill()
			return
		}
	}()
}
