package totoro

import (
	"github.com/Workiva/go-datastructures/queue"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/strslice"
	"strconv"
	"sync"
	"time"
	"totoro/util"
)

const (
	RUNNING 		int = 0
	STOPPED 		int = 1
	COMPLETE 		int = 2
	WAITING			int = 3
)

type ServerlessManager struct {
	tasksRunning		map[int][]Task		// tasks are running
	taskContinueQueue 	*queue.Queue		// tasks wait to be restarted
	taskWaitQueue 		*queue.Queue		// tasks wait to be executed
	taskLoader			*TaskLoader
	mu					sync.Mutex

	cpuTasksLimits		map[int]int			// maximal number of tasks in each cpu core
}

type Task struct {
	Name 			string
	ImageName		string
	ContainerId 	string
	Cmd  			strslice.StrSlice
	Status 			int

	CpuSet			int
}

func MakeServerlessManager() *ServerlessManager {
	serm := new(ServerlessManager)
	serm.tasksRunning = make(map[int][]Task, CoreNums)
	for i := 0; i < CoreNums; i++ {
		serm.tasksRunning[i] = make([]Task, 0)
	}
	serm.taskContinueQueue = new(queue.Queue)
	serm.taskWaitQueue = new(queue.Queue)

	serm.cpuTasksLimits = make(map[int]int, CoreNums)
	for i := 0; i < CoreNums; i++ {
		serm.cpuTasksLimits[i] = 1
	}

	serm.taskLoader = MakeTaskLoader()
	tasks := serm.taskLoader.loadTasks(CPU)
	for _, tsk := range tasks {
		err := serm.taskWaitQueue.Put(tsk)
		if err != nil {}
	}

	return serm
}

func (slm *ServerlessManager) GetResourceInfo(cpuInd int, taskInd int, status int) {
	util.PrintInfo("[info] ----------------  Serverless App Info  ----------------")
	switch status {
	case RUNNING:
		GetAppResourceInfo(slm.tasksRunning[cpuInd][taskInd].ContainerId)
	default:
		util.PrintInfo("[info] This task is not running")
	}

}

// start a new task from waiting queue
func (slm *ServerlessManager) StartTask(cpu int) {
	slm.mu.Lock()
	defer slm.mu.Unlock()
	if len(slm.tasksRunning[cpu]) == slm.cpuTasksLimits[cpu] {
		return
	}
	items, _ := slm.taskWaitQueue.Poll(1, 1)
	task, _ := items[0].(Task)
	task.CpuSet = cpu
	task.ContainerId = slm.launchTask(&task)
	slm.tasksRunning[cpu] = append(slm.tasksRunning[cpu], task)
	util.PrintInfo("[info] ----------  start serverless task: %s | cpu (%d) ---------- %v", task.Name, cpu, time.Now().Unix())
}

// start the containers stopped before from queue
func (slm *ServerlessManager) ContinueTask(cpu int) {
	slm.mu.Lock()
	defer slm.mu.Unlock()
	if len(slm.tasksRunning[cpu]) == slm.cpuTasksLimits[cpu] {
		return
	}
	items, _ := slm.taskContinueQueue.Poll(1, 1)
	task, _ := items[0].(Task)
	task.CpuSet = cpu
	UpdateContainerCpuSetsById(task.ContainerId, strconv.Itoa(cpu))
	StartContainerById(task.ContainerId)
	slm.tasksRunning[cpu] = append(slm.tasksRunning[cpu], task)
	util.PrintInfo("[info] ----------  continue serverless task: %s | cpu (%d) ---------- %v", task.Name, cpu, time.Now().Unix())
}

func (slm *ServerlessManager) UpdateTask() {

}

func (slm *ServerlessManager) StopTask(cpuInd int) {
	slm.mu.Lock()
	defer slm.mu.Unlock()
	for _, task := range slm.tasksRunning[cpuInd] {
		StopContainerById(task.ContainerId, 2*time.Second)
		task.Status = STOPPED
		slm.taskContinueQueue.Put(task)
		util.PrintInfo("[info] ----------  stop serverless task: %s | cpu (%d) ------------ %v", task.Name, cpuInd, time.Now().Unix())
	}
	slm.tasksRunning[cpuInd] = make([]Task, 0)
}

func (slm *ServerlessManager) KillTask(cpuInd int) {
	slm.mu.Lock()
	defer slm.mu.Unlock()
	for _, task := range slm.tasksRunning[cpuInd] {
		KillContainerById(task.ContainerId)
		task.Status = STOPPED
		slm.taskContinueQueue.Put(task)
		util.PrintInfo("[info] ----------  kill serverless task: %s | cpu (%d) ------------ %v", task.Name, cpuInd, time.Now().Unix())
	}
	slm.tasksRunning[cpuInd] = make([]Task, 0)
}

func (slm *ServerlessManager) launchTask(task *Task) string {
	// util.PrintInfo("[info] ----------------  launch serverless task  ----------------")
	id := CreateContainerByImageName(task.Name, &container.Config{
		Image: task.ImageName,
		Cmd: task.Cmd,
	}, &container.HostConfig{
		Resources: container.Resources{
			CpusetCpus: strconv.Itoa(task.CpuSet),
		},
	})
	return id
}

func (slm *ServerlessManager) GetWaitTasks() *queue.Queue {
	return slm.taskWaitQueue
}

func (slm *ServerlessManager) GetRunningTasks() map[int][]Task {
	return slm.tasksRunning
}

func (slm *ServerlessManager) HasWaitTasks() bool {
	return !slm.taskWaitQueue.Empty()
}

func (slm *ServerlessManager) HasContinueTasks() bool {
	return !slm.taskContinueQueue.Empty()
}
