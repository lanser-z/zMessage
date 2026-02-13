package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"zmessage/server/api"
	"zmessage/server/dal"
	"zmessage/server/modules/media"
	"zmessage/server/modules/message"
	"zmessage/server/modules/user"
	"zmessage/server/ws"
)

func main() {
	dataDir := "./data"
	if len(os.Args) > 1 {
		dataDir = os.Args[1]
	}

	if err := os.MkdirAll(dataDir, 0755); err != nil {
		log.Fatalf("创建数据目录失败: %v", err)
	}

	dalMgr, err := dal.NewManager(dataDir)
	if err != nil {
		log.Fatalf("初始化DAL失败: %v", err)
	}
	defer dalMgr.Close()

	userSvc := user.NewService(dalMgr, "test-secret")
	mediaSvc := media.NewService(dalMgr, dataDir+"/media")
	msgSvc := message.NewService(dalMgr)
	wsMgr := ws.NewManager(msgSvc, userSvc)

	r := gin.Default()

	r.Static("/assets", "./client/assets")
	r.GET("/", func(c *gin.Context) {
		c.File("./client/index.html")
	})

	api.RegisterAuthRoutes(r, userSvc)
	api.RegisterUsersRoutes(r, userSvc)
	api.RegisterConversationRoutes(r, msgSvc, nil)
	api.RegisterMessageRoutes(r, msgSvc, nil)
	api.RegisterMediaRoutes(r, mediaSvc)

	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	r.GET("/ws", func(c *gin.Context) {
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Printf("WebSocket升级失败: %v", err)
			return
		}
		wsMgr.HandleConnection(conn)
	})

	addr := ":9405"
	if len(os.Args) > 2 {
		addr = os.Args[2]
	}

	server := &http.Server{
		Addr:           addr,
		Handler:        r,
		ReadTimeout:     10 * time.Second,
		WriteTimeout:    10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("服务器启动，监听 %s", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("服务器错误: %v", err)
		}
	}()

	<-shutdown
	log.Println("正在关闭服务器...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = server.Shutdown(ctx)
	log.Println("服务器已关闭")
}
