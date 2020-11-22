package totoro

import (
	"sync"
	"time"
	"totoro/util"
)

type PolicyEngine struct {
	ttr 	*Totoro
	mu		sync.Mutex
}

func MakePolicyEngine(ttr *Totoro) *PolicyEngine {
	pe := new(PolicyEngine)
	pe.ttr = ttr

	return pe
}

func (pe *PolicyEngine) PolicyWithoutTask(cpuUsage float64) {
	pe.mu.Lock()
	defer pe.mu.Unlock()
	cpuNums := pe.ttr.mainAppManager.cpuNums
	if cpuUsage <= 50 {
		if cpuNums > 1 {
			util.PrintInfo("[time] --- trigger policy for main app (cpu: %f) --- %v", cpuUsage, time.Now().Unix())
			pe.ttr.mainAppManager.UpdateCpuSet("0")
			pe.ttr.mainAppManager.cpuNums = 1
			util.PrintInfo("[time] --- complete policy for main app --- %v", time.Now().Unix())
		}
	} else if cpuUsage > 70 && cpuUsage <= 140 {
		if cpuNums != 2 {
			util.PrintInfo("[time] --- trigger policy for main app (cpu: %f) --- %v", cpuUsage, time.Now().Unix())
			pe.ttr.mainAppManager.UpdateCpuSet("0-1")
			pe.ttr.mainAppManager.cpuNums = 2
			util.PrintInfo("[time] --- complete policy for main app --- %v", time.Now().Unix())
		}
	} else if cpuUsage > 150 && cpuUsage <= 240 {
		if cpuNums != 3 {
			util.PrintInfo("[time] --- trigger policy for main app (cpu: %f) --- %v", cpuUsage, time.Now().Unix())
			pe.ttr.mainAppManager.UpdateCpuSet("0-2")
			pe.ttr.mainAppManager.cpuNums = 3
			util.PrintInfo("[time] --- complete policy for main app --- %v", time.Now().Unix())
		}
	} else if cpuUsage > 250 {
		if cpuNums != 4 {
			util.PrintInfo("[time] --- trigger policy for main app (cpu: %f) --- %v", cpuUsage, time.Now().Unix())
			pe.ttr.mainAppManager.UpdateCpuSet("0-3")
			pe.ttr.mainAppManager.cpuNums = 4
			util.PrintInfo("[time] --- complete policy for main app --- %v", time.Now().Unix())
		}
	}
}

func (pe *PolicyEngine) SimplePolicy(cpuUsage float64) {
	pe.mu.Lock()
	defer pe.mu.Unlock()
	cpuNums := pe.ttr.mainAppManager.cpuNums
	if cpuUsage <= 50 {
		cpus := []bool{true, false, false, false}
		if cpuNums > 1 {
			util.PrintInfo("[time] --- trigger policy for main app (cpu: %f) --- %v", cpuUsage, time.Now().Unix())
			pe.ttr.mainAppManager.UpdateCpuSet("0")
			pe.ttr.mainAppManager.cpuNums = 1
			util.PrintInfo("[time] --- complete policy for main app --- %v", time.Now().Unix())
		} else {
			pe.startServerless(cpus)
		}
	} else if cpuUsage > 70 && cpuUsage <= 140 {
		cpus := []bool{true, true, false, false}
		if cpuNums != 2 {
			util.PrintInfo("[time] --- trigger policy for main app (cpu: %f) --- %v", cpuUsage, time.Now().Unix())
			pe.killServerless(cpus)
			pe.ttr.mainAppManager.UpdateCpuSet("0-1")
			pe.ttr.mainAppManager.cpuNums = 2
			util.PrintInfo("[time] --- complete policy for main app --- %v", time.Now().Unix())
		} else {
			pe.startServerless(cpus)
		}
	} else if cpuUsage > 150 && cpuUsage <= 240 {
		cpus := []bool{true, true, true, false}
		if cpuNums != 3 {
			util.PrintInfo("[time] --- trigger policy for main app (cpu: %f) --- %v", cpuUsage, time.Now().Unix())
			pe.killServerless(cpus)
			pe.ttr.mainAppManager.UpdateCpuSet("0-2")
			pe.ttr.mainAppManager.cpuNums = 3
			util.PrintInfo("[time] --- complete policy for main app --- %v", time.Now().Unix())
		} else {
			pe.startServerless(cpus)
		}
	} else if cpuUsage > 250 {
		cpus := []bool{true, true, true, true}
		if cpuNums != 4 {
			util.PrintInfo("[time] --- trigger policy for main app (cpu: %f) --- %v", cpuUsage, time.Now().Unix())
			pe.killServerless(cpus)
			pe.ttr.mainAppManager.UpdateCpuSet("0-3")
			pe.ttr.mainAppManager.cpuNums = 4
			util.PrintInfo("[time] --- complete policy for main app --- %v", time.Now().Unix())
		} else {
			pe.startServerless(cpus)
		}
	}
}

func (pe* PolicyEngine) killServerless(cpus []bool) {
	for i, occupied := range cpus {
		if occupied {
			pe.ttr.serverlessManager.KillTask(i)
		}
	}
}

func (pe* PolicyEngine) startServerless(cpus []bool) {
	for i, occupied := range cpus {
		if !occupied {
			if pe.ttr.serverlessManager.HasContinueTasks() {
				pe.ttr.serverlessManager.ContinueTask(i)
			} else if pe.ttr.serverlessManager.HasWaitTasks() {
				pe.ttr.serverlessManager.StartTask(i)
			}
		}
	}
}
