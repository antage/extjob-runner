package main

type jobQueue struct {
	queue      chan *job
	quit       chan bool
	workers    []*worker
	freeWorker chan *worker
}

var processJob chan<- *job

func newJobQueue(quit chan bool) *jobQueue {
	q := &jobQueue{}
	q.queue = make(chan *job)
	q.quit = quit

	q.freeWorker = make(chan *worker, config.Command.Workers)
	q.workers = make([]*worker, config.Command.Workers)
	for i := uint(0); i < config.Command.Workers; i++ {
		worker := newWorker()
		q.workers[i] = worker
		q.freeWorker <- worker
	}

	return q
}

func (instance *jobQueue) loop() {
	defer threads.Done()

loop:
	for {
		select {
		case <-instance.quit:
			break loop
		case job := <-instance.queue:
			logger.Printf("Process job #%d\n", job.id)
			select {
			case <-instance.quit:
				for _, worker := range instance.workers {
					worker.quit <- true
				}
				break loop
			case worker := <-instance.freeWorker:
				worker.run(instance, job)
			}
		}
	}
}

func runJobQueue() {
	quit := make(chan bool, 1)
	quit_chans = append(quit_chans, quit)

	jobQueue := newJobQueue(quit)
	processJob = jobQueue.queue
	go jobQueue.loop()

	threads.Add(1)
}
