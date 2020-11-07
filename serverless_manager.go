package totoro

import (
	"time"
	"totoro/util"
	"github.com/docker/docker/api/types/container"
)

const (
	RUNNING 		int = 0
	STOPPED 		int = 1
	COMPLETE 		int = 2
)

type ServerlessManager struct {
	tasks	[]string
	status  []int
}

func MakeServerlessManager() *ServerlessManager {
	serm := new (ServerlessManager)
	serm.tasks = make([]string, 1)
	serm.status = make([]int, 1)
	serm.tasks[0] = "non-exist"
	return serm
}

func (slm *ServerlessManager) GetResourceInfo() {
	util.PrintInfo("[info] ----------------  Serverless App Info  ----------------")
	GetAppResourceInfo(slm.tasks[0])
}

func (slm *ServerlessManager) RunningTask(imageName string, taskName string, cpu string) {
	if slm.tasks[0] == "non-exist" {
		slm.launchTask(imageName, taskName, cpu)
	} else if slm.status[0] == STOPPED {
		util.PrintInfo("[info] restart serverless app")
		StartContainerById(slm.tasks[0])
	}
}

func (slm *ServerlessManager) launchTask(imageName string, containerName string, cpu string) {
	util.PrintInfo("[info] launch serverless task")
	CreateContainerByImageName(containerName, &container.Config{
		Image: imageName,
	}, &container.HostConfig{
		Resources: container.Resources{
			CpusetCpus: cpu,
		},
	})
	slm.tasks[0] = GetContainerIdByName("/"+containerName)
	slm.status[0] = RUNNING
}

func (slm *ServerlessManager) UpdateTask() {
	if slm.status[0] == RUNNING {
		util.PrintInfo("[info] current serverless app info")
		GetAppResourceInfo(slm.tasks[0])
		util.PrintInfo("[info] update serverless app cpu share to 25")
	}
}

func (slm *ServerlessManager) StopTask() {
	if slm.status[0] == RUNNING {
		util.PrintInfo("[info] current serverless app info")
		GetAppResourceInfo(slm.tasks[0])
		util.PrintInfo("[info] stop serverless task")
		StopContainerById(slm.tasks[0], 5*time.Second)
	}
}
