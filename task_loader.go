package totoro

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
	if task == L3 {
		tl.tasks = []Task {
			{
				Name: "l3-1",
				ImageName: "zehwei/ibench",
				Cmd: []string{"./l3", "200"},
				Status: WAITING,
			},
			{
				Name: "l3-2",
				ImageName: "zehwei/ibench",
				Cmd: []string{"./l3", "200"},
				Status: WAITING,
			},
			{
				Name: "l3-3",
				ImageName: "zehwei/ibench",
				Cmd: []string{"./l3", "200"},
				Status: WAITING,
			},
		}
	} else if task == CPU {
		tl.tasks = []Task {
			{
				Name: "sort1",
				ImageName: "zehwei/sort:1.0",
				Status: WAITING,
			},
			{
				Name: "cpu1",
				ImageName: "zehwei/ibench",
				Cmd: []string{"./cpu", "200"},
				Status: WAITING,
			},
			{
				Name: "cpu2",
				ImageName: "zehwei/ibench",
				Cmd: []string{"./cpu", "200"},
				Status: WAITING,
			},
		}
	} else if task == MEMbd {
		tl.tasks = []Task {
			{
				Name: "mem1",
				ImageName: "zehwei/ibench",
				Cmd: []string{"./memBw", "200"},
				Status: WAITING,
			},
			{
				Name: "mem2",
				ImageName: "zehwei/ibench",
				Cmd: []string{"./memBw", "200"},
				Status: WAITING,
			},
			{
				Name: "mem3",
				ImageName: "zehwei/ibench",
				Cmd: []string{"./memBw", "200"},
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
				Cmd: []string{"./l1i", "200"},
				Status: WAITING,
			},
			{
				Name: "l1i-2",
				ImageName: "zehwei/ibench",
				Cmd: []string{"./l1i", "200"},
				Status: WAITING,
			},
			{
				Name: "l1i-3",
				ImageName: "zehwei/ibench",
				Cmd: []string{"./l1i", "200"},
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