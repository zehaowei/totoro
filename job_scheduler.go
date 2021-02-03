package totoro

import (
	"github.com/Workiva/go-datastructures/queue"
	"os"
	"strconv"
	"sync"
	"time"
	"totoro/util"
)

/*
 * receive requests from clients and send requests of tasks to TaskScheduler
 */
type JobScheduler struct {
	jobRunning			map[int]*Job			// jobs are running
	jobWaitQueue 		*queue.Queue  			// jobs wait to be executed
	jobFinished			*queue.Queue 			// jobs are finished
	idGenerator			*util.UUIDGenerator		// generate unique id for a job
	taskScheduler		*TaskScheduler			// scheduler to dispatch a specific task
	jobNotify			chan struct{}			// notify goroutine to deal with new arriving jobs
	jobInfoFile			*os.File				// file to store information of job execution
	shutdown            chan struct{}			// do some cleaning when exit
	mu					sync.Mutex
}

/*
 * constructor
 */
func MakeJobScheduler(ts *TaskScheduler) *JobScheduler {
	js := new(JobScheduler)
	js.jobRunning = make(map[int]*Job, 0)
	js.jobWaitQueue = new(queue.Queue)
	js.jobFinished = new(queue.Queue)
	js.idGenerator = util.MakeUUIDGenerator(JobIdPrefix)
	js.taskScheduler = ts
	js.jobNotify = make(chan struct{}, 8)
	js.shutdown = make(chan struct{})

	if util.CheckFileIsExist(JobInfoFile) {
		err := os.Remove(JobInfoFile)
		if err != nil { util.PrintErr("[error] file remove error") }
	}
	js.jobInfoFile, _ = os.Create(JobInfoFile)

	go js.monitorJobs()
	return js
}

/*
 * receive requests from simulator
 */
func (js *JobScheduler) ReceiveRPC(job *Job) {
	job.JobId = js.idGenerator.Get()
	err := js.jobWaitQueue.Put(job)
	if err != nil { util.PrintErr("[error] queue put error") }
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
			err := js.jobInfoFile.Close()
			if err != nil { util.PrintErr("[error] file close error") }
			return
		}
	}
}

/*
 * execute a job
 */
func (js *JobScheduler) executeJob(job *Job) {
	util.PrintInfo("[info] -- start job: %s --", job.JobName)
	for _, t := range job.Tasks {
		// push a task to TaskScheduler
		js.taskScheduler.ReceiveTask(t)
		// wait for the task
		<-t.NotifyFinish
	}
	job.End = time.Now()
	js.mu.Lock()
	err := js.jobFinished.Put(job)
	if err != nil { util.PrintErr("[error] jobFinished queue put error") }
	js.mu.Unlock()
	js.outputJobInfo(job)
	util.PrintInfo("[info] -- finish job: %s --", job.JobName)
}

func (js *JobScheduler) outputJobInfo(job *Job) {
	start := job.Start.Format(TimeLayoutStr)
	end := job.End.Format(TimeLayoutStr)
	duration := strconv.FormatFloat(float64(job.End.Sub(job.Start))/1e9, 'f', 3, 64)
	content := []byte("JobName: "+job.JobName+"\n"+"StartTime: "+start+"\n"+"EndTime: "+end+"\n"+"Duration: "+duration+"\n")
	_, err := js.jobInfoFile.Write(content)
	if err != nil { util.PrintErr("[error] file write error") }
	for _, task := range job.Tasks {
		content2 := []byte("TaskName: "+task.Name+"\n"+"ExeTimes: "+strconv.Itoa(task.ExecuteTimes)+"\n"+"Type: "+task.Type+"\n")
		_, err = js.jobInfoFile.Write(content2)
	}
	_, err = js.jobInfoFile.Write([]byte("\n"))
}