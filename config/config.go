package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Proxy struct {
		Host string `yaml:"host"`
		Port int    `yaml:"port"`
	} `yaml:"proxy"`
	RefreshToken string `yaml:"refresh_token"`
	FfmpegPath   string `yaml:"ffmpeg_path"`
}

var (
	Conf Config
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
	fmt.Printf("load config: %v\n", Conf)
	return nil
}
