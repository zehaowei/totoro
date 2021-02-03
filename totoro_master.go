package totoro

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"time"
	"totoro/util"
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
	cpuUsageFile		*os.File			// file to store information of cpu usage
	exitChan			chan os.Signal		// capture exit signal from OS
	shutdown       		chan struct{}		// shutdown policy
	shutdown2			chan struct{}		// shutdown cpu monitor
	jobTraceType		string              // type of job trace
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
	totoro.exitChan = make(chan os.Signal)
	signal.Notify(totoro.exitChan, os.Interrupt, os.Kill)
	totoro.shutdown = make(chan struct{})
	totoro.shutdown2 = make(chan struct{})

	if util.CheckFileIsExist(CpuInfoFile) {
		err := os.Remove(CpuInfoFile)
		if err != nil { util.PrintErr("[error] file remove error") }
	}
	totoro.cpuUsageFile, _ = os.Create(CpuInfoFile)

	totoro.jobTraceType = JobTracePathHighLoad
	go totoro.monitorSignal()
	return totoro
}

/*
 * start the system
 */
func (ttr *Totoro) Start() {
	ttr.mainAppManager.LaunchMainApp()
	go ttr.monitorMainApp()
	go ttr.monitorCpuUsageMachine()
	//go ttr.monitorCpuUsage()
	ttr.requestSimulator.ReadJobs(ttr.jobTraceType)
}

/*
 * monitor exit signal from OS to do necessary cleaning
 */
func (ttr *Totoro) monitorSignal() {
	select {
	case <- ttr.exitChan:
		ttr.shutdown <- struct{}{}
		ttr.shutdown2 <- struct{}{}
		ttr.taskScheduler.shutdown <- struct{}{}
		ttr.taskScheduler.shutdown2 <- struct{}{}
		ttr.jobScheduler.shutdown <- struct{}{}
		os.Exit(0)
	}
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
	//ttr.policyEngine.SecondPolicyWithoutTask(cpuUsage)
	ttr.policyEngine.SecondPolicy(cpuUsage)
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

/*
 * Monitor cpu usage of the whole machine
 */
func (ttr *Totoro) monitorCpuUsageMachine() {
	for {
		select {
		case <- ttr.shutdown2:
			err := ttr.cpuUsageFile.Close()
			if err != nil { util.PrintErr("[error] file close error") }
			return
		default:
			cpuUsage := util.GetMachineCpuUsage()
			content := []byte(strconv.FormatFloat(cpuUsage, 'f', 3, 64)+" "+strconv.FormatInt(time.Now().Unix(), 10)+"\n")
			_, err := ttr.cpuUsageFile.Write(content)
			if err != nil { util.PrintErr("[error] file write error") }
		}
	}
}
