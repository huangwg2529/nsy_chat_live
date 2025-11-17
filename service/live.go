package service

import (
	"fmt"
	"net/url"
	"nsy_chat_live/model"
	"nsy_chat_live/rep_api"
	"sync"
	"time"

	"github.com/cloudwego/hertz/pkg/common/hlog"
)

type NsyLiveInfo struct {
	*model.LiveStream
	Name    string
	RtmpUrl string
}

var (
	knownLives        sync.Map
	sendLives         sync.Map
	liveRecordChannel chan *NsyLiveInfo
)

func checkLive() error {
	// 大半夜的不轮询了
	if hour := time.Now().Hour(); hour < 7 && hour >= 1 {
		if hour == 4 && time.Now().Minute() == 0 {
			knownLives.Clear()
			sendLives.Clear()
		}
		return nil
	}
	msgResp, err := rep_api.GetStreamingLive()
	if err != nil {
		return fmt.Errorf("failed to get streaming live: %v", err)
	}
	for i, live := range msgResp.LiveInfo {
		if _, exist := knownLives.Load(live.LiveId); exist {
			continue
		}
		isFandomOnly := len(live.WebrtcUrl) == 0
		nsyInfo := msgResp.UserProfile[i]
		rtmpUrl, err := parseRtmpUrl(live.WebrtcUrl)
		if err != nil {
			return fmt.Errorf("failed to parse url: %v", err)
		}
		if _, ok := sendLives.Load(live.LiveId); !ok {
			sendLiveEmail(live, nsyInfo, rtmpUrl, isFandomOnly)
			sendLives.Store(live.LiveId, true)
		}
		if isFandomOnly {
			hlog.Warnf("live %s is fandom only", nsyInfo.Info.DisplayName)
		}
		nsyLiveInfo := &NsyLiveInfo{
			LiveStream: live,
			Name:       nsyInfo.Info.DisplayName,
			RtmpUrl:    rtmpUrl,
		}
		hlog.Infof("%s start live, title: %s, \nrtmp url: %s", nsyInfo.Info.DisplayName, live.Title, rtmpUrl)
		liveRecordChannel <- nsyLiveInfo
		knownLives.Store(nsyLiveInfo.LiveId, nsyLiveInfo)
	}
	return nil
}

func parseRtmpUrl(webrtcUrl string) (string, error) {
	u, err := url.Parse(webrtcUrl)
	if err != nil {
		return "", err
	}
	u.Scheme = "rtmp"
	queryParams := u.Query()
	keepParams := map[string]bool{
		"txSecret": true,
		"txTime":   true,
	}
	newQueryParams := url.Values{}
	for key, values := range queryParams {
		if keepParams[key] {
			for _, value := range values {
				newQueryParams.Add(key, value)
			}
		}
	}
	u.RawQuery = newQueryParams.Encode()
	return u.String(), nil
}
