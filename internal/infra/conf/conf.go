package conf

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Qbserver struct {
	Name     string `yaml:"name"`
	Host     string `yaml:"host"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}
type Torrent struct {
	TorrentDir          string `yaml:"torrentDir"`
	LimitCount          int    `yaml:"limitCount"`
	StoreSaveFilePath   string `yaml:"storeSaveFilePath"`
	DownloadingRepoPath string `yaml:"downloadingRepoPath"`
	VideoRepoPath       string `yaml:"videoRepoPath"`
	TosBucketDir        string `yaml:"tosBucketDir"`
}

type Tos struct {
	Bucket   string `yaml:"bucket"`
	AK       string `yaml:"ak"`
	SK       string `yaml:"sk"`
	Endpoint string `yaml:"endpoint"`
}

type Lake struct {
	Webhook string `yaml:"webhook"`
}

// 定义与JSON结构对应的结构体
type Config struct {
	Qbserver Qbserver `yaml:"qbserver"`
	Torrent  Torrent  `yaml:"torrent"`
	Tos      Tos      `yaml:"tos"`
	Lake     Lake     `yaml:"lake"`
	Features []string `yaml:"features"`
}

var (
	ConfigFile = "conf/config.yaml"
)

var config Config

func Init() error {
	if os.Getenv("DEBUG") == "true" {
		ConfigFile = "/Users/tiny/repo/bt-download/conf/config.yaml"
	}

	fmt.Println("初始化配置")

	// 读取文件内容
	data, err := os.ReadFile(ConfigFile)
	if err != nil {
		fmt.Printf("读取文件失败: %v\n", err)
		return err
	}

	// 解析JSON

	err = yaml.Unmarshal(data, &config)
	if err != nil {
		fmt.Printf("解析JSON失败: %v\n", err)
		return err
	}

	// 使用解析后的数据
	fmt.Printf("服务器地址: %s\n", config.Qbserver.Host)
	fmt.Printf("用户名: %s\n", config.Qbserver.Username)
	fmt.Printf("种子目录: %s\n", config.Torrent.TorrentDir)
	fmt.Printf("保存目录: %s\n", config.Torrent.StoreSaveFilePath)
	fmt.Printf("下载目录: %s\n", config.Torrent.DownloadingRepoPath)
	fmt.Printf("视频目录: %s\n", config.Torrent.VideoRepoPath)
	fmt.Printf("限制数量: %d\n", config.Torrent.LimitCount)

	return nil
}

func Get() *Config {
	return &config
}
