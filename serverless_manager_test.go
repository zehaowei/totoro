package totoro

import "testing"

func TestServerlessManager_StopTask(t *testing.T) {
	ttr := MakeTotoro()
	runningTasks := ttr.serverlessManager.GetRunningTasks()
	i := 1
	if ttr.serverlessManager.HasContinueTasks() {
		ttr.serverlessManager.ContinueTask(i)
	} else if ttr.serverlessManager.HasWaitTasks() {
		ttr.serverlessManager.StartTask(i)
	}
	cpus := []bool{false, true, false, false}
	for i, occupied := range cpus {
		if occupied && len(runningTasks[i]) != 0 {
			println("len: %d", len(runningTasks[i]))
			ttr.serverlessManager.StopTask(i)
		}
	}
}
