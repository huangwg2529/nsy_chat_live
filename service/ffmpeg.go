package service

import (
	"fmt"
	"nsy_chat_live/config"
	"nsy_chat_live/utils"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/common/hlog"
)

var ffmpegPath string

func initFfmpeg() {
	if _, err := os.Stat(config.Conf.FfmpegPath); err == nil {
		ffmpegPath = config.Conf.FfmpegPath
	} else if _, err := os.Stat("./ffmpeg.exe"); err == nil {
		ffmpegPath = "./ffmpeg.exe"
	} else if _, err := os.Stat("./ffmpeg"); err == nil {
		ffmpegPath = "./ffmpeg"
	} else {
		if utils.IsWindows() {
			ffmpegPath = "ffmpeg.exe"
		} else {
			ffmpegPath = "ffmpeg"
		}
	}
}

func startFfmpegWatcher() {
	liveRecordChannel = make(chan *NsyLiveInfo, 1000)
	initFfmpeg()
	hlog.Infof("ffmpeg path: %s, Starting ffmpeg watcher...", ffmpegPath)
	go func() {
		for nsyLiveInfo := range liveRecordChannel {
			if err := startFfmpegRecord(nsyLiveInfo); err != nil {
				hlog.Errorf("Error starting ffmpeg record: %v", err)
				liveRecordChannel <- nsyLiveInfo
			}
		}
	}()
}

func startFfmpegRecord(nsyLiveInfo *NsyLiveInfo) error {
	defer func() {
		if r := recover(); r != nil {
			hlog.Errorf("Recovered in ffmpeg record: %v", r)
		}
	}()
	outputFile := fmt.Sprintf("%s_%s.mp4",
		nsyLiveInfo.Name, time.Now().Format("200601021504"))
	for i := 1; i < 100; i++ {
		if _, err := os.Stat(outputFile); err == nil {
			hlog.Warnf("file %s already exist? err: %v, try %d", outputFile, err, i)
			outputFile = fmt.Sprintf("%s_%s_%d.mp4",
				nsyLiveInfo.Name, time.Now().Format("200601021504"), i)
		} else {
			break
		}
	}
	path := config.GetLivePath()
	logFileName := strings.TrimSuffix(outputFile, ".mp4") + ".txt"
	if len(path) > 0 {
		_ = os.MkdirAll(path, 0755)
		outputFile = path + "/" + outputFile
		logFileName = path + "/" + logFileName
	}
	logFile, err := os.Create(logFileName)
	if err != nil {
		return fmt.Errorf("创建日志文件失败: %v", err)
	}
	cmd := exec.Command(
		//"cmd", "/C", "start", "cmd", "/K",
		ffmpegPath,
		"-i", nsyLiveInfo.RtmpUrl,
		"-c", "copy",
		outputFile,
	)
	hlog.Infof("start recording by command: %s", cmd.String())
	go func() {
		defer func() {
			if r := recover(); r != nil {
				hlog.Errorf("Recovered in ffmpeg record: %v", r)
			}
		}()
		defer logFile.Close()
		hlog.Infof("ffmpeg 日志: %s", logFile.Name())
		cmd.Stdout = logFile
		cmd.Stderr = logFile

		if err := cmd.Start(); err != nil {
			hlog.Errorf("ffmpeg err: " + err.Error())
		}

		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		done := make(chan error, 1)
		go func() { done <- cmd.Wait() }()

		hlog.Infof("[%s] 录制开始:", outputFile)

		for {
			select {
			case <-ticker.C:
				hlog.Infof("[%s] 录制中...日志: %s", outputFile, logFile.Name())
			case err := <-done: // 进程结束处理
				hlog.Infof("[%s] 录制结束:", outputFile)
				sendLiveEndEmail(nsyLiveInfo.Name, outputFile)
				if err != nil {
					hlog.Errorf("[%s] 录制错误: %v", outputFile, err)
				}
				return
			}
		}
	}()
	return nil
}
