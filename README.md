## 这是什么

你有一个女秘书，她负责帮你从楼下传达室收包裹。当你需要取包裹的时候，只要写张小纸条放她桌上，她就会立刻去取，如果她还在去传达室的路上，你发现又有一个包裹需要取了，就再写一张小纸条放她桌子上。她回来之后发现桌子上一沓纸条，她会取出最下面那张（也就是最早那张）再去取，然后回来再拿一张纸条再取，直到纸条取完为止。她的力气比较小，而你的包裹都特别重，她只好一件一件去拿。这就是她一天的所有工作，你只需要给她纸条然后就什么都不需要管，下班的时候所有的包裹已经送到了你的办公室。她的名字叫 **任务队列** （task queue）。

这位秘书的勤勉让懒惰的你感到汗颜，于是你开除了她，然后雇了一个比较懒惰的。新的秘书稍微聪明些，经过多天对你的观察，发现你收到的包裹都是从淘宝买给家里用的东西，而且你每天准时下午三点带这些包裹回家，也就是说你三点之前不需要这些包裹。新秘书对此发现欣喜若狂，于是她每天下午2点半才到办公室上班，然后从两点半到三点清点她桌子上的纸条，一件一件地从传达室拿包裹（她的力气也比较小）。这位秘书叫 **延迟任务队列** （delayed task queue）。

时间就这样一天天过去了，直到有一天。。。你前几天从淘宝买的100件商品到了，这位秘书和她的小伙伴对你浪费公司人力资源的行为都惊呆了，而给她取包裹的时间只有半个小时！她以最快的速度把肝儿都累出来了还是没能在三点之前搬完，于是她两点半才上班的小秘密被你发现了。你无情地开除了她！这不公平！

于是，就有了你的现任，一位地道的女汉子。她也是每天两点半上班取包裹，但不同的是，她是位肌肉型秘书，一次可以搬十件包裹！这就极大缩短了搬运时间，因为她可以一次从桌子上抽十张纸条进行批量处理。从此以后，无论你从淘宝上买了多少东西，在三点之前他们都会一件不少地整齐地摆放在你的办公室里等着你带回家。你终于找到了心中梦寐以求的完美秘书，她的名字叫 **批处理延迟任务队列** （batch delayed task queue）。

## 安装/更新

    go get -u github.com/huichen/batchqueue

## 使用

定义你的任务，任务必须继承batchqueue.Task接口，也就是需要实现下面的BatchRun函数

```go
type MyTask struct {
}

// 批处理运行任务，tasks切片中的第一个任务就是此任务自己
func (t MyTask) BatchRun(queue *batchqueue.Queue, tasks []Task) {
	for _, task := range tasks {
		myTask := task.(MyTask)
		// 做点儿什么
    }
}
```

然后就可以用了
```go
var queue batchqueue.Queue
options := batchqueue.InitOptions{
	TimeUnit:         1000000, // 时间单位一毫秒，也就是最短的时间间隔
	NumWorkers:       10,      // 开辟多少个协程完成工作
	NumTasksPerBatch: 2}       // 一次最多能搬运多少个箱子
queue.Init(options)

// 加入任务到队列，
// 第一个参数是从现在开始的延迟时间，也就是说十个时间单位后开始做该任务
// 第二个参数是任务最大延迟，也就是说要在从现在开始的第十三个时间单位之前做掉
// 第三个参数是任务本身
queue.AddTask(10, 3, MyTask{})
// 可以反复添加任务
```

具体用法见[queue.go](/queue.go)中注释。

## 注意

* 当NumTasksPerBatch>1时，队列中的任务必须是同质的，也就是同样的类。否则BatchRun函数无法完成批量操作。
* 当NumTasksPerBathc=1时，队列中的任务可以不同质，这实际上退化成了无批处理的延迟任务队列
