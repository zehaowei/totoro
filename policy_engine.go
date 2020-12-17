package totoro

import (
	"fmt"
	"strconv"
	"sync"
	"time"
	"totoro/util"
)

const (
	UpBase	 float64 = 70
	OffSet	 float64 = 5
	Diff	 float64 = 20
)

type PolicyEngine struct {
	ttr 			*Totoro
	mu				sync.Mutex
	coreNums 		int
	upThreshold 	[]float64
	downThreshold 	[]float64

	timestamps		[]int64
	nums			int
}

func MakePolicyEngine(ttr *Totoro) *PolicyEngine {
	pe := new(PolicyEngine)
	pe.ttr = ttr
	pe.coreNums = CoreNums
	pe.upThreshold = make([]float64, pe.coreNums)
	pe.downThreshold = make([]float64, pe.coreNums)

	offset := 0.0
	pe.upThreshold[0] = 0
	for i := 1; i < pe.coreNums; i++ {
		pe.upThreshold[i] = UpBase * float64(i) + offset
		pe.downThreshold[i-1] = UpBase * float64(i) + offset - Diff
		offset += OffSet * float64(i+1)
	}
	pe.downThreshold[pe.coreNums-1] = 100 * float64(CoreNums)

	pe.timestamps = make([]int64, CoreNums-1)
	for i := 0; i <= CoreNums-2; i++ {
		pe.timestamps[i] = -1
	}
	pe.nums = 0
	return pe
}

func (pe *PolicyEngine) PolicyWithoutTask(cpuUsage float64) {
	pe.mu.Lock()
	defer pe.mu.Unlock()
	cpuNums := pe.ttr.mainAppManager.cpuNums
	for i := 1; i <= pe.coreNums; i++ {
		if cpuUsage >= pe.upThreshold[i-1] && cpuUsage <= pe.downThreshold[i-1] && cpuNums != i {
			util.PrintInfo("[time] --- trigger policy for main app (cpu: %f) --- %v", cpuUsage, time.Now().Unix())

			if i == 1 {
				pe.ttr.mainAppManager.UpdateCpuSet("0")
			} else {
				cpuSet := "0-" + strconv.Itoa(i-1)
				pe.ttr.mainAppManager.UpdateCpuSet(cpuSet)
			}
			pe.ttr.mainAppManager.cpuNums = i

			// for information collecting
			if i > 1 && pe.timestamps[i-2] == -1 {
				pe.timestamps[i-2] = time.Now().Unix()
				pe.nums++
				for _,v := range pe.timestamps {
					fmt.Print(v)
					fmt.Print(",")
				}
				fmt.Println()
			}
			//util.PrintInfo("[time] --- complete policy for main app --- %v", time.Now().Unix())
		}
	}
}

func (pe *PolicyEngine) SimplePolicy(cpuUsage float64) {
	pe.mu.Lock()
	defer pe.mu.Unlock()
	cpuNums := pe.ttr.mainAppManager.cpuNums
	for i := 1; i <= pe.coreNums; i++ {
		if cpuUsage >= pe.upThreshold[i-1] && cpuUsage <= pe.downThreshold[i-1] {
			cpus := make([]bool, pe.coreNums)
			for j := 0; j <= i-1; j++ {
				cpus[j] = true
			}
			if cpuNums != i {
				util.PrintInfo("[time] --- trigger policy for main app (cpu: %f) --- %v", cpuUsage, time.Now().Unix())

				pe.killServerless(cpus)
				if i == 1 {
					pe.ttr.mainAppManager.UpdateCpuSet("0")
				} else {
					cpuSet := "0-" + strconv.Itoa(i-1)
					pe.ttr.mainAppManager.UpdateCpuSet(cpuSet)
				}
				pe.ttr.mainAppManager.cpuNums = i

				if i > 1 && pe.timestamps[i-2] == -1 {
					pe.timestamps[i-2] = time.Now().Unix()
					pe.nums++
					for _,v := range pe.timestamps {
						fmt.Print(v)
						fmt.Print(",")
					}
					fmt.Println()
				}
				//util.PrintInfo("[time] --- complete policy for main app --- %v", time.Now().Unix())
			} else {
				pe.startServerless(cpus)
			}
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
	for i := len(cpus)-1; i >= 0; i-- {
		if !cpus[i] {
			if pe.ttr.serverlessManager.HasContinueTasks() {
				pe.ttr.serverlessManager.ContinueTask(i)
			} else if pe.ttr.serverlessManager.HasWaitTasks() {
				pe.ttr.serverlessManager.StartTask(i)
			}
		}
	}
}
