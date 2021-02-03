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
	Proportion	float64 = 0.625  // normal situation
	Proportion2 float64 = 0.585	 // job: high load
)

/*
 * determine the cpu resource allocation of main app and serverless tasks
 */
type PolicyEngine struct {
	ttr 			*Totoro
	mu				sync.Mutex
	procNums 		int				// number of hyperthread, one core has two hyperthread
	coreNums		int				// number of physical cores

	timestamps		[]int64			// just for experiments to collect statistics
	nums			int				// just for experiments to collect statistics

	thresholdBase	float64			// base threshold to generate thresholds of inflation and deflation
	inflateThresh  	[]float64		// thresholds to inflate memcached
	inflateThresh2	[]float64		// thresholds to apply aggressive policy
	deflateThresh	[]float64		// thresholds to deflate memcached
	deflateThresh2	[]float64		// thresholds to apply aggressive policy
	timestamps2		[]int64			// just for experiments to collect statistics
}

/*
 * constructor
 */
func MakePolicyEngine(ttr *Totoro) *PolicyEngine {
	pe := new(PolicyEngine)
	pe.ttr = ttr
	pe.coreNums = CoreNums
	pe.procNums = CoreNums*2

	if ttr.jobTraceType == JobTracePathHighLoad {
		pe.thresholdBase = Proportion2
	} else {
		pe.thresholdBase = Proportion
	}
	pe.inflateThresh = make([]float64, pe.coreNums+1)
	pe.deflateThresh = make([]float64, pe.coreNums+1)
	pe.inflateThresh2 = make([]float64, pe.coreNums+1)
	pe.deflateThresh2 = make([]float64, pe.coreNums+1)

	for i := 1; i <= pe.coreNums; i++ {
		if i == 1 {
			pe.inflateThresh[i] = 100
			pe.inflateThresh2[i] = 180

			pe.deflateThresh[i] = 0
			pe.deflateThresh2[i] = 0
		} else if i == pe.coreNums {
			pe.inflateThresh[i] = 12000
			pe.inflateThresh2[i] = 14000

			pe.deflateThresh[i] = pe.inflateThresh[i-1] - pe.thresholdBase * 100
			pe.deflateThresh2[i] = pe.inflateThresh[i-2] - pe.thresholdBase * 100
		} else {
			pe.inflateThresh[i] = pe.thresholdBase * 200 * float64(i)
			pe.inflateThresh2[i] = pe.thresholdBase * 200 * float64(i+1)

			pe.deflateThresh[i] = pe.inflateThresh[i-1] - pe.thresholdBase * 100
			pe.deflateThresh2[i] = pe.inflateThresh[i-2] - pe.thresholdBase * 100
		}
	}
	pe.timestamps2 = make([]int64, 0)
	return pe
}

/*
 * For experiment
 * policy without running any task
 */
func (pe *PolicyEngine) SecondPolicyWithoutTask(cpuUsage float64) {
	pe.mu.Lock()
	defer pe.mu.Unlock()
	curNums := pe.ttr.mainAppManager.cpuNums / 2

	if cpuUsage > pe.inflateThresh[curNums] { // current cpuUsage exceeds threshold, add one more physical core
		util.PrintInfo("[time] --- trigger policy for main app (cpu: %f) --- %v", cpuUsage, time.Now().Unix())
		//pe.timestamps2 = append(pe.timestamps2, time.Now().Unix())
		newNums := curNums+1
		if cpuUsage > pe.inflateThresh2[curNums] {
			newNums += 1
		}
		newNums = int(math.Min(float64(newNums), float64(pe.coreNums)))

		cpuSets := "0,16"
		for k := 2; k <= newNums; k++ {
			cpuSets += "," + strconv.Itoa((k-1)*2) + "," + strconv.Itoa((k-1)*2+CpuIndexGap)
		}
		pe.ttr.mainAppManager.UpdateCpuSet(cpuSets)
		pe.ttr.mainAppManager.cpuNums = newNums*2

	} else if cpuUsage < pe.deflateThresh[curNums] { // current cpuUsage is below threshold, remove one physical core
		util.PrintInfo("[time] --- trigger policy for main app (cpu: %f) --- %v", cpuUsage, time.Now().Unix())
		//pe.timestamps2 = append(pe.timestamps2, time.Now().Unix())
		newNums := curNums-1
		if cpuUsage < pe.deflateThresh2[curNums] {
			newNums -= 1
		}
		newNums = int(math.Max(float64(newNums), float64(1)))

		cpuSets := "0,16"
		for k := 2; k <= newNums; k++ {
			cpuSets += "," + strconv.Itoa((k-1)*2) + "," + strconv.Itoa((k-1)*2+CpuIndexGap)
		}
		pe.ttr.mainAppManager.UpdateCpuSet(cpuSets)
		pe.ttr.mainAppManager.cpuNums = newNums*2
	}

	// just for experiments to collect statistics
	if cpuUsage < 0 {
		for _,v := range pe.timestamps2 {
			fmt.Print(v)
			fmt.Print(",")
		}
		fmt.Println()
	}
}

func (pe *PolicyEngine) SecondPolicy(cpuUsage float64) {
	pe.mu.Lock()
	defer pe.mu.Unlock()
	curNums := pe.ttr.mainAppManager.cpuNums / 2

	if cpuUsage > pe.inflateThresh[curNums] { // current cpuUsage exceeds threshold, add one more physical core
		util.PrintInfo("[time] --- trigger policy for main app (cpu: %f) --- %v", cpuUsage, time.Now().Unix())
		newNums := curNums+1
		if cpuUsage > pe.inflateThresh2[curNums] {
			newNums += 1
		}
		newNums = int(math.Min(float64(newNums), float64(pe.coreNums)))

		cpuSets := "0,16"
		for k := 2; k <= newNums; k++ {
			cpuSets += "," + strconv.Itoa((k-1)*2) + "," + strconv.Itoa((k-1)*2+CpuIndexGap)
		}
		// notify taskScheduler the change of cpu allocation
		pe.ttr.taskScheduler.Deflate(GetCpuState(newNums))
		pe.ttr.mainAppManager.UpdateCpuSet(cpuSets)
		pe.ttr.mainAppManager.cpuNums = newNums*2

	} else if cpuUsage < pe.deflateThresh[curNums] { // current cpuUsage is below threshold, remove one physical core
		util.PrintInfo("[time] --- trigger policy for main app (cpu: %f) --- %v", cpuUsage, time.Now().Unix())
		newNums := curNums-1
		if cpuUsage < pe.deflateThresh2[curNums] {
			newNums -= 1
		}
		newNums = int(math.Max(float64(newNums), float64(1)))

		cpuSets := "0,16"
		for k := 2; k <= newNums; k++ {
			cpuSets += "," + strconv.Itoa((k-1)*2) + "," + strconv.Itoa((k-1)*2+CpuIndexGap)
		}
		pe.ttr.taskScheduler.Inflate(GetCpuState(newNums))
		pe.ttr.mainAppManager.UpdateCpuSet(cpuSets)
		pe.ttr.mainAppManager.cpuNums = newNums*2
	}
}

func GetCpuState(newNums int) []int {
	cpuState := make([]int, CoreNums)
	for i := 0; i < newNums; i++ {
		cpuState[i] = OCCUPIED
	}
	for i := newNums; i < CoreNums; i++ {
		cpuState[i] = AVAILABLE
	}
	return cpuState
}
