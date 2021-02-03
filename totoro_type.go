package totoro

import (
	"github.com/docker/docker/api/types/strslice"
	"time"
)

// task status
const (
	RUNNING 	int = 0
	STOPPED 	int = 1
	COMPLETE 	int = 2
	WAITING		int = 3
)

// cpu status
const (
	OCCUPIED	int = 0
	AVAILABLE	int = 1
)

const CoreNums int = 8									// number of physical cores of the server
const MonitorInterval = 2500 * time.Millisecond  		// interval for triggering policy
const HeartbeatInterval = 1000 * time.Millisecond		// interval for collecting information of cpu usage
const TaskSchedulingInterval = 100 * time.Millisecond   // interval for task scheduling policy
const TaskStatusCheckInterval = 100 * time.Millisecond  // interval for task status checking
const ShortToLongThreshold = 2							// threshold to determine whether change a short term task into long term
const LowHighLoadThreshold = CoreNums/2+1				// threshold to determine whether main app is under high pressure

const CpuIndexGap = 16
const CpuSetsNum = 2
const TaskLimitPerCore = 2
const JobIdPrefix = "job"
const TaskIdPrefix = "task"
const ShortTask = "shortTask"
const LongTask = "longTask"
const JobInfoFile = "../output/jobs.out"
const CpuInfoFile = "../output/cpuUsage.out"
const TimeLayoutStr = "2006-01-02 15:04:05"
const JobTracePathLowLoad = "../resources/task_trace/trace_ibench_low.json"
const JobTracePathNormalLoad = "../resources/task_trace/trace_ibench_normal.json"
const JobTracePathHighLoad = "../resources/task_trace/trace_ibench_high.json"
const JobTracePathStepLoad = "../resources/task_trace/trace_ibench_step.json"

// a request from a client, containing time
type Request struct {
	job         	*Job
	requestTime 	float64
}

// a job contains one or several tasks
type Job struct {
	JobId			string
	JobName 		string
	Tasks 			[]*Task
	Start 			time.Time
	End   			time.Time
}

// an executing unit, running in a separate container
type Task struct {
	TaskId			string
	Name 			string
	ImageName		string
	Cmd  			strslice.StrSlice

	ContainerId 	string
	Status 			int
	CpuSet			int
	ExecuteTimes	int
	Type			string
	NotifyFinish 	chan struct{}
}

type RequestJson struct {
	JobName 		string
	StartTime 		float64
	Tasks 			[]TaskJson
}

type TaskJson struct {
	Name 			string
	ImageName		string
	Cmd  			strslice.StrSlice
}
