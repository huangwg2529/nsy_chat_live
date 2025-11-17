package config

import (
	"fmt"
	"nsy_chat_live/utils"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Proxy struct {
		Host string `yaml:"host"`
		Port int    `yaml:"port"`
	} `yaml:"proxy"`
	RefreshToken   string `yaml:"refresh_token"`
	FfmpegPath     string `yaml:"ffmpeg_path"`
	MediaPathWin   string `yaml:"media_path_win"`
	MediaPathLinux string `yaml:"media_path_linux"`
	Email          struct {
		SmtpHost string `yaml:"smtp_host"`
		Sender   string `yaml:"sender"`
		AuthCode string `yaml:"auth_code"`
		Receiver string `yaml:"receiver"`
	} `yaml:"email"`
}

var (
	Conf Config
	path string
)

func LoadConfig() error {
	data, err := os.ReadFile("config.yaml")
	if err != nil {
		return fmt.Errorf("读取配置文件失败: %v", err)
	}
	if err := yaml.Unmarshal(data, &Conf); err != nil {
		return fmt.Errorf("解析YAML配置失败: %v", err)
	}
	if Conf.RefreshToken == "" {
		return fmt.Errorf("refresh_token 不能为空")
	}
	if utils.IsWindows() {
		path = Conf.MediaPathWin
	} else {
		path = Conf.MediaPathLinux
	}
	if path == "" {
		path = "./media"
	}
	fmt.Printf("load config: %v,\npath: %s\n", Conf, path)
	return nil
}

func GetMediaPath() string {
	return path
}

func GetLivePath() string {
	return GetMediaPath() + "/live"
}
