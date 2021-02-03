package totoro

import (
	"fmt"
	"github.com/Workiva/go-datastructures/queue"
	"github.com/docker/docker/api/types/container"
	"strconv"
	"sync"
	"time"
	"totoro/util"
)

/*
 * schedule tasks and manage all tasks' containers
 */
type TaskScheduler struct {
	shortTaskWaitQueue 		*queue.Queue		// short tasks wait to be executed
	shortTaskContinueQueue 	*queue.Queue		// short tasks wait to be restarted
	shortTaskRunning		map[int][]*Task		// short tasks are running

	longTaskWaitQueue   	*queue.Queue  		// long tasks wait to be executed (this queue is useless according to current policy)
	longTaskContinueQueue 	*queue.Queue  		// long tasks wait to be restarted
	longTaskRunning     	map[int][]*Task 	// long tasks are running

	cpuTasksLimits			map[int]int			// maximal number of tasks in each cpu core
	cpuStatus				[]int				// show which physical cpu core is available
	idGenerator				*util.UUIDGenerator	// generate unique id for a task
	mu                		sync.Mutex
	shutdown            	chan struct{}		// do some cleaning when exit
	shutdown2 				chan struct{}
}

/*
 * constructor
 */
func MakeTaskScheduler() *TaskScheduler {
	ts := new(TaskScheduler)

	ts.shortTaskWaitQueue = new(queue.Queue)
	ts.shortTaskContinueQueue = new(queue.Queue)
	ts.shortTaskRunning = make(map[int][]*Task, 0)
	// short task could run on physical core 1 - 7 (0 is always for main app)
	for i := 1; i < CoreNums; i++ {
		ts.shortTaskRunning[i*CpuSetsNum] = make([]*Task, 0)
		ts.shortTaskRunning[i*CpuSetsNum+CpuIndexGap] = make([]*Task, 0)
	}

	ts.longTaskWaitQueue = new(queue.Queue)
	ts.longTaskContinueQueue = new(queue.Queue)
	ts.longTaskRunning = make(map[int][]*Task, 0)
	// long task only run on physical core 5 - 7 (less chance to be occupied by main app)
	for i := CoreNums/2+1; i < CoreNums; i++ {
		ts.longTaskRunning[i*CpuSetsNum] = make([]*Task, 0)
		ts.longTaskRunning[i*CpuSetsNum+CpuIndexGap] = make([]*Task, 0)
	}

	ts.cpuTasksLimits = make(map[int]int, 0)
	for i := 1; i < CoreNums; i++ {
		ts.cpuTasksLimits[i*CpuSetsNum] = TaskLimitPerCore
		ts.cpuTasksLimits[i*CpuSetsNum+CpuIndexGap] = TaskLimitPerCore
	}

	ts.cpuStatus = make([]int, CoreNums)
	ts.cpuStatus[0] = OCCUPIED
	for i := 1; i < CoreNums; i++ {
		ts.cpuStatus[i] = AVAILABLE
	}

	ts.idGenerator = util.MakeUUIDGenerator(TaskIdPrefix)
	ts.shutdown = make(chan struct{})
	ts.shutdown2 = make(chan struct{})

	go ts.monitorTasks()
	go ts.monitorTaskStatus()

	return ts
}

/*
 * receive task from JobScheduler
 */
func (ts *TaskScheduler) ReceiveTask(task *Task) {
	task.TaskId = ts.idGenerator.Get()
	task.ExecuteTimes = 0
	task.Status = WAITING
	task.Type = ShortTask
	err := ts.shortTaskWaitQueue.Put(task)
	if err != nil { util.PrintErr("[error] shortTaskWaitQueue put error") }
}

/*
 * start a task on a given cpu proc
 */
func (ts *TaskScheduler) launchTask(task *Task) {
	id := CreateContainerByImageName(task.Name, &container.Config{
		Image: task.ImageName,
		Cmd: task.Cmd,
	}, &container.HostConfig{
		Resources: container.Resources{
			CpusetCpus: strconv.Itoa(task.CpuSet),
		},
	})
	task.ContainerId = id
	task.Status = RUNNING
	util.PrintInfo("[info] -- launch task on cpu %d --", task.CpuSet)
}

/*
 * start the containers stopped before from queue
 */
func (ts *TaskScheduler) continueTask(task *Task) {
	UpdateContainerCpuSetsById(task.ContainerId, strconv.Itoa(task.CpuSet))
	StartContainerById(task.ContainerId)
	util.PrintInfo("[info] ----------  continue serverless task: %s | cpu (%d) ---------- %v", task.Name, task.CpuSet, time.Now().Unix())
}

/*
 * kill all the tasks on a physical core (two proc)
 */
func (ts *TaskScheduler) killTaskByCore(cpuInd int) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	for _, task := range ts.shortTaskRunning[cpuInd*CpuSetsNum] {
		ts.killTask(task)
	}
	for _, task := range ts.shortTaskRunning[cpuInd*CpuSetsNum+CpuIndexGap] {
		ts.killTask(task)
	}
	ts.shortTaskRunning[cpuInd*CpuSetsNum] = make([]*Task, 0)
	ts.shortTaskRunning[cpuInd*CpuSetsNum+CpuIndexGap] = make([]*Task, 0)

	if cpuInd >= CoreNums/2+1 {
		for _, task := range ts.longTaskRunning[cpuInd*CpuSetsNum] {
			ts.killTask(task)
		}
		for _, task := range ts.longTaskRunning[cpuInd*CpuSetsNum+CpuIndexGap] {
			ts.killTask(task)
		}
		ts.longTaskRunning[cpuInd*CpuSetsNum] = make([]*Task, 0)
		ts.longTaskRunning[cpuInd*CpuSetsNum+CpuIndexGap] = make([]*Task, 0)
	}
}

/*
 * kill one task
 */
func (ts *TaskScheduler) killTask(task *Task) {
	status, code := InspectContainerById(task.ContainerId)
	if status != "running" {
		// if task has finished, notify the client
		if status == "exited" && code == 0 {
			task.Status = COMPLETE
			task.NotifyFinish <- struct{}{}
			RemoveContainerById(task.ContainerId)
		}
		return
	}
	util.PrintInfo("[info] ----------  kill serverless task: %s | cpu (%d) ------------ %v", task.Name, task.CpuSet, time.Now().Unix())
	KillContainerById(task.ContainerId)
	task.Status = STOPPED
	task.ExecuteTimes += 1
	var err error
	// if the number of execution of a task exceeds a threshold, change its type to LongTask
	if task.Type == LongTask || task.ExecuteTimes >= ShortToLongThreshold {
		err = ts.longTaskContinueQueue.Put(task)
		task.Type = LongTask
	} else {
		err = ts.shortTaskContinueQueue.Put(task)
	}
	if err != nil { util.PrintErr("[error] killTask: queue put error") }
}

/*
 * when main app occupy more cores, TaskScheduler need to kill tasks on these cores
 */
func (ts *TaskScheduler) Deflate(cpuState []int) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	for i := 1; i < CoreNums; i++ {
		if cpuState[i] == OCCUPIED && ts.cpuStatus[i] == AVAILABLE {
			ts.cpuStatus[i] = OCCUPIED
			go ts.killTaskByCore(i)
		}
	}
}

/*
 * notify TaskScheduler that more cores are available
 */
func (ts *TaskScheduler) Inflate(cpuState []int) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	for i := 1; i < CoreNums; i++ {
		if cpuState[i] == AVAILABLE && ts.cpuStatus[i] == OCCUPIED {
			ts.cpuStatus[i] = AVAILABLE
		}
	}
}

/*
 * schedule tasks at regular intervals
 */
func (ts *TaskScheduler) monitorTasks() {
	timer := time.NewTimer(TaskSchedulingInterval)
	for {
		select {
		case <- timer.C:
			go ts.scheduleTasks()
			timer.Reset(TaskSchedulingInterval)
		case <- ts.shutdown:
			return
		}
	}
}

/*
 * launch tasks according to cpu status and load of main app
 */
func (ts *TaskScheduler) scheduleTasks() {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	cpuOccupiedNums := 0
	for _, status := range ts.cpuStatus {
		if status == OCCUPIED {
			cpuOccupiedNums++
		}
	}

	// low pressure, run both short and long term tasks
	if cpuOccupiedNums <= LowHighLoadThreshold {
		// Firstly run long term tasks on core 5-7 (long term task has failed to execute for several times)
		for i := CoreNums-1; i >= CoreNums/2+1; i-- {
			cpus := []int{i*CpuSetsNum, i*CpuSetsNum+CpuIndexGap}
			// check the two proc of physical core i
			for _, ind := range cpus {
				taskNum := len(ts.shortTaskRunning[ind]) + len(ts.longTaskRunning[ind])
				if ts.cpuStatus[i] == AVAILABLE && taskNum < ts.cpuTasksLimits[ind] && !ts.longTaskContinueQueue.Empty(){
					t, _ := ts.longTaskContinueQueue.Poll(1, 1)
					task := t[0].(*Task)
					task.CpuSet = ind
					ts.continueTask(task)
					ts.longTaskRunning[ind] = append(ts.shortTaskRunning[ind], task)
				}
			}
		}
	}
	// under high pressure, stop running long term tasks, only run short term tasks
	for i := CoreNums-1; i >= 1; i-- {
		cpus := []int{i*CpuSetsNum, i*CpuSetsNum+CpuIndexGap}
		for _, ind := range cpus {
			taskNum := len(ts.shortTaskRunning[ind])
			if i >= CoreNums/2+1 {
				taskNum += len(ts.longTaskRunning[ind])
			}
			if ts.cpuStatus[i] == AVAILABLE && taskNum < ts.cpuTasksLimits[ind] {
				var task *Task = nil
				// give priority to tasks in ContinueQueue
				if !ts.shortTaskContinueQueue.Empty() {
					t, _ := ts.shortTaskContinueQueue.Poll(1, 1)
					task = t[0].(*Task)
					task.CpuSet = ind
					ts.continueTask(task)
				} else if !ts.shortTaskWaitQueue.Empty() {
					t, _ := ts.shortTaskWaitQueue.Poll(1, 1)
					task = t[0].(*Task)
					task.CpuSet = ind
					ts.launchTask(task)
				}
				if task != nil {
					fmt.Printf("add task shortTaskRunning[%d] %s \n", ind, task.Name)
					ts.shortTaskRunning[ind] = append(ts.shortTaskRunning[ind], task)
				}
			}
		}
	}
}

/*
 * check task's status at regular interval
 */
func (ts *TaskScheduler) monitorTaskStatus() {
	timer := time.NewTimer(TaskStatusCheckInterval)
	for {
		select {
		case <- timer.C:
			go ts.checkTaskStatus()
			timer.Reset(TaskStatusCheckInterval)
		case <- ts.shutdown2:
			return
		}
	}
}

/*
 * check tasks in short and long term running list
 */
func (ts *TaskScheduler) checkTaskStatus() {
	ts.checkRunningList(ts.shortTaskRunning)
	ts.checkRunningList(ts.longTaskRunning)
}

/*
 * delete finished task from list and notify the client
 */
func (ts *TaskScheduler) checkRunningList(taskRunning map[int][]*Task) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	for proc, tasks := range taskRunning {
		if len(tasks) != 0 {
			offset := 0
			for ind, task := range tasks {
				status, code := InspectContainerById(task.ContainerId)
				if status == "exited" && code == 0 {
					task.Status = COMPLETE
					task.NotifyFinish <- struct{}{}
					taskRunning[proc] = append(taskRunning[proc][:ind-offset], taskRunning[proc][ind-offset+1:]...)
					offset++
					RemoveContainerById(task.ContainerId)
				} else if status == "" && code == 0 {
					taskRunning[proc] = append(taskRunning[proc][:ind-offset], taskRunning[proc][ind-offset+1:]...)
					offset++
				}
			}
		}
	}
}