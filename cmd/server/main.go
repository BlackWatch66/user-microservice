package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	grpcApi "github.com/blackwatch66/user-microservice/api/grpc"
	pb "github.com/blackwatch66/user-microservice/api/grpc/proto"
	httpHandler "github.com/blackwatch66/user-microservice/api/http/handler"
	httpMiddleware "github.com/blackwatch66/user-microservice/api/http/middleware"
	"github.com/blackwatch66/user-microservice/config"
	"github.com/blackwatch66/user-microservice/internal/database"
	"github.com/blackwatch66/user-microservice/internal/redis"
	"github.com/blackwatch66/user-microservice/internal/service"
	"github.com/gin-gonic/gin"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

func main() {
	// 加载配置
	cfg := config.LoadConfig()

	// 初始化数据库
	db, err := database.InitDatabase(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// 初始化 Redis
	rdb, err := redis.InitRedis(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
	if err != nil {
		log.Fatalf("Failed to initialize redis: %v", err)
	}

	// 初始化 Service
	userService := service.NewUserService(db, rdb, cfg)

	// 初始化 Gin Engine
	router := gin.Default()

	// 设置 JWT 中间件
	authMiddleware := httpMiddleware.AuthMiddleware(cfg)

	// 初始化 HTTP Handler 并注册路由
	userHttpHandler := httpHandler.NewUserHandler(userService)
	userHttpHandler.RegisterRoutes(router, authMiddleware)

	// 创建 HTTP Server
	httpServer := &http.Server{
		Addr:    ":" + cfg.HTTPPort,
		Handler: router,
	}

	// 创建 gRPC Server
	grpcServer := grpc.NewServer()
	userGrpcServer := grpcApi.NewUserServer(userService)
	pb.RegisterUserServiceServer(grpcServer, userGrpcServer)

	// 使用 errgroup 管理 goroutines 和错误处理
	g, ctx := errgroup.WithContext(context.Background())

	// 启动 HTTP Server
	g.Go(func() error {
		log.Printf("Starting HTTP server on port %s\n", cfg.HTTPPort)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("HTTP server ListenAndServe error: %v", err)
			return err
		}
		log.Println("HTTP server stopped gracefully.")
		return nil
	})

	// 启动 gRPC Server
	g.Go(func() error {
		lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
		if err != nil {
			log.Printf("Failed to listen for gRPC: %v", err)
			return err
		}
		log.Printf("Starting gRPC server on port %s\n", cfg.GRPCPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Printf("gRPC server Serve error: %v", err)
			return err
		}
		log.Println("gRPC server stopped gracefully.")
		return nil
	})

	// 监听退出信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// 等待退出信号或 goroutine 错误
	select {
	case <-quit:
		log.Println("Received shutdown signal. Shutting down servers...")
	case <-ctx.Done():
		log.Println("Context cancelled. Shutting down servers due to error...")
	}

	// 设置超时上下文用于优雅关闭
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 关闭 HTTP Server
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}

	// 关闭 gRPC Server
	grpcServer.GracefulStop()
	log.Println("gRPC server stopped.")

	// 等待所有 goroutines 完成
	if err := g.Wait(); err != nil {
		log.Printf("Server shutdown completed with error: %v", err)
	} else {
		log.Println("Server shutdown completed successfully.")
	}
}
