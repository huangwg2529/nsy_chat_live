package service

import (
	"crypto/tls"
	"fmt"
	"nsy_chat_live/config"
	"nsy_chat_live/dal"
	"nsy_chat_live/model"
	"time"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"gopkg.in/gomail.v2"
)

var (
	sendCli      *gomail.Dialer
	useEmail     bool
	emailChannel chan *EMailInfo
)

func initEmailSender() {
	if len(config.Conf.Email.SmtpHost) == 0 || len(config.Conf.Email.Sender) == 0 || len(config.Conf.Email.AuthCode) == 0 || len(config.Conf.Email.Receiver) == 0 {
		useEmail = false
		hlog.Infof("email config is empty, useEmail: %v", useEmail)
		return
	}
	// 初始化邮件发送器
	sendCli = gomail.NewDialer(config.Conf.Email.SmtpHost, 587, config.Conf.Email.Sender, config.Conf.Email.AuthCode)
	sendCli.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	emailChannel = make(chan *EMailInfo, 1000)
	useEmail = true
	hlog.Infof("email sender init done, useEmail: %v", useEmail)
	go asyncSendEmail()
}

type EMailInfo struct {
	Title    string
	Content  string
	FilePath string
}

func sendLiveEmail(liveInfo *model.LiveStream, nsyInfo *model.LiveUser, rtmpUrl string, isFandomOnly bool) {
	content := fmt.Sprintf("标题: %s\nrtmp: %s\n", liveInfo.Title, rtmpUrl)
	if isFandomOnly {
		content += "fandom only 且未加入，无法获取链接\n"
	}
	emailChannel <- &EMailInfo{
		Title:   fmt.Sprintf("直播开播: %s", nsyInfo.Info.DisplayName),
		Content: content,
	}
}

func sendLiveEndEmail(nsyName string, fileName string) {
	emailChannel <- &EMailInfo{
		Title:   fmt.Sprintf("直播结束: %s", nsyName),
		Content: fmt.Sprintf("fileName: %s\n", fileName),
	}
}

func sendChatEmail(msg *dal.ChatMessage) {
	if !useEmail {
		return
	}
	info := &EMailInfo{}
	if msg.MsgType == int32(model.ChatMessageType_Image) {
		info.Title = fmt.Sprintf("图片消息: %v", msg.DisplayName)
		info.FilePath = msg.ImagePath
	} else if msg.MsgType == int32(model.ChatMessageType_Video) {
		info.Title = fmt.Sprintf("视频消息: %v", msg.DisplayName)
		info.FilePath = msg.VideoPath
	} else {
		return
	}
	// 半小时往前的历史的不发，自己去文件夹看就行
	if msg.SendTime < time.Now().Unix()-1800 {
		hlog.Infof("don't send email, msg is too old, msgId: %v, sendTime: %v", msg.ChatMessageId, msg.SendTime)
		return
	}
	// 查询前后10分钟内的几条消息，一并发送
	msgList := make([]*dal.ChatMessage, 0)
	err := dal.ReadDB().Table(msg.TableName()).
		Where("chat_room_id = ? AND user_id != ?", msg.ChatRoomId, msg.UserId).
		Where("send_time >= ? AND send_time <= ?", msg.SendTime-600, msg.SendTime+600).
		Order("id").
		Limit(100).
		Find(&msgList).Error
	if err != nil {
		hlog.Errorf("查询时间相近的几条消息失败: %v", err)
		return
	}
	for _, m := range msgList {
		content := m.Content
		if m.MsgType == int32(model.ChatMessageType_Image) {
			content = "图片"
		} else if m.MsgType == int32(model.ChatMessageType_Video) {
			content = "视频"
		}
		info.Content += fmt.Sprintf("%s\n", content)
	}
	if len(info.Content) == 0 {
		info.Content = "暂未获取上下文"
	}
	hlog.Infof("sendChatEmail with local file: %v, msgId: %v", info.FilePath, msg.ChatMessageId)
	emailChannel <- info
}

func asyncSendEmail() {
	for info := range emailChannel {
		msg := gomail.NewMessage()
		msg.SetHeader("From", config.Conf.Email.Sender)
		msg.SetHeader("To", config.Conf.Email.Receiver)
		msg.SetHeader("Subject", info.Title)
		msg.SetBody("text/plain", info.Content)
		if len(info.FilePath) > 0 {
			msg.Attach(info.FilePath)
		}
		hlog.Infof("发送邮件: %v", info)
		if err := sendCli.DialAndSend(msg); err != nil {
			hlog.Errorf("发送邮件失败: %v", err)
		}
	}
}
