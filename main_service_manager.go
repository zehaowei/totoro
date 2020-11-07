package totoro

import (
	"totoro/util"
)

type MainAppManager struct {
	imageName 		string
	appName 		string
	containerId 	string

	// resources config
	cpuSets			string
}

func MakeMainAppManager(imageName string, appName string) *MainAppManager {
	mam := new (MainAppManager)
	mam.appName = appName
	mam.imageName = imageName

	return mam
}

func (mam *MainAppManager) LaunchMainApp() {
	util.PrintInfo("[info] launch Main App")
	//var portMap = make(nat.PortMap)
	//portMap["11211/tcp"] = []nat.PortBinding{
	//	{HostPort: "11211"},
	//}
	//config := &container.Config{
	//	Image: mam.imageName,
	//}
	//config.ExposedPorts["11211/tcp"] = struct{}{}
	//CreateContainerByImageName(mam.appName, config, &container.HostConfig{
	//	PortBindings: portMap,
	//})

	mam.containerId = GetContainerIdByName("/"+mam.appName)
	UpdateContainerCpuQuotaById(mam.containerId, 50000)
}

func (mam *MainAppManager) GetResourceInfo() (float64, float64){
	util.PrintInfo("[info] ----------------  Main App Info  ----------------")
	return GetAppResourceInfo(mam.containerId)
}
