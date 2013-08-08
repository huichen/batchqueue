package batchqueue

type Runner struct {
	time    uint64
	timeout uint64
	task    Task
	next    *Runner
}

type Task interface {
	BatchRun(tasks []Task)
}

type TaskNode struct {
	task *Runner
	next *TaskNode
}

type TaskList struct {
	head *Runner
}
