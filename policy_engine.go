package totoro

type PolicyEngine struct {

}

func MakePolicyEngine() *PolicyEngine {

}

func (pe *PolicyEngine) simplePolicy(cpuUsage float64) {
	if cpuUsage <= 10.0 {
		ttr.serverlessManager.RunningTask()
	} else if cpuUsage > 40.0 && cpuUsage < 60.0{
		ttr.serverlessManager.UpdateTask()
	} else if cpuUsage >= 60.0 {
		ttr.serverlessManager.StopTask()
	}
}
