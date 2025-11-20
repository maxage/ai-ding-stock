package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// Notifier 通知器接口
type Notifier interface {
	SendSignal(signal *TradingSignal) error
	SendMessage(message string) error
}

// TradingSignal 交易信号
type TradingSignal struct {
	StockCode     string                 `json:"stock_code"`               // 股票代码
	StockName     string                 `json:"stock_name"`               // 股票名称
	Signal        string                 `json:"signal"`                   // 信号类型: BUY/SELL/HOLD
	Price         float64                `json:"price"`                    // 当前价格
	Confidence    int                    `json:"confidence"`               // 信心度 (0-100)
	Reasoning     string                 `json:"reasoning"`                // 推理原因
	TargetPrice   float64                `json:"target_price"`             // 目标价格
	StopLoss      float64                `json:"stop_loss"`                // 止损价格
	RiskReward    string                 `json:"risk_reward"`              // 风险回报比
	Timestamp     time.Time              `json:"timestamp"`                // 时间戳
	TechnicalData map[string]interface{} `json:"technical_data,omitempty"` // 技术指标数据
	
	// 新增：持仓止盈止损价格（持仓模式下有效）
	PositionProfitTarget float64                `json:"position_profit_target,omitempty"` // 持仓止盈价
	PositionStopLoss     float64                `json:"position_stop_loss,omitempty"`     // 持仓止损价
	PositionInfo         map[string]interface{} `json:"position_info,omitempty"`          // 持仓信息（可选）
}

// DingTalkNotifier 钉钉通知器
type DingTalkNotifier struct {
	WebhookURL string
	Secret     string // 加签密钥（可选）
}

// NewDingTalkNotifier 创建钉钉通知器
func NewDingTalkNotifier(webhookURL string, secret string) *DingTalkNotifier {
	return &DingTalkNotifier{
		WebhookURL: webhookURL,
		Secret:     secret,
	}
}

// SendSignal 发送交易信号到钉钉
func (d *DingTalkNotifier) SendSignal(signal *TradingSignal) error {
	// 构建Markdown格式的消息
	markdown := d.formatSignalMarkdown(signal)

	// 钉钉消息格式
	message := map[string]interface{}{
		"msgtype": "markdown",
		"markdown": map[string]string{
			"title": fmt.Sprintf("【%s】%s %s", signal.Signal, signal.StockName, signal.StockCode),
			"text":  markdown,
		},
		"at": map[string]interface{}{
			"isAtAll": false,
		},
	}

	return d.sendRequest(message)
}

// SendMessage 发送普通消息到钉钉
func (d *DingTalkNotifier) SendMessage(message string) error {
	msg := map[string]interface{}{
		"msgtype": "text",
		"text": map[string]string{
			"content": message,
		},
	}
	return d.sendRequest(msg)
}

// getSignalText 获取信号的中文显示文本
func getSignalText(signal string) string {
	switch signal {
	case "BUY":
		return "买入"
	case "SELL":
		return "卖出"
	case "HOLD":
		return "持有"
	default:
		return signal
	}
}

// formatSignalMarkdown 格式化信号为Markdown
func (d *DingTalkNotifier) formatSignalMarkdown(signal *TradingSignal) string {
	var emoji string
	switch signal.Signal {
	case "BUY":
		emoji = "🚀"
	case "SELL":
		emoji = "⚠️"
	case "HOLD":
		emoji = "⏸️"
	default:
		emoji = "📊"
	}

	// 获取信号中文文本
	signalText := getSignalText(signal.Signal)

	// 构建标题和系统标识
	markdown := fmt.Sprintf("# %s %s信号 - %s(%s)\n\n", emoji, signalText, signal.StockName, signal.StockCode)
	markdown += fmt.Sprintf("**【AI股票分析系统】**\n\n")
	markdown += fmt.Sprintf("---\n\n")
	
	// 1️⃣ 核心指标区域
	markdown += fmt.Sprintf("**1️⃣  核心指标**\n\n")
	markdown += fmt.Sprintf("💰 **当前价格**: %.2f元\n\n", signal.Price)
	markdown += fmt.Sprintf("📈 **信心度**: %d%%\n\n", signal.Confidence)

	// 2️⃣ 交易建议区域
	if signal.TargetPrice > 0 || signal.StopLoss > 0 || signal.RiskReward != "" {
		markdown += fmt.Sprintf("**2️⃣  交易建议**\n\n")
		if signal.TargetPrice > 0 {
			markdown += fmt.Sprintf("🎯 **目标价格**: %.2f元\n\n", signal.TargetPrice)
		}
		if signal.StopLoss > 0 {
			markdown += fmt.Sprintf("🛑 **止损价格**: %.2f元\n\n", signal.StopLoss)
		}
		if signal.RiskReward != "" {
			markdown += fmt.Sprintf("⚖️ **风险回报比**: %s\n\n", signal.RiskReward)
		}
	}

	// 如果有持仓信息，添加到交易建议中
	if signal.PositionInfo != nil {
		if quantity, ok := signal.PositionInfo["quantity"].(int); ok && quantity > 0 {
			markdown += fmt.Sprintf("📦 **持仓数量**: %d股\n\n", quantity)
		}
		if buyPrice, ok := signal.PositionInfo["buy_price"].(float64); ok && buyPrice > 0 {
			markdown += fmt.Sprintf("💵 **购买价格**: %.2f元/股\n\n", buyPrice)
		}
		if currentPrice, ok := signal.PositionInfo["current_price"].(float64); ok && currentPrice > 0 {
			markdown += fmt.Sprintf("💰 **持仓当前价格**: %.2f元/股\n\n", currentPrice)
		}
		if profitLoss, ok := signal.PositionInfo["profit_loss"].(float64); ok {
			profitLossPercent := signal.PositionInfo["profit_loss_percent"].(float64)
			profitEmoji := "📈"
			sign := "+"
			if profitLoss < 0 {
				profitEmoji = "📉"
				sign = ""
			}
			markdown += fmt.Sprintf("%s **浮动盈亏**: %s%.2f元 (%.2f%%)\n\n", profitEmoji, sign, profitLoss, profitLossPercent)
		}
		
		// 添加持仓止盈止损价格
		if signal.PositionProfitTarget > 0 || signal.PositionStopLoss > 0 {
			if signal.PositionProfitTarget > 0 {
				markdown += fmt.Sprintf("📈 **持仓止盈价**: %.2f元\n\n", signal.PositionProfitTarget)
			}
			if signal.PositionStopLoss > 0 {
				markdown += fmt.Sprintf("📉 **持仓止损价**: %.2f元\n\n", signal.PositionStopLoss)
			}
		}
	}

	// 3️⃣ 分析原因
	markdown += fmt.Sprintf("**3️⃣  分析原因**\n\n")
	markdown += fmt.Sprintf("%s\n\n", formatReasoning(signal.Reasoning))

	// 4️⃣ 分析时间和风险提示
	markdown += fmt.Sprintf("**4️⃣  分析时间**  %s\n\n", signal.Timestamp.Format("2006-01-02 15:04:05"))
	markdown += fmt.Sprintf("‼️ **本分析仅供参考，投资有风险，决策需谨慎**")

	return markdown
}

// formatReasoning 格式化分析原因，将编号列表项分行显示
func formatReasoning(reasoning string) string {
	if reasoning == "" {
		return reasoning
	}

	// 匹配模式：数字后跟中文右括号或英文右括号，例如：1）、2）、1)、2)
	re := regexp.MustCompile(`(\d+)[）)]`)
	
	// 替换为换行格式：在编号前添加换行，在冒号/分号后添加换行（如果存在）
	result := re.ReplaceAllStringFunc(reasoning, func(match string) string {
		// 在匹配项前添加换行（如果不是文本开头）
		return "\n\n" + match
	})
	
	// 清理多余的空白行
	result = regexp.MustCompile(`\n{3,}`).ReplaceAllString(result, "\n\n")
	
	// 如果开头有多余的换行，移除
	result = strings.TrimLeft(result, "\n")
	
	return result
}

// sendRequest 发送HTTP请求到钉钉
func (d *DingTalkNotifier) sendRequest(message map[string]interface{}) error {
	jsonData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("序列化消息失败: %w", err)
	}

	// TODO: 如果有Secret，需要进行加签处理
	// 钉钉加签文档: https://open.dingtalk.com/document/robots/custom-robot-access

	resp, err := http.Post(d.WebhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("解析响应失败: %w", err)
	}

	if errcode, ok := result["errcode"].(float64); ok && errcode != 0 {
		return fmt.Errorf("钉钉API错误: %v", result["errmsg"])
	}

	return nil
}

// FeishuNotifier 飞书通知器
type FeishuNotifier struct {
	WebhookURL string
	Secret     string // 签名密钥（可选）
}

// NewFeishuNotifier 创建飞书通知器
func NewFeishuNotifier(webhookURL string, secret string) *FeishuNotifier {
	return &FeishuNotifier{
		WebhookURL: webhookURL,
		Secret:     secret,
	}
}

// SendSignal 发送交易信号到飞书
func (f *FeishuNotifier) SendSignal(signal *TradingSignal) error {
	// 构建富文本消息
	content := f.formatSignalRichText(signal)

	// 飞书消息格式
	message := map[string]interface{}{
		"msg_type": "interactive",
		"card":     content,
	}

	return f.sendRequest(message)
}

// SendMessage 发送普通消息到飞书
func (f *FeishuNotifier) SendMessage(message string) error {
	msg := map[string]interface{}{
		"msg_type": "text",
		"content": map[string]string{
			"text": message,
		},
	}
	return f.sendRequest(msg)
}

// formatSignalRichText 格式化信号为飞书卡片
func (f *FeishuNotifier) formatSignalRichText(signal *TradingSignal) map[string]interface{} {
	var emoji string
	var color string
	switch signal.Signal {
	case "BUY":
		emoji = "🚀"
		color = "red"
	case "SELL":
		emoji = "⚠️"
		color = "green"
	case "HOLD":
		emoji = "⏸️"
		color = "yellow"
	default:
		emoji = "📊"
		color = "grey"
	}

	// 飞书卡片消息
	card := map[string]interface{}{
		"config": map[string]bool{
			"wide_screen_mode": true,
		},
		"header": map[string]interface{}{
			"title": map[string]interface{}{
				"tag":     "plain_text",
				"content": fmt.Sprintf("%s %s信号 - %s(%s)", emoji, getSignalText(signal.Signal), signal.StockName, signal.StockCode),
			},
			"template": color,
		},
		"elements": []map[string]interface{}{
			// 系统标识
			{
				"tag": "note",
				"elements": []map[string]string{
					{
						"tag":     "plain_text",
						"content": "【AI股票分析系统】",
					},
				},
			},
			// 分割线
			{
				"tag": "hr",
			},
			// 1️⃣ 核心指标
			{
				"tag": "div",
				"text": map[string]string{
					"tag":     "lark_md",
					"content": "**1️⃣  核心指标**",
				},
			},
			{
				"tag": "div",
				"fields": []map[string]interface{}{
					{
						"is_short": true,
						"text": map[string]string{
							"tag":     "lark_md",
							"content": fmt.Sprintf("💰 **当前价格**\n%.2f元", signal.Price),
						},
					},
					{
						"is_short": true,
						"text": map[string]string{
							"tag":     "lark_md",
							"content": fmt.Sprintf("📈 **信心度**\n%d%%", signal.Confidence),
						},
					},
				},
			},
		},
	}

	// 2️⃣ 添加目标价格和止损
	if signal.TargetPrice > 0 || signal.StopLoss > 0 || signal.RiskReward != "" {
		// 添加标题
		card["elements"] = append(card["elements"].([]map[string]interface{}), map[string]interface{}{
			"tag": "div",
			"text": map[string]string{
				"tag":     "lark_md",
				"content": "**2️⃣  交易建议**",
			},
		})
		
		fields := []map[string]interface{}{}
		if signal.TargetPrice > 0 {
			fields = append(fields, map[string]interface{}{
				"is_short": true,
				"text": map[string]string{
					"tag":     "lark_md",
					"content": fmt.Sprintf("**目标价格**\n%.2f元", signal.TargetPrice),
				},
			})
		}
		if signal.StopLoss > 0 {
			fields = append(fields, map[string]interface{}{
				"is_short": true,
				"text": map[string]string{
					"tag":     "lark_md",
					"content": fmt.Sprintf("**止损价格**\n%.2f元", signal.StopLoss),
				},
			})
		}
		if signal.RiskReward != "" {
			fields = append(fields, map[string]interface{}{
				"is_short": true,
				"text": map[string]string{
					"tag":     "lark_md",
					"content": fmt.Sprintf("**风险回报比**\n%s", signal.RiskReward),
				},
			})
		}
		if len(fields) > 0 {
			card["elements"] = append(card["elements"].([]map[string]interface{}), map[string]interface{}{
				"tag":    "div",
				"fields": fields,
			})
		}
	}

	// 如果有持仓信息，添加到交易建议中
	if signal.PositionInfo != nil {
		
		positionFields := []map[string]interface{}{}
		
		if quantity, ok := signal.PositionInfo["quantity"].(int); ok && quantity > 0 {
			positionFields = append(positionFields, map[string]interface{}{
				"is_short": true,
				"text": map[string]string{
					"tag":     "lark_md",
					"content": fmt.Sprintf("**持仓数量**\n%d股", quantity),
				},
			})
		}
		if buyPrice, ok := signal.PositionInfo["buy_price"].(float64); ok && buyPrice > 0 {
			positionFields = append(positionFields, map[string]interface{}{
				"is_short": true,
				"text": map[string]string{
					"tag":     "lark_md",
					"content": fmt.Sprintf("**购买价格**\n%.2f元/股", buyPrice),
				},
			})
		}
		if currentPrice, ok := signal.PositionInfo["current_price"].(float64); ok && currentPrice > 0 {
			positionFields = append(positionFields, map[string]interface{}{
				"is_short": true,
				"text": map[string]string{
					"tag":     "lark_md",
					"content": fmt.Sprintf("**当前价格**\n%.2f元/股", currentPrice),
				},
			})
		}
		if profitLoss, ok := signal.PositionInfo["profit_loss"].(float64); ok {
			profitLossPercent := 0.0
			if percent, ok := signal.PositionInfo["profit_loss_percent"].(float64); ok {
				profitLossPercent = percent
			}
			profitEmoji := "📈"
			if profitLoss < 0 {
				profitEmoji = "📉"
			}
			positionFields = append(positionFields, map[string]interface{}{
				"is_short": true,
				"text": map[string]string{
					"tag":     "lark_md",
					"content": fmt.Sprintf("**浮动盈亏**\n%s%.2f元\n%.2f%%", profitEmoji, profitLoss, profitLossPercent),
				},
			})
		}
		
		if len(positionFields) > 0 {
			card["elements"] = append(card["elements"].([]map[string]interface{}), map[string]interface{}{
				"tag":    "div",
				"fields": positionFields,
			})
		}
		
		// 添加持仓止盈止损价格
		if signal.PositionProfitTarget > 0 || signal.PositionStopLoss > 0 {
			
			profitStopFields := []map[string]interface{}{}
			if signal.PositionProfitTarget > 0 {
				profitStopFields = append(profitStopFields, map[string]interface{}{
					"is_short": true,
					"text": map[string]string{
						"tag":     "lark_md",
						"content": fmt.Sprintf("**持仓止盈价**\n%.2f元", signal.PositionProfitTarget),
					},
				})
			}
			if signal.PositionStopLoss > 0 {
				profitStopFields = append(profitStopFields, map[string]interface{}{
					"is_short": true,
					"text": map[string]string{
						"tag":     "lark_md",
						"content": fmt.Sprintf("**持仓止损价**\n%.2f元", signal.PositionStopLoss),
					},
				})
			}
			if len(profitStopFields) > 0 {
				card["elements"] = append(card["elements"].([]map[string]interface{}), map[string]interface{}{
					"tag":    "div",
					"fields": profitStopFields,
				})
			}
		}
	}

	// 添加分割线
	card["elements"] = append(card["elements"].([]map[string]interface{}), map[string]interface{}{
		"tag": "hr",
	})

	// 3️⃣ 添加分析原因
	card["elements"] = append(card["elements"].([]map[string]interface{}), map[string]interface{}{
		"tag": "hr",
	})
	card["elements"] = append(card["elements"].([]map[string]interface{}), map[string]interface{}{
		"tag": "div",
		"text": map[string]string{
			"tag":     "lark_md",
			"content": "**3️⃣  分析原因**",
		},
	})
	card["elements"] = append(card["elements"].([]map[string]interface{}), map[string]interface{}{
		"tag": "div",
		"text": map[string]string{
			"tag":     "lark_md",
			"content": formatReasoning(signal.Reasoning),
		},
	})

	// 4️⃣ 添加时间戳和风险提示
	card["elements"] = append(card["elements"].([]map[string]interface{}), map[string]interface{}{
		"tag": "hr",
	})
	card["elements"] = append(card["elements"].([]map[string]interface{}), map[string]interface{}{
		"tag": "div",
		"text": map[string]string{
			"tag":     "lark_md",
			"content": fmt.Sprintf("**4️⃣  分析时间**  %s", signal.Timestamp.Format("2006-01-02 15:04:05")),
		},
	})
	card["elements"] = append(card["elements"].([]map[string]interface{}), map[string]interface{}{
		"tag": "note",
		"elements": []map[string]string{
			{
				"tag":     "plain_text",
				"content": "‼️ 本分析仅供参考，投资有风险，决策需谨慎",
			},
		},
	})

	return card
}

// sendRequest 发送HTTP请求到飞书
func (f *FeishuNotifier) sendRequest(message map[string]interface{}) error {
	jsonData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("序列化消息失败: %w", err)
	}

	// TODO: 如果有Secret，需要进行签名处理
	// 飞书签名文档: https://open.feishu.cn/document/ukTMukTMukTM/ucTM5YjL3ETO24yNxkjN

	resp, err := http.Post(f.WebhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("解析响应失败: %w", err)
	}

	if code, ok := result["code"].(float64); ok && code != 0 {
		return fmt.Errorf("飞书API错误: %v", result["msg"])
	}

	return nil
}

// MultiNotifier 多通知器（同时发送到多个平台）
type MultiNotifier struct {
	Notifiers []Notifier
}

// NewMultiNotifier 创建多通知器
func NewMultiNotifier(notifiers ...Notifier) *MultiNotifier {
	return &MultiNotifier{
		Notifiers: notifiers,
	}
}

// SendSignal 发送信号到所有通知器
func (m *MultiNotifier) SendSignal(signal *TradingSignal) error {
	var errors []error
	for _, notifier := range m.Notifiers {
		if err := notifier.SendSignal(signal); err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("部分通知器发送失败: %v", errors)
	}
	return nil
}

// SendMessage 发送消息到所有通知器
func (m *MultiNotifier) SendMessage(message string) error {
	var errors []error
	for _, notifier := range m.Notifiers {
		if err := notifier.SendMessage(message); err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("部分通知器发送失败: %v", errors)
	}
	return nil
}
