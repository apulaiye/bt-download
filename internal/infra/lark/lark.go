package lark

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// 飞书Webhook配置
type LarkWebhookConfig struct {
	WebhookURL string // 飞书机器人的Webhook地址
	Timeout    int    // 超时时间(秒)
}

var (
	config *LarkWebhookConfig
	client *http.Client
	once   sync.Once
)

// 统计信息数据结构
type StatisticsData struct {
	HostName      string
	TotalTasks    int            // 总任务数
	SuccessTasks  int            // 成功任务数
	FailedTasks   int            // 失败任务数
	SuccessRate   float64        // 成功率(%)
	TotalDuration time.Duration  // 总耗时
	Details       map[string]int // 详细分类统计
}

// 飞书文本消息结构
type larkTextMessage struct {
	MsgType string `json:"msg_type"`
	Content struct {
		Text string `json:"text"`
	} `json:"content"`
}

// 飞书卡片消息结构(富文本)
type larkCardMessage struct {
	MsgType string `json:"msg_type"`
	Card    struct {
		Config struct {
			WideScreenMode bool `json:"wide_screen_mode"`
		} `json:"config"`
		Header struct {
			Title struct {
				Tag     string `json:"tag"`
				Content string `json:"content"`
			} `json:"title"`
		} `json:"header"`
		Elements []interface{} `json:"elements"`
	} `json:"card"`
}

// 初始化飞书Webhook客户端
func InitLarkClient(webhookURL string) {
	once.Do(func() {
		config = &LarkWebhookConfig{
			WebhookURL: webhookURL,
			Timeout:    30,
		}
		client = &http.Client{
			Timeout: time.Duration(config.Timeout) * time.Second,
		}
	})

	return

}

// 发送卡片格式统计信息(更美观)
func SendCardStatistics(stats StatisticsData) error {
	// 构建卡片消息
	msg := larkCardMessage{
		MsgType: "interactive",
	}
	msg.Card.Config.WideScreenMode = true

	// 卡片标题
	msg.Card.Header.Title.Tag = "plain_text"
	msg.Card.Header.Title.Content = fmt.Sprintf("%s 任务统计报告", stats.HostName)

	// 卡片内容
	msg.Card.Elements = []interface{}{
		// 时间信息
		map[string]interface{}{
			"tag": "div",
			"text": map[string]interface{}{
				"tag":     "plain_text",
				"content": fmt.Sprintf("统计时间: %s", time.Now().Format("2006-01-02 15:04:05")),
			},
		},
		// 分隔线
		map[string]interface{}{"tag": "hr"},
		// 主要统计数据
		map[string]interface{}{
			"tag": "div",
			"text": map[string]interface{}{
				"tag": "lark_md",
				"content": fmt.Sprintf("总任务数: **%d**\n成功: **%d** (%.2f%%)\n失败: **%d**\n总耗时: **%v**",
					stats.TotalTasks, stats.SuccessTasks, stats.SuccessRate,
					stats.FailedTasks, stats.TotalDuration),
			},
		},
		// 详细分类
		map[string]interface{}{"tag": "hr"},
		map[string]interface{}{
			"tag": "div",
			"text": map[string]interface{}{
				"tag":     "lark_md",
				"content": "**详细分类统计**",
			},
		},
	}

	// 添加详细分类到卡片
	for category, count := range stats.Details {
		msg.Card.Elements = append(msg.Card.Elements, map[string]interface{}{
			"tag": "div",
			"text": map[string]interface{}{
				"tag":     "lark_md",
				"content": fmt.Sprintf("- %s: %d", category, count),
			},
		})
	}

	return send(msg)
}

// 实际发送请求的内部方法
func send(msg interface{}) error {

	// 序列化消息
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("序列化消息失败: %w", err)
	}

	// 发送POST请求
	resp, err := client.Post(config.WebhookURL, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("请求失败，状态码: %d", resp.StatusCode)
	}

	return nil
}
