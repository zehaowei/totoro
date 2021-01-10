package totoro

import (
	"encoding/json"
	"os"
	"sort"
	"testing"
	"totoro/util"
)

func TestAll(t *testing.T) {
	f, err := os.Open(JobTracePath)
	if err != nil {
		util.PrintErr("[error] file read err")
		return
	}
	var requests []RequestJson
	decoder := json.NewDecoder(f)
	err = decoder.Decode(&requests)
	if err != nil {
		util.PrintErr("[error] decode json err")
		return
	}

	// sort requests by sending time
	sort.Slice(requests, func(i, j int) bool {
		return requests[i].StartTime < requests[j].StartTime
	})

	for _, request := range requests {
		r := Request{}
		r.requestTime = request.StartTime

		j := Job{}
		j.JobName = request.JobName
		tasks := make([]*Task, 0)
		for _, t := range request.Tasks {
			task := Task{}
			task.Name = t.Name
			task.ImageName = t.ImageName
			task.Cmd = t.Cmd
			tasks = append(tasks, &task)
		}
		j.Tasks = tasks

		r.job = &j
	}
}
