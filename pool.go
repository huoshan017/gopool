package gopool

import (
	"context"
	"time"
)

const (
	GOROUTINE_RUNNING_SECONDS = 10
)

// 任务
type task struct {
	param interface{}
	fun   func(interface{}) interface{}
}

// 线程安全的协程池
type Pool struct {
	taskCh    chan *task    // 传递任务，这个通道是多写多读
	countCh   chan struct{} // 用于计数
	goTimeout int           // 协程执行完退出超时：秒
}

// 新建协程池
func NewPool(s int) *Pool {
	p := &Pool{
		taskCh:    make(chan *task),
		countCh:   make(chan struct{}, s),
		goTimeout: GOROUTINE_RUNNING_SECONDS,
	}
	return p
}

// 提交任务
func (p *Pool) CommitTask(ctx context.Context, f func(interface{}) interface{}, param interface{}) {
	select {
		// 进行计数，如果阻塞则说明协程数已达最大，等待有协程执行完释放
	case p.countCh <- struct{}{}:
		go p.addCandidate(ctx, &task{fun: f, param: param})
		// 只有上面阻塞时才会跑到下面来，如果已经有一个协程在运行了，会通过taskCh通道获取到该任务，协程会继续运行
	case p.taskCh <- &task{fun: f, param: param}:
	}
}

// 设置goroutine超时
func (p *Pool) SetGoTimeout(timeout int) {
	p.goTimeout = timeout
}

// 添加任务到一个协程的候选中
func (p *Pool) addCandidate(ctx context.Context, t *task) {
	defer func() {
		<-p.countCh
	}()

	for {
		t.fun(t.param)
		select {
		case <-ctx.Done():
			return
		case t = <-p.taskCh:
		case <-time.After(time.Duration(p.goTimeout) * time.Second):
			return
		}
	}
}
