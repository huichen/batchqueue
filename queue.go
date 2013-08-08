package batchqueue

import (
	"log"
	"runtime"
	"sync"
	"time"
)

// 批处理延迟任务队列
type Queue struct {
	taskList         TaskList
	timeUnit         uint64
	startTime        time.Time
	numTasksPerBatch int
	runnerChannel    chan []Task
	sync.RWMutex
}

type InitOptions struct {
	// 时间单位的纳秒数，任务执行的最小时间间隔。
	TimeUnit uint64

	// 执行批处理操作的最大协程数目，至少为1。
	// 请根据负载合理设置此值，否则AddTask可能会产生阻塞。
	NumWorkers int

	// 批处理最大的任务数目。
	NumTasksPerBatch int
}

// 初始化任务队列，并开始计时。
func (q *Queue) Init(options InitOptions) {
	q.timeUnit = options.TimeUnit
	q.startTime = time.Now()
	if options.NumTasksPerBatch <= 0 {
		q.numTasksPerBatch = 1
	} else {
		q.numTasksPerBatch = options.NumTasksPerBatch
	}

	runtime.GOMAXPROCS(runtime.NumCPU())
	if options.NumWorkers <= 0 {
		log.Fatal("InitOptions.NumWorkers必须大于零")
	}
	q.runnerChannel = make(chan []Task, options.NumWorkers)
	for i := 0; i < options.NumWorkers; i++ {
		go q.worker()
	}
	go q.start()
}

// 在当前时间加delay个时间单位后执行任务，任务的过期时间（从delay之后开始算）
// 为timeout个时间单位。也就是说，任务的执行时间窗口为
// [now+delay, now+delay+timeout]
func (q *Queue) AddTask(delay uint64, timeout uint64, task Task) {
	runner := new(Runner)
	runner.task = task
	runner.time = q.Now() + delay
	runner.timeout = timeout
	q.Lock()
	q.insert(&(q.taskList), runner)
	q.Unlock()
}

// 当前时间，以Init调用开始为零点。单位为初始化时定义的时间单位。
func (q *Queue) Now() uint64 {
	return uint64(time.Now().Sub(q.startTime).Nanoseconds()) / q.timeUnit
}

func (q *Queue) insert(list *TaskList, runner *Runner) {
	if list.head == nil {
		list.head = runner
		return
	}

	if list.head.time > runner.time {
		runner.next = list.head
		list.head = runner
		return
	}

	current := list.head
	for ; current.next != nil && current.next.time <= runner.time; current = current.next {
	}

	if current.next == nil {
		current.next = runner
	} else {
		runner.next = current.next
		current.next = runner
	}
}

func (q *Queue) start() {
	oldTick := q.Now()
	for {
		q.Lock()
		if q.taskList.head == nil {
			continue
		}
		tasks := make([]Task, q.numTasksPerBatch)
		expiredTasks := make([]Task, q.numTasksPerBatch)
		aliveRunners := make([]*Runner, q.numTasksPerBatch)
		now := q.Now()
		taskCount := 0
		current := q.taskList.head
		numExpiredTasks := 0
		numAliveRunners := 0
		for ; current != nil && current.time <= now; current = current.next {
			if current.timeout+current.time <= now {
				expiredTasks[numExpiredTasks] = current.task
				numExpiredTasks++
			} else {
				aliveRunners[numAliveRunners] = current
				numAliveRunners++
			}
			tasks[taskCount] = current.task
			taskCount++
			if taskCount >= q.numTasksPerBatch {
				q.runnerChannel <- tasks
				tasks = make([]Task, q.numTasksPerBatch)
				taskCount = 0
				numExpiredTasks = 0
				numAliveRunners = 0
			}
		}

		if numExpiredTasks != 0 {
			q.runnerChannel <- expiredTasks[0:numExpiredTasks]
		}

		if numAliveRunners != 0 {
			q.taskList.head = aliveRunners[0]
			for iRunner := 0; iRunner < numAliveRunners-1; iRunner++ {
				aliveRunners[iRunner].next = aliveRunners[iRunner+1]
			}
			aliveRunners[numAliveRunners-1].next = current
		} else {
			q.taskList.head = current
		}
		q.Unlock()

		// 等待下一时间
		newTick := q.Now() + 1
		if newTick == oldTick {
			newTick++
		}
		delay := q.startTime.Add(
			time.Nanosecond * time.Duration(q.timeUnit*newTick)).Sub(time.Now())
		select {
		case <-time.After(delay):
		}
		oldTick = newTick
	}
}

func (q *Queue) worker() {
	for {
		tasks := <-q.runnerChannel
		if len(tasks) > 0 {
			tasks[0].BatchRun(tasks)
		}
	}
}
