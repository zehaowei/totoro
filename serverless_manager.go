package totoro

import (
	"github.com/Workiva/go-datastructures/queue"
	"github.com/docker/docker/api/types/container"
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
	tasksRunning		map[int][]Task
	taskContinueQueue 	*queue.Queue
	taskWaitQueue 		*queue.Queue
	mu					sync.Mutex
}

type Task struct {
	name 			string
	imageName		string
	containerId 	string
	status 			int

	cpuSet			int
}

func MakeServerlessManager() *ServerlessManager {
	serm := new(ServerlessManager)
	serm.tasksRunning = make(map[int][]Task, 4)
	for i := 0; i <= 3; i++ {
		serm.tasksRunning[i] = make([]Task, 0)
	}
	serm.taskContinueQueue = new(queue.Queue)
	serm.taskWaitQueue = new(queue.Queue)

	// just for test
	serm.taskWaitQueue.Put(Task{
		name: "java_hello1",
		imageName: "zehwei/hello_java",
		status: WAITING,
	})
	serm.taskWaitQueue.Put(Task{
		name: "sort1",
		imageName: "zehwei/sort:1.0",
		status: WAITING,
	})
	serm.taskWaitQueue.Put(Task{
		name: "sort2",
		imageName: "zehwei/sort:1.0",
		status: WAITING,
	})
	serm.taskWaitQueue.Put(Task{
		name: "sort3",
		imageName: "zehwei/sort:1.0",
		status: WAITING,
	})
	serm.taskWaitQueue.Put(Task{
		name: "sort4",
		imageName: "zehwei/sort:1.0",
		status: WAITING,
	})

	return serm
}

func (slm *ServerlessManager) GetResourceInfo(cpuInd int, taskInd int, status int) {
	util.PrintInfo("[info] ----------------  Serverless App Info  ----------------")
	switch status {
	case RUNNING:
		GetAppResourceInfo(slm.tasksRunning[cpuInd][taskInd].containerId)
	default:
		util.PrintInfo("[info] This task is not running")
	}

}

// start a new task from waiting queue
func (slm *ServerlessManager) StartTask(cpu int) {
	slm.mu.Lock()
	defer slm.mu.Unlock()
	items, _ := slm.taskWaitQueue.Poll(1, 1)
	task, _ := items[0].(Task)
	task.cpuSet = cpu
	task.containerId = slm.launchTask(task.imageName, task.name, task.cpuSet)
	slm.tasksRunning[cpu] = append(slm.tasksRunning[cpu], task)
	util.PrintInfo("[info] ----------  start serverless task: %s | cpu (%d) ----------", task.name, cpu)
}

// start the containers stopped before from queue
func (slm *ServerlessManager) ContinueTask(cpu int) {
	slm.mu.Lock()
	defer slm.mu.Unlock()
	items, _ := slm.taskContinueQueue.Poll(1, 1)
	task, _ := items[0].(Task)
	task.cpuSet = cpu
	slm.tasksRunning[cpu] = append(slm.tasksRunning[cpu], task)
	UpdateContainerCpuSetsById(task.containerId, strconv.Itoa(cpu))
	StartContainerById(task.containerId)
	util.PrintInfo("[info] ----------  continue serverless task: %s | cpu (%d) ----------", task.name, cpu)
}

func (slm *ServerlessManager) UpdateTask() {

}

func (slm *ServerlessManager) StopTask(cpuInd int) {
	slm.mu.Lock()
	defer slm.mu.Unlock()
	for _, task := range slm.tasksRunning[cpuInd] {
		StopContainerById(task.containerId, 2*time.Second)
		task.status = STOPPED
		slm.taskContinueQueue.Put(task)
		util.PrintInfo("[info] ----------  stop serverless task: %s | cpu (%d) ------------", task.name, cpuInd)
	}
	slm.tasksRunning[cpuInd] = make([]Task, 0)
}

func (slm *ServerlessManager) launchTask(imageName string, containerName string, cpu int) string {
	util.PrintInfo("[info] ----------------  launch serverless task  ----------------")
	id := CreateContainerByImageName(containerName, &container.Config{
		Image: imageName,
	}, &container.HostConfig{
		Resources: container.Resources{
			CpusetCpus: strconv.Itoa(cpu),
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
