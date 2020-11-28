package totoro

import "strconv"

const (
	L3 			int = 0
	CPU 		int = 1
	MEMbd 		int = 2
	L1d			int = 3
	L1i			int = 4
	L2			int = 5
)

type TaskLoader struct {
	tasks []Task
}

func MakeTaskLoader() *TaskLoader {
	tl := new(TaskLoader)
	tl.tasks = nil
	return tl
}

func (tl *TaskLoader) loadTasks(task int) []Task {
	duration := "300"
	taskNum := 5
	if task == L3 {
		tl.tasks = make([]Task, 0)
		for i := 0; i < taskNum; i++ {
			tl.tasks = append(tl.tasks, Task{
				Name: "l3-" + strconv.Itoa(i),
				ImageName: "zehwei/ibench",
				Cmd: []string{"./l3", duration},
				Status: WAITING,
			})
		}
	} else if task == CPU {
		tl.tasks = make([]Task, 0)
		for i := 0; i < taskNum; i++ {
			tl.tasks = append(tl.tasks, Task{
				Name: "cpu-" + strconv.Itoa(i),
				ImageName: "zehwei/ibench",
				Cmd: []string{"./cpu", duration},
				Status: WAITING,
			})
		}
	} else if task == MEMbd {
		tl.tasks = make([]Task, 0)
		for i := 0; i < taskNum; i++ {
			tl.tasks = append(tl.tasks, Task{
				Name: "mem-" + strconv.Itoa(i),
				ImageName: "zehwei/ibench",
				Cmd: []string{"./memBw", duration},
				Status: WAITING,
			})
		}
	} else if task == L1d {
		tl.tasks = make([]Task, 0)
		for i := 0; i < taskNum; i++ {
			tl.tasks = append(tl.tasks, Task{
				Name: "l1d-" + strconv.Itoa(i),
				ImageName: "zehwei/ibench",
				Cmd: []string{"./l1d", duration},
				Status: WAITING,
			})
		}
	} else if task == L1i {
		tl.tasks = make([]Task, 0)
		for i := 0; i < taskNum; i++ {
			tl.tasks = append(tl.tasks, Task{
				Name: "l1i-" + strconv.Itoa(i),
				ImageName: "zehwei/ibench",
				Cmd: []string{"./l1i", duration, "18"},
				Status: WAITING,
			})
		}
	} else if task == L2 {
		tl.tasks = make([]Task, 0)
		for i := 0; i < taskNum; i++ {
			tl.tasks = append(tl.tasks, Task{
				Name: "l2-" + strconv.Itoa(i),
				ImageName: "zehwei/ibench",
				Cmd: []string{"./l2", duration},
				Status: WAITING,
			})
		}
	}
	return tl.tasks
}

func (tl *TaskLoader) loadTasks2(task int) []Task {
	duration := "300"
	if task == L3 {
		tl.tasks = []Task {
			{
				Name: "l3-1",
				ImageName: "zehwei/ibench",
				Cmd: []string{"./l3", duration},
				Status: WAITING,
			},
			{
				Name: "l3-2",
				ImageName: "zehwei/ibench",
				Cmd: []string{"./l3", duration},
				Status: WAITING,
			},
			{
				Name: "l3-3",
				ImageName: "zehwei/ibench",
				Cmd: []string{"./l3", duration},
				Status: WAITING,
			},
		}
	} else if task == CPU {
		tl.tasks = []Task {
			{
				Name: "cpu1",
				ImageName: "zehwei/ibench",
				Cmd: []string{"./cpu", duration},
				Status: WAITING,
			},
			{
				Name: "cpu2",
				ImageName: "zehwei/ibench",
				Cmd: []string{"./cpu", duration},
				Status: WAITING,
			},
		}
	} else if task == MEMbd {
		tl.tasks = []Task {
			{
				Name: "mem1",
				ImageName: "zehwei/ibench",
				Cmd: []string{"./memBw", duration},
				Status: WAITING,
			},
			{
				Name: "mem2",
				ImageName: "zehwei/ibench",
				Cmd: []string{"./memBw", duration},
				Status: WAITING,
			},
			{
				Name: "mem3",
				ImageName: "zehwei/ibench",
				Cmd: []string{"./memBw", duration},
				Status: WAITING,
			},
		}
	} else if task == L1d {
		tl.tasks = []Task {
			{
				Name: "l1d-1",
				ImageName: "zehwei/ibench",
				Cmd: []string{"./l1d", "200"},
				Status: WAITING,
			},
			{
				Name: "l1d-2",
				ImageName: "zehwei/ibench",
				Cmd: []string{"./l1d", "200"},
				Status: WAITING,
			},
			{
				Name: "l1d-3",
				ImageName: "zehwei/ibench",
				Cmd: []string{"./l1d", "200"},
				Status: WAITING,
			},
		}
	} else if task == L1i {
		tl.tasks = []Task {
			{
				Name: "l1i-1",
				ImageName: "zehwei/ibench",
				Cmd: []string{"./l1i", "200", "18"},
				Status: WAITING,
			},
			{
				Name: "l1i-2",
				ImageName: "zehwei/ibench",
				Cmd: []string{"./l1i", "200", "18"},
				Status: WAITING,
			},
			{
				Name: "l1i-3",
				ImageName: "zehwei/ibench",
				Cmd: []string{"./l1i", "200", "18"},
				Status: WAITING,
			},
		}
	} else if task == L2 {
		tl.tasks = []Task {
			{
				Name: "l2-1",
				ImageName: "zehwei/ibench",
				Cmd: []string{"./l2", "200"},
				Status: WAITING,
			},
			{
				Name: "l2-2",
				ImageName: "zehwei/ibench",
				Cmd: []string{"./l2", "200"},
				Status: WAITING,
			},
			{
				Name: "l2-3",
				ImageName: "zehwei/ibench",
				Cmd: []string{"./l2", "200"},
				Status: WAITING,
			},
		}
	}
	return tl.tasks
}
