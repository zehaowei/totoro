package totoro

import (
	"fmt"
	"math"
	"strconv"
	"sync"
	"time"
	"totoro/util"
)

const (
	UpBase	 	float64 = 70
	OffSet	 	float64 = 5
	Diff	 	float64 = 20

	Proportion	float64 = 0.625
)

type PolicyEngine struct {
	ttr 			*Totoro
	mu				sync.Mutex
	procNums 		int				// number of hyperthread, one core has two hyperthread
	coreNums		int				// number of physical cores

	upThreshold 	[]float64
	downThreshold 	[]float64
	timestamps		[]int64			// just for experiments to collect statistics
	nums			int				// just for experiments to collect statistics
	
	inflateThresh  	[]float64		// thresholds to inflate memcached
	inflateThresh2	[]float64		// thresholds to apply aggressive policy
	deflateThresh	[]float64		// thresholds to deflate memcached
	deflateThresh2	[]float64		// thresholds to apply aggressive policy
	timestamps2		[]int64			// just for experiments to collect statistics
}

func MakePolicyEngine(ttr *Totoro) *PolicyEngine {
	pe := new(PolicyEngine)
	pe.ttr = ttr
	pe.coreNums = CoreNums
	pe.procNums = CoreNums*2
	//pe.upThreshold = make([]float64, pe.procNums)
	//pe.downThreshold = make([]float64, pe.procNums)

	pe.inflateThresh = make([]float64, pe.coreNums+1)
	pe.deflateThresh = make([]float64, pe.coreNums+1)
	pe.inflateThresh2 = make([]float64, pe.coreNums+1)
	pe.deflateThresh2 = make([]float64, pe.coreNums+1)

	//offset := 0.0
	//pe.upThreshold[0] = 0
	//for i := 1; i < pe.procNums; i++ {
	//	pe.upThreshold[i] = UpBase * float64(i) + offset
	//	pe.downThreshold[i-1] = UpBase * float64(i) + offset - Diff
	//	offset += OffSet * float64(i+1)
	//}
	//pe.downThreshold[pe.procNums-1] = 100 * float64(CoreNums)
	//
	//pe.timestamps = make([]int64, CoreNums-1)
	//for i := 0; i <= CoreNums-2; i++ {
	//	pe.timestamps[i] = -1
	//}
	//pe.nums = 0

	for i := 1; i <= pe.coreNums; i++ {
		if i == 1 {
			pe.inflateThresh[i] = 100
			pe.inflateThresh2[i] = 180

			pe.deflateThresh[i] = 0
			pe.deflateThresh2[i] = 0
		} else if i == pe.coreNums {
			pe.inflateThresh[i] = 12000
			pe.inflateThresh2[i] = 14000

			pe.deflateThresh[i] = pe.inflateThresh[i-1] - Proportion * 100
			pe.deflateThresh2[i] = pe.inflateThresh[i-2] - Proportion * 100
		} else {
			pe.inflateThresh[i] = Proportion * 200 * float64(i)
			pe.inflateThresh2[i] = Proportion * 200 * float64(i+1)

			pe.deflateThresh[i] = pe.inflateThresh[i-1] - Proportion * 100
			pe.deflateThresh2[i] = pe.inflateThresh[i-2] - Proportion * 100
		}
	}
	pe.timestamps2 = make([]int64, 0)
	return pe
}

func (pe *PolicyEngine) PolicyWithoutTask(cpuUsage float64) {
	pe.mu.Lock()
	defer pe.mu.Unlock()
	cpuNums := pe.ttr.mainAppManager.cpuNums
	for i := 1; i <= pe.procNums; i++ {
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
	for i := 1; i <= pe.procNums; i++ {
		if cpuUsage >= pe.upThreshold[i-1] && cpuUsage <= pe.downThreshold[i-1] {
			cpus := make([]bool, pe.procNums)
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

func (pe *PolicyEngine) SecondPolicyWithoutTask(cpuUsage float64) {
	pe.mu.Lock()
	defer pe.mu.Unlock()
	curNums := pe.ttr.mainAppManager.cpuNums / 2

	if cpuUsage > pe.inflateThresh[curNums] { // current cpuUsage exceeds threshold, add one more physical core
		util.PrintInfo("[time] --- trigger policy for main app (cpu: %f) --- %v", cpuUsage, time.Now().Unix())
		pe.timestamps2 = append(pe.timestamps2, time.Now().Unix())
		newNums := curNums+1
		if cpuUsage > pe.inflateThresh2[curNums] {
			newNums += 1
		}
		newNums = int(math.Min(float64(newNums), float64(pe.coreNums)))

		cpuSets := "0,16"
		for k := 2; k <= newNums; k++ {
			cpuSets += "," + strconv.Itoa((k-1)*2) + "," + strconv.Itoa((k-1)*2+16)
		}
		pe.ttr.mainAppManager.UpdateCpuSet(cpuSets)
		pe.ttr.mainAppManager.cpuNums = newNums*2

	} else if cpuUsage < pe.deflateThresh[curNums] { // current cpuUsage is below threshold, remove one physical core
		util.PrintInfo("[time] --- trigger policy for main app (cpu: %f) --- %v", cpuUsage, time.Now().Unix())
		pe.timestamps2 = append(pe.timestamps2, time.Now().Unix())
		newNums := curNums-1
		if cpuUsage < pe.deflateThresh2[curNums] {
			newNums -= 1
		}
		newNums = int(math.Max(float64(newNums), float64(1)))

		cpuSets := "0,16"
		for k := 2; k <= newNums; k++ {
			cpuSets += "," + strconv.Itoa((k-1)*2) + "," + strconv.Itoa((k-1)*2+16)
		}
		pe.ttr.mainAppManager.UpdateCpuSet(cpuSets)
		pe.ttr.mainAppManager.cpuNums = newNums*2
	}

	// just for experiments to collect statistics
	if cpuUsage < 10 {
		for _,v := range pe.timestamps2 {
			fmt.Print(v)
			fmt.Print(",")
		}
		fmt.Println()
	}
}
