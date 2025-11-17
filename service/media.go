package service

import (
	"fmt"
	"net/url"
	"nsy_chat_live/rep_api"
	"os"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/common/hlog"
)

func getImgFileName(mediaUrl string, imgTime time.Time, msgId string) (name string, path string, err error) {
	var u *url.URL
	u, err = url.Parse(mediaUrl)
	if err != nil {
		err = fmt.Errorf("failed to parse url: %v", err)
		return
	}
	paths := strings.Split(u.Path, "/")
	name = fmt.Sprintf("%s-%s-%s", imgTime.Format("2006-01-02"), msgId, paths[len(paths)-1])
	if !strings.Contains(name, ".") {
		name = fmt.Sprintf("%s.jpeg", name)
	}
	path = fmt.Sprintf("%d/%d", imgTime.Year(), imgTime.Month())
	return
}

func DownloadImage(mediaUrl string, imgTime time.Time, pathPrefix string, msgId string) (string, error) {
	name, path, err := getImgFileName(mediaUrl, imgTime, msgId)
	if err != nil {
		return "", fmt.Errorf("failed to get img file name: %v", err)
	}
	path = fmt.Sprintf("%s/%s", pathPrefix, path)
	return downloadMedia(mediaUrl, path, name)
}

func DownloadVideo(mediaUrl string, videoTime time.Time, pathPrefix string, msgId string) (string, error) {
	path := fmt.Sprintf("%d/%d", videoTime.Year(), videoTime.Month())
	path = fmt.Sprintf("%s/%s", pathPrefix, path)
	name := fmt.Sprintf("%s_%s.mp4", videoTime.Format("2006-01-02"), msgId)
	return downloadMedia(mediaUrl, path, name)
}

func downloadMedia(media, path, name string) (string, error) {
	body, err := rep_api.Get(media)
	if err != nil {
		return "", fmt.Errorf("failed to get media: %v", err)
	}
	hlog.Infof("download media: %s, path: %s, name: %s", media, path, name)
	err = os.MkdirAll(path, 0755)
	if err != nil {
		return "", fmt.Errorf("failed to create dir: %v", err)
	}
	err = os.WriteFile(fmt.Sprintf("%s/%s", path, name), body, 0644)
	if err != nil {
		return "", fmt.Errorf("failed to write file: %v", err)
	}
	return fmt.Sprintf("%s/%s", path, name), nil
}
