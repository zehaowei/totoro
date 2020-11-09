package totoro

import "sync"

type PolicyEngine struct {
	ttr 	*Totoro
	mu		sync.Mutex
}

func MakePolicyEngine(ttr *Totoro) *PolicyEngine {
	pe := new(PolicyEngine)
	pe.ttr = ttr

	return pe
}

func (pe *PolicyEngine) SimplePolicy(cpuUsage float64) {
	cpuNums := pe.ttr.mainAppManager.cpuNums
	if cpuUsage <= 50 {
		if cpuNums > 1 {
			pe.ttr.mainAppManager.UpdateCpuSet("0")
			pe.ttr.mainAppManager.cpuNums = 1
		}
		cpus := []bool{true, false, false, false}
		pe.dealWithServerless(cpus)
	} else if cpuUsage > 75 && cpuUsage <= 150 {
		if cpuNums != 2 {
			pe.ttr.mainAppManager.UpdateCpuSet("0-1")
			pe.ttr.mainAppManager.cpuNums = 2
		}
		cpus := []bool{true, true, false, false}
		pe.dealWithServerless(cpus)
	} else if cpuUsage > 160 && cpuUsage < 240 {
		if cpuNums != 3 {
			pe.ttr.mainAppManager.UpdateCpuSet("0-2")
			pe.ttr.mainAppManager.cpuNums = 3
		}
		cpus := []bool{true, true, true, false}
		pe.dealWithServerless(cpus)
	} else if cpuUsage > 260 {
		if cpuNums != 4 {
			pe.ttr.mainAppManager.UpdateCpuSet("0-3")
			pe.ttr.mainAppManager.cpuNums = 4
		}
		cpus := []bool{true, true, true, true}
		pe.dealWithServerless(cpus)
	}
}

func (pe* PolicyEngine) dealWithServerless(cpus []bool) {
	pe.mu.Lock()
	defer pe.mu.Unlock()
	for i, occupied := range cpus {
		if occupied {
			pe.ttr.serverlessManager.StopTask(i)
		} else if !occupied {
			if pe.ttr.serverlessManager.HasContinueTasks() {
				pe.ttr.serverlessManager.ContinueTask(i)
			} else if pe.ttr.serverlessManager.HasWaitTasks() {
				pe.ttr.serverlessManager.StartTask(i)
			}
		}
	}
}
