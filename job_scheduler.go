package totoro

import (
	"github.com/Workiva/go-datastructures/queue"
	"time"
	"totoro/util"
)

/*
 * receive requests from clients and send requests of tasks to TaskScheduler
 */
type JobScheduler struct {
	jobRunning			map[int]Job				// jobs are running
	jobWaitQueue 		*queue.Queue  			// jobs wait to be executed
	idGenerator			*util.UUIDGenerator		// generate unique id for a job
	taskScheduler		*TaskScheduler			// scheduler to dispatch a specific task
	jobNotify			chan struct{}			// notify goroutine to deal with new arriving jobs
	shutdown            chan struct{}			// do some cleaning when exit
}

/*
 * constructor
 */
func MakeJobScheduler(ts *TaskScheduler) *JobScheduler {
	js := new(JobScheduler)
	js.jobRunning = make(map[int]Job, 0)
	js.jobWaitQueue = new(queue.Queue)
	js.idGenerator = util.MakeUUIDGenerator(JobIdPrefix)
	js.taskScheduler = ts
	js.jobNotify = make(chan struct{}, 8)
	js.shutdown = make(chan struct{})
	go js.monitorJobs()
	return js
}

/*
 * receive requests from simulator
 */
func (js *JobScheduler) ReceiveRPC(job *Job) {
	job.JobId = js.idGenerator.Get()
	err := js.jobWaitQueue.Put(job)
	if err != nil {
		util.PrintErr("[error] queue put error")
	}
	js.jobNotify <- struct{}{}
}

/*
 * monitor the job waiting queue to launch new jobs
 */
func (js *JobScheduler) monitorJobs() {
	for {
		select {
		case <- js.jobNotify: {
			if !js.jobWaitQueue.Empty() {
				jobs, _ := js.jobWaitQueue.Poll(js.jobWaitQueue.Len(), 5)
				for _, j := range jobs {
					go js.executeJob(j.(*Job))
				}
			}
		}
		case <- js.shutdown:
			return
		}
	}
}

/*
 * execute a job
 */
func (js *JobScheduler) executeJob(job *Job) {
	for _, t := range job.Tasks {
		// push a task to TaskScheduler
		js.taskScheduler.ReceiveTask(t)
		// wait for the task
		<-t.NotifyFinish
	}
	job.End = time.Now()
}