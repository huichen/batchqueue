package batchqueue

import (
	"fmt"
	"testing"
	"time"
)

type MyTask struct {
	id string
}

func (t MyTask) BatchRun(followingTasks []Task) {
	output := ""
	for _, t := range followingTasks {
		output += fmt.Sprintf("%s ", t.(MyTask).id)
	}
	fmt.Println(output)
}

func TestAddTask(t *testing.T) {
	var queue Queue
	options := InitOptions{
		TimeUnit:         1000000, // 时间单位一毫秒
		NumWorkers:       1,
		NumTasksPerBatch: 2}
	queue.Init(options)
	queue.AddTask(10, 0, MyTask{"10"})
	queue.AddTask(0, 5, MyTask{"0"})
	queue.AddTask(7, 2, MyTask{"7"})
	queue.AddTask(1, 5, MyTask{"1"})
	queue.AddTask(3, 5, MyTask{"3"})
	queue.AddTask(2, 5, MyTask{"2"})
	queue.AddTask(17, 5, MyTask{"17"})
	fmt.Println("任务总数", queue.NumTasks())
	time.Sleep(time.Nanosecond * time.Duration(options.TimeUnit*20))
}
