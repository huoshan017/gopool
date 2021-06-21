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

// 协程池
type Pool struct {
	taskCh    chan *task    // 传递任务
	countCh   chan struct{} // 用于计数
	goTimeout int           // 超时：秒
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
	case p.countCh <- struct{}{}:
		go p.addCandidate(ctx, &task{fun: f, param: param})
		// 只有在p.semCh满的情况下才会跑到下面来
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
