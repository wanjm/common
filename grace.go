package common

import (
	"context"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
)

// ==========================================
// 1. 封装优雅退出管理器
// ==========================================

type gracefulManager struct {
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
	activeJobs sync.Map // map[string]bool - track active job names
}

// NewGracefulManager 创建管理器。信号在 Wait 时才注册，避免未调用 Wait 时吞掉 Ctrl+C。
func NewGracefulManager() *gracefulManager {
	ctx, cancel := context.WithCancel(context.Background())
	return &gracefulManager{
		ctx:    ctx,
		cancel: cancel,
	}
}

// Go 核心点 3: 启动协程帮助函数
// 自动注册 (Add) 和 解注册 (Done)
func (m *gracefulManager) Go(jobName string, fn func(ctx context.Context)) {
	m.wg.Add(1)
	m.activeJobs.Store(jobName, true) // Track job as active
	go func() {
		defer func() {
			m.activeJobs.Delete(jobName) // Remove from tracking when job finishes
			m.wg.Done()                  // 协程退出时取消注册
		}()
		ctx := context.WithValue(m.ctx, TraceIdNameInContext, jobName)
		defer Recover(ctx, jobName)
		Info(ctx, "Job started", String("jobName", jobName))
		fn(ctx) // 将可被 Wait 取消的 ctx 传递给业务逻辑
		Info(ctx, "Job finished", String("jobName", jobName))
	}()
}

// Wait 注册 SIGINT/SIGTERM，收到信号后 cancel 共享 ctx，并等待所有注册的协程退出
func (m *gracefulManager) Wait() {
	sigCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-sigCtx.Done():
	case <-m.ctx.Done():
	}
	stop()

	Info(m.ctx, "Received exit signal, waiting for all jobs to finish...")
	m.cancel() // 通知 Go 启动的协程退出

	// 第二次 Ctrl+C：打印仍在等待的任务并强杀
	secondSignalChan := make(chan os.Signal, 1)
	signal.Notify(secondSignalChan, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(secondSignalChan)

	go func() {
		<-secondSignalChan
		var waitingJobs []string
		m.activeJobs.Range(func(key, value interface{}) bool {
			if jobName, ok := key.(string); ok {
				waitingJobs = append(waitingJobs, jobName)
			}
			return true
		})

		if len(waitingJobs) > 0 {
			Warn(m.ctx, "Force kill requested, jobs still waiting",
				String("waitingJobs", strings.Join(waitingJobs, ", ")),
				Int("count", len(waitingJobs)))
		} else {
			Info(m.ctx, "Force kill requested, no jobs waiting")
		}

		os.Exit(1)
	}()

	m.wg.Wait()
	Info(m.ctx, "All jobs finished, program exiting safely...")
}

// Context 获取上下文，用于传递给 Gin 或 Database
func (m *gracefulManager) Context() context.Context {
	return m.ctx
}

var GracefulManager *gracefulManager = NewGracefulManager()
