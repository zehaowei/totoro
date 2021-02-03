package totoro

import (
	"encoding/json"
	"os"
	"sort"
	"time"
	"totoro/util"
)

/*
 * simulate the client to send serverless tasks to controller
 */
type RequestSimulator struct {
	requests 		[]*Request		// all the requests to be sent
	sendIndex	 	int				// the next request to be sent
	server 			*JobScheduler	// server that handle requests
}

/*
 * constructor
 */
func MakeRequestSimulator(js *JobScheduler) *RequestSimulator {
	rs := new(RequestSimulator)
	rs.requests = make([]*Request, 0)
	rs.sendIndex = 0
	rs.server = js
	return rs
}

/*
 * read json file to load jobs's information
 */
func (rs *RequestSimulator) ReadJobs(filePath string) {
	f, err := os.Open(filePath)
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
			task.NotifyFinish = make(chan struct{})
			tasks = append(tasks, &task)
		}
		j.Tasks = tasks

		r.job = &j
		rs.requests = append(rs.requests, &r)
	}

	go rs.monitorRequests()
}

/*
 * send requests to the JobLoader according to the StartTime in job description
 */
func (rs *RequestSimulator) monitorRequests() {
	startTime := time.Now()
	for rs.sendIndex < len(rs.requests)  {
		r := rs.requests[rs.sendIndex]
		t := time.Duration(r.requestTime*1000) * time.Millisecond - time.Now().Sub(startTime)
		if t > 0 {
			time.Sleep(t)
		}
		if time.Now().Sub(startTime) >= time.Duration(r.requestTime*1000) * time.Millisecond {
			// use RPC to simulate sending requests to controller
			r.job.Start = time.Now()
			go rs.server.ReceiveRPC(r.job)
			rs.sendIndex++
		}
	}
}
