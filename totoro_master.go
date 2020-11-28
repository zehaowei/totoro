package totoro

import (
	"fmt"
	"time"
)

const MonitorInterval = 2500 * time.Millisecond
const HeartbeatInterval = 1000 * time.Millisecond

type Totoro struct {
	mainAppManager 		*MainAppManager
	serverlessManager 	*ServerlessManager
	policyEngine		*PolicyEngine
	shutdown			chan struct{}
}

func MakeTotoro() *Totoro {
	totoro := new(Totoro)
	totoro.mainAppManager = MakeMainAppManager("zehwei/memcached", "memcached")
	totoro.serverlessManager = MakeServerlessManager()
	totoro.policyEngine = MakePolicyEngine(totoro)
	totoro.shutdown = make(chan struct{})

	return totoro
}

func (ttr *Totoro) Start(notify chan struct{}) {
	ttr.mainAppManager.LaunchMainApp()
	go ttr.monitorMainApp()
	//go ttr.monitorCpuUsage()
}

func (ttr *Totoro) monitorMainApp() {
	timer := time.NewTimer(MonitorInterval)
	for {
		select {
			case <- timer.C:
				go ttr.triggerPolicy()
				timer.Reset(MonitorInterval)
			case <- ttr.shutdown:
				return
		}
	}
}

func (ttr *Totoro) triggerPolicy() {
	cpuUsage, _ := ttr.mainAppManager.GetResourceInfo()
	//ttr.policyEngine.PolicyWithoutTask(cpuUsage)
	ttr.policyEngine.SimplePolicy(cpuUsage)
}

func (ttr *Totoro) monitorCpuUsage() {
	timer := time.NewTimer(HeartbeatInterval)
	for {
		select {
		case <- timer.C:
			go ttr.getResourceInfo()
			timer.Reset(HeartbeatInterval)
		case <- ttr.shutdown:
			return
		}
	}
}

func (ttr *Totoro) getResourceInfo() {
	cpuUsage, _ := ttr.mainAppManager.GetResourceInfo()
	fmt.Printf("%f %v\n", cpuUsage, time.Now().Unix())
}
