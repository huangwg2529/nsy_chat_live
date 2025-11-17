package handler

import (
	"context"
	"nsy_chat_live/dal"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

func HandleGetChatRooms(ctx context.Context, c *app.RequestContext) {
	resp := &Resp{}
	rooms, err := dal.GetChatRooms()
	if err != nil {
		c.JSON(consts.StatusOK, BadResp(err.Error()))
	}
	resp.Data = rooms
	c.JSON(consts.StatusOK, resp)
}

func HandleGetChatMessages(ctx context.Context, c *app.RequestContext) {
	//req := struct {
	//}{}
	c.JSON(consts.StatusOK, utils.H{"message": "pong"})
}
