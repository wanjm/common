package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/wanjm/common"
)

// ==========================================
// Main 业务逻辑
// ==========================================

func main() {
	// 初始化管理器
	manager := common.GracefulManager

	// --- 模拟业务后台协程 ---
	manager.Go("数据清理任务", func(ctx context.Context) {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done(): // 监听退出信号
				fmt.Println("-> 数据清理任务检测到退出信号，正在清理资源...")
				time.Sleep(5 * time.Second) // 模拟清理耗时
				return
			case t := <-ticker.C:
				fmt.Printf("   正在执行后台任务: %v\n", t.Format("15:04:05"))
				time.Sleep(5 * time.Second) // 模拟任务耗时
				fmt.Printf("   后台任务执行完成: %v\n", t.Format("15:04:05"))
			}
		}
	})

	// --- 核心点 5: Gin 集成 ---
	// Gin 的启动和退出需要特殊处理，因为它是一个阻塞服务

	r := gin.Default()
	r.GET("/", func(c *gin.Context) {
		time.Sleep(5 * time.Second) // 模拟一个耗时的请求
		c.String(200, "Hello World")
	})

	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	// 启动 HTTP 服务 (不使用 manager.Go，因为我们不需要它去 wait ListenAndServe)
	// 我们需要的是由 manager 通知 server 去 Shutdown
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// 注册一个特殊的“关闭器”协程给 Manager
	// 当 Manager 收到 Ctrl+C 时，这个协程会被通知，然后它负责关闭 Gin
	manager.Go("HTTP Server Shutdown", func(ctx context.Context) {
		// 这里我们其实是在等待 ctx.Done()，因为 manager.Go 内部调用的 fn 会立即执行
		// 但我们需要阻塞在这里等待信号
		<-ctx.Done()

		fmt.Println("-> 正在关闭 HTTP Server...")

		// 设定一个超时时间，强制结束未完成的请求（例如 5 秒）
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Fatal("Server Shutdown Forced:", err)
		}
		fmt.Println("-> HTTP Server 已优雅停止")
	})

	fmt.Println("程序已启动 (PID: ", os.Getpid(), ")，按 Ctrl+C 退出...")

	// 阻塞主线程，直到所有工作完成
	manager.Wait()
}
