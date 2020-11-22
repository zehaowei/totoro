package totoro

import (
	"time"
)

const MonitorInterval = 2500 * time.Millisecond

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
}

func (ttr *Totoro) monitorMainApp() {
	timer := time.NewTimer(MonitorInterval)
	for {
		select {
			case <- timer.C:
				go ttr.collectResourceInfo()
				timer.Reset(MonitorInterval)
			case <- ttr.shutdown:
				return
		}
	}
}

func (ttr *Totoro) collectResourceInfo() {
	cpuUsage, _ := ttr.mainAppManager.GetResourceInfo()
	//ttr.policyEngine.PolicyWithoutTask(cpuUsage)
	ttr.policyEngine.SimplePolicy(cpuUsage)
}
