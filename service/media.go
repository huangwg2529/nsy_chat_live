package service

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"replive/rep_api"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/common/hlog"
)

var fetchMedia = rep_api.Get

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

func DownloadProfileMedia(mediaUrl string, mediaTime time.Time, pathPrefix string, owner string, kind string) (string, error) {
	name, err := getProfileMediaFileName(mediaUrl, mediaTime, owner, kind)
	if err != nil {
		return "", fmt.Errorf("failed to get profile media file name: %v", err)
	}
	path := filepath.Join(pathPrefix, sanitizeFileName(owner), mediaTime.Format("2006"), mediaTime.Format("01"), mediaTime.Format("02"))
	return downloadMedia(mediaUrl, path, name)
}

func getProfileMediaFileName(mediaUrl string, mediaTime time.Time, owner string, kind string) (string, error) {
	u, err := url.Parse(mediaUrl)
	if err != nil {
		return "", fmt.Errorf("failed to parse url: %v", err)
	}
	paths := strings.Split(u.Path, "/")
	base := strings.TrimSpace(paths[len(paths)-1])
	if base == "" || !strings.Contains(base, ".") {
		base = "image.jpeg"
	}
	ext := filepath.Ext(base)
	base = strings.TrimSuffix(base, ext)
	sum := sha1.Sum([]byte(mediaUrl))
	hash := hex.EncodeToString(sum[:])[:10]
	name := fmt.Sprintf(
		"%s_%s_%s_%s_%s%s",
		mediaTime.Format("2006-01-02"),
		sanitizeFileName(owner),
		sanitizeFileName(kind),
		hash,
		sanitizeFileName(base),
		ext,
	)
	return name, nil
}

func sanitizeFileName(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return "unknown"
	}
	replacer := strings.NewReplacer(
		"\\", "_",
		"/", "_",
		":", "_",
		"*", "_",
		"?", "_",
		"\"", "_",
		"<", "_",
		">", "_",
		"|", "_",
	)
	return replacer.Replace(s)
}

func downloadMedia(media, path, name string) (string, error) {
	err := os.MkdirAll(path, 0755)
	if err != nil {
		return "", fmt.Errorf("failed to create dir: %v", err)
	}
	filePath := filepath.Join(path, name)
	if info, err := os.Stat(filePath); err == nil && !info.IsDir() {
		hlog.Infof("skip existing media: %s", filePath)
		return filePath, nil
	} else if err != nil && !os.IsNotExist(err) {
		return "", fmt.Errorf("failed to stat media file: %v", err)
	}

	body, err := fetchMedia(media)
	if err != nil {
		return "", fmt.Errorf("failed to get media: %v", err)
	}
	hlog.Infof("download media: %s, path: %s, name: %s", media, path, name)
	err = os.WriteFile(filePath, body, 0644)
	if err != nil {
		return "", fmt.Errorf("failed to write file: %v", err)
	}
	return filePath, nil
}
