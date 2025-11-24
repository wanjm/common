package common

import (
	"context"
	"os/signal"
	"sync"
	"syscall"
)

// ==========================================
// 1. 封装优雅退出管理器
// ==========================================

type gracefulManager struct {
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// NewGracefulManager 创建管理器，监听 SIGINT (Ctrl+C) 和 SIGTERM
func NewGracefulManager() *gracefulManager {
	// 核心点 1 & 2: 使用 Background 作为根，并监听系统信号
	// NotifyContext 会在收到信号时自动 cancel ctx
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	return &gracefulManager{
		ctx:    ctx,
		cancel: cancel,
	}
}

// Go 核心点 3: 启动协程帮助函数
// 自动注册 (Add) 和 解注册 (Done)
func (m *gracefulManager) Go(jobName string, fn func(ctx context.Context)) {
	m.wg.Add(1)
	go func() {
		defer m.wg.Done() // 协程退出时取消注册
		defer Recover(m.ctx, jobName)
		Info(m.ctx, "Job started", String("jobName", jobName))
		fn(m.ctx) // 将带有信号监听的 ctx 传递给业务逻辑
		Info(m.ctx, "Job finished", String("jobName", jobName))
	}()
}

// Wait 核心点 4: 等待所有注册的协程退出
func (m *gracefulManager) Wait() {
	// 等待信号发生（用户按 Ctrl+C）
	<-m.ctx.Done()
	Info(m.ctx, "Received exit signal, waiting for all jobs to finish...")

	// 恢复默认信号处理（如果用户再次按 Ctrl+C，立即强杀）
	m.cancel()

	// 等待所有 Add 的协程 Done
	m.wg.Wait()
	Info(m.ctx, "All jobs finished, program exiting safely...")
}

// Context 获取上下文，用于传递给 Gin 或 Database
func (m *gracefulManager) Context() context.Context {
	return m.ctx
}

var GracefulManager *gracefulManager = NewGracefulManager()
