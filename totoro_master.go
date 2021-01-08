package totoro

import (
	"fmt"
	"time"
)

/*
 * the master of the entire system
 */
type Totoro struct {
	mainAppManager 		*MainAppManager  	// control memcached
	jobScheduler   		*JobScheduler    	// job scheduling, receive requests from client, push tasks to TaskScheduler
	taskScheduler  		*TaskScheduler   	// task scheduling, responsible for task container management
	policyEngine   		*PolicyEngine		// send instructions to MainAppManager and TaskScheduler
	requestSimulator 	*RequestSimulator	// simulate clients to send serverless job to Totoro
	shutdown       		chan struct{}
}

/*
 * constructor
 */
func MakeTotoro() *Totoro {
	totoro := new(Totoro)
	totoro.mainAppManager = MakeMainAppManager("zehwei/memcached", "memcached")
	totoro.policyEngine = MakePolicyEngine(totoro)
	totoro.taskScheduler = MakeTaskScheduler()
	totoro.jobScheduler = MakeJobScheduler(totoro.taskScheduler)
	totoro.requestSimulator = MakeRequestSimulator(totoro.jobScheduler)
	totoro.shutdown = make(chan struct{})
	return totoro
}

/*
 * start the system
 */
func (ttr *Totoro) Start(notify chan struct{}) {
	ttr.mainAppManager.LaunchMainApp()
	go ttr.monitorMainApp()
	//go ttr.monitorCpuUsage()
}

/*
 * monitor the situation of memcached and periodically trigger policy checking
 */
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

/*
 * trigger a specific policy to determine the resources allocation
 */
func (ttr *Totoro) triggerPolicy() {
	cpuUsage, _ := ttr.mainAppManager.GetResourceInfo()
	//ttr.policyEngine.PolicyWithoutTask(cpuUsage)
	//ttr.policyEngine.SimplePolicy(cpuUsage)
	ttr.policyEngine.SecondPolicyWithoutTask(cpuUsage)
}

/*
 * For experiment
 * Monitor cpu usage every second to determine the relationship between latency and cpu usage
 */
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
