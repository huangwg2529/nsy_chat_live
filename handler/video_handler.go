package handler

import (
	"context"
	"io"
	"nsy_chat_live/config"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

type TokenBucketLimiter struct {
	capacity       int64
	rate           int64
	tokens         int64
	lastRefillTime time.Time
	mu             sync.Mutex
}

func NewTokenBucketLimiter(rate, capacity int64) *TokenBucketLimiter {
	return &TokenBucketLimiter{
		capacity:       capacity,
		rate:           rate,
		tokens:         capacity,
		lastRefillTime: time.Now(),
	}
}

func (t *TokenBucketLimiter) Take(n int64) {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(t.lastRefillTime)
	tokensToAdd := int64(elapsed.Seconds() * float64(t.rate))
	if tokensToAdd > 0 {
		t.tokens = t.tokens + tokensToAdd
		if t.tokens > t.capacity {
			t.tokens = t.capacity
		}
		t.lastRefillTime = now
	}

	if t.tokens < n {
		needed := n - t.tokens
		waitTime := time.Duration(float64(needed)/float64(t.rate)) * time.Second
		time.Sleep(waitTime)
		t.tokens = 0
	} else {
		t.tokens -= n
	}
}

var downloadLimiter = NewTokenBucketLimiter(2*1024*1024, 2*1024*1024)

func HandleDownloadVideo(c context.Context, ctx *app.RequestContext) {
	filename := ctx.Query("filename")
	if filename == "" {
		ctx.JSON(consts.StatusBadRequest, map[string]interface{}{
			"code": 400,
			"msg":  "文件名不能为空",
		})
		return
	}

	videoPath := filepath.Join(config.GetLivePath(), filename)
	fileInfo, err := os.Stat(videoPath)
	if err != nil || fileInfo.IsDir() {
		ctx.JSON(consts.StatusNotFound, map[string]interface{}{
			"code": 404,
			"msg":  "视频文件不存在",
		})
		return
	}
	fileSize := fileInfo.Size()

	ctx.Response.Header.Set("Content-Type", "application/octet-stream")
	ctx.Response.Header.Set("Content-Disposition", `attachment; filename="`+filename+`"`)
	ctx.Response.Header.Set("Content-Length", strconv.FormatInt(fileSize, 10))
	ctx.Response.SetStatusCode(consts.StatusOK)

	file, err := os.Open(videoPath)
	if err != nil {
		ctx.JSON(consts.StatusInternalServerError, map[string]interface{}{
			"code": 500,
			"msg":  "无法打开视频文件",
		})
		return
	}
	defer file.Close()

	buf := make([]byte, 1024*1024)
	for {
		n, err := file.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			return
		}
		if n == 0 {
			continue
		}

		downloadLimiter.Take(int64(n))

		_, err = ctx.Response.BodyWriter().Write(buf[:n])
		if err != nil {
			return
		}
	}
}
