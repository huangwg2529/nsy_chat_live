package service

import (
	"context"
	"errors"
	"nsy_chat_live/config"
	"nsy_chat_live/dal"
	"nsy_chat_live/model"
	"nsy_chat_live/rep_api"
	"sort"
	"time"

	"github.com/cloudwego/hertz/pkg/common/hlog"
)

func saveChatRooms() error {
	chatRooms, err := rep_api.GetChatRooms()
	if err != nil {
		hlog.Error("GetChatRooms failed, err: %v", err)
		return err
	}
	innerChatRooms, err := dal.GetChatRooms()
	if err != nil {
		hlog.Error("GetChatRooms failed, err: %v", err)
		return err
	}
	chatRoomMap := make(map[string]*dal.ChatRoom)
	for _, chatRoom := range innerChatRooms {
		chatRoomMap[chatRoom.ChatRoomId] = chatRoom
	}
	db := dal.DB()
	newRooms := make([]*dal.ChatRoom, 0)
	for _, chatRoom := range chatRooms {
		if _, ok := chatRoomMap[chatRoom.ChatRoomId]; ok {
			continue
		}
		innerChatRoom := &dal.ChatRoom{
			UserId:      chatRoom.UserId,
			UniqueId:    chatRoom.UserProfile.UniqueId,
			DisplayName: chatRoom.UserProfile.DisplayName,
			ChatRoomId:  chatRoom.ChatRoomId,
			AvatarUrl:   chatRoom.UserProfile.AvatarUrl,
		}
		if err := db.Create(innerChatRoom).Error; err != nil {
			hlog.Errorf("CreateChatRoom failed, err: %v", err)
			return err
		}
		hlog.Infof("CreateChatRoom success, result: %v", innerChatRoom)
		newRooms = append(newRooms, innerChatRoom)
	}
	dal.ReleaseDB()
	for _, chatRoom := range newRooms {
		if err := refreshOldChatMessages(chatRoom.UserId, chatRoom.ChatRoomId); err != nil {
			hlog.Errorf("refreshOldChatMessages failed, err: %v", err)
			return err
		}
	}
	return nil
}

func refreshNewMessages() error {
	ctx := context.Background()
	chatRooms, err := dal.GetChatRooms()
	if err != nil {
		hlog.Error("GetChatRooms failed, err: %v", err)
		return err
	}
	for _, chatRoom := range chatRooms {
		if err := updateNewMessages(ctx, chatRoom.UserId, chatRoom.ChatRoomId); err != nil {
			hlog.Errorf("updateNewMessages failed, err: %v", err)
			return err
		}
	}
	return nil
}

func updateNewMessages(ctx context.Context, uid, roomId string) error {
	existMessages := make([]dal.ChatMessage, 0)
	err := dal.ReadDB().Table(dal.ChatMessage{}.TableName()).
		Where("user_id = ? AND chat_room_id = ?", uid, roomId).
		Order("id desc").
		Limit(1).
		Find(&existMessages).Error
	if err != nil {
		hlog.Errorf("Failed to get chat messages: %v", err)
		return err
	}
	if len(existMessages) == 0 {
		return nil
	}
	nextId := existMessages[0].ChatMessageId
	totalMsg := make([]*model.ListChatMessages, 0)
	end := false
	for !end {
		messages, cursor, err := rep_api.GetChatMessages(ctx, uid, roomId, &nextId, false, 100)
		if err != nil {
			hlog.Errorf("Failed to get chat messages: %v", err)
			return err
		}
		totalMsg = append(totalMsg, messages...)
		nextId = cursor
		if len(messages) == 0 {
			end = true
		}
	}
	if len(totalMsg) == 0 {
		return nil
	}
	hlog.Infof("updateNewMessages, uid: %v, roomId: %v, msg size: %v", uid, roomId, len(totalMsg))
	if err := saveMessage(totalMsg); err != nil {
		hlog.Errorf("Failed to save chat message: %v", err)
		return err
	}
	return nil
}

func refreshOldChatMessages(uid, roomId string) error {
	// 1. 先获取最新一条
	ctx := context.Background()
	messages, cursor, err := rep_api.GetChatMessages(ctx, uid, roomId, nil, false, 1)
	if err != nil {
		hlog.Errorf("Failed to get chat messages: %v", err)
		return err
	}
	if len(messages) == 0 {
		hlog.Errorf("Failed to get chat messages, len(messages) == 0")
		return errors.New("len(messages) == 0")
	}
	// 2. 往前一直获取，直到游标跟已获取消息相同
	nextId := cursor
	totalMsg := make([]*model.ListChatMessages, 0)
	end := false
	for !end {
		messages, cursor, err = rep_api.GetChatMessages(ctx, uid, roomId, &nextId, true, 100)
		if err != nil {
			hlog.Errorf("Failed to get chat messages: %v", err)
			return err
		}
		totalMsg = append(totalMsg, messages...)
		nextId = cursor
		if len(messages) == 0 {
			end = true
		}
	}
	// 3. 保存
	if err := saveMessage(totalMsg); err != nil {
		hlog.Errorf("Failed to save chat message: %v", err)
		return err
	}
	return nil
}

func saveMessage(messages []*model.ListChatMessages) error {
	sort.Slice(messages, func(i, j int) bool {
		return messages[i].Timestamp.Seconds < messages[j].Timestamp.Seconds
	})

	dbMsgList := make([]*dal.ChatMessage, 0)
	err := func() error {
		db := dal.DB()
		defer dal.ReleaseDB()
		for _, msg := range messages {
			dbMsg := &dal.ChatMessage{
				UserId:        msg.UserId,
				DisplayName:   msg.UserProfile.DisplayName,
				ChatRoomId:    msg.ChatRoomId,
				ChatMessageId: msg.ChatMessageId,
				MsgType:       msg.Type,
				Content:       msg.Content,
				ImageUrl:      msg.ImageUrl,
				VideoUrl:      msg.VideoUrl,
				TimeStr:       msg.TimeStr,
				SendTime:      msg.Timestamp.Seconds,
			}
			dbMsgList = append(dbMsgList, dbMsg)
			switch msg.Type {
			case int32(model.ChatMessageType_Text):
				// do nothing
			case int32(model.ChatMessageType_Image):
				hlog.Infof("DownloadImage, imageUrl: %v, time: %v, path: %v, msgId: %v", msg.ImageUrl, time.Unix(msg.Timestamp.Seconds, 0), getMediaPath(msg.UserProfile.DisplayName), msg.ChatMessageId)
				imgPath, err := DownloadImage(msg.ImageUrl, time.Unix(msg.Timestamp.Seconds, 0), getMediaPath(msg.UserProfile.DisplayName), msg.ChatMessageId)
				if err != nil {
					hlog.Errorf("Failed to download image: %v", err)
					return err
				}
				dbMsg.ImagePath = imgPath
			case int32(model.ChatMessageType_Video):
				hlog.Infof("DownloadVideo, videoUrl: %v, time: %v, path: %v, msgId: %v", msg.VideoUrl, time.Unix(msg.Timestamp.Seconds, 0), getMediaPath(msg.UserProfile.DisplayName), msg.ChatMessageId)
				videoPath, err := DownloadVideo(msg.VideoUrl, time.Unix(msg.Timestamp.Seconds, 0), getMediaPath(msg.UserProfile.DisplayName), msg.ChatMessageId)
				if err != nil {
					hlog.Errorf("Failed to download video: %v", err)
					return err
				}
				dbMsg.VideoPath = videoPath
			default:
				hlog.Warnf("Unknown chat message type: %v, msgId: %v", msg.Type, msg.ChatMessageId)
			}
			if err := db.Create(dbMsg).Error; err != nil {
				hlog.Errorf("Failed to save chat message: %v", err)
				return err
			}
		}
		return nil
	}()
	if err != nil {
		return err
	}

	for _, msg := range dbMsgList {
		sendChatEmail(msg)
	}
	return nil
}

func getMediaPath(name string) string {
	return config.GetMediaPath() + "/" + name
}
