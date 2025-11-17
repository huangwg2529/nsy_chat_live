package main

import (
	"context"
	"fmt"
	"nsy_chat_live/config"
	"nsy_chat_live/dal"
	"nsy_chat_live/handler"
	"nsy_chat_live/rep_api"
	"nsy_chat_live/service"
	"os"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

func main() {
	Init()

	h := server.Default()
	f, err := os.Create(fmt.Sprintf("replive_%v.log", time.Now().Format("200601021504")))
	if err != nil {
		panic(err)
	}
	defer f.Close()

	hlog.SetOutput(f)
	hlog.SetLevel(hlog.LevelInfo)

	service.Init()

	registerRoutes(h)

	for {
		time.Sleep(time.Minute * 1)
		hlog.Infof("listening...")
	}

	//h.Spin()
	//ffmpeg -i "rtmp://lvplay.rep_api.com/rep_api/4e20d62f-47da-4dca-8364-6e2cd3574f28?txSecret=e415ac573fd7d4e274d575584c0b52a842f6a09e44a9ccf2128eb1f97db29ffd&txTime=6A7735BD" -c copy ..\..\output.ts
}

func Init() {
	if err := config.LoadConfig(); err != nil {
		panic(err)
	}
	if err := rep_api.InitHttp(); err != nil {
		panic(err)
	}
	if err := dal.InitDB(); err != nil {
		panic(err)
	}
}

func registerRoutes(h *server.Hertz) {
	h.GET("/ping", func(ctx context.Context, c *app.RequestContext) {
		c.JSON(consts.StatusOK, utils.H{"message": "pong"})
	})

	chatGroup := h.Group("/api/chat")
	{
		chatGroup.GET("/rooms", handler.HandleGetChatRooms)
		chatGroup.GET("/messages", handler.HandleGetChatMessages)
	}

	videoGroup := h.Group("/api/video")
	{
		videoGroup.GET("/download", handler.HandleDownloadVideo)
	}

}
