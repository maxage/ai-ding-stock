package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"
)

// StockConfig 股票分析系统配置
type StockConfig struct {
	TDXAPIUrl     string             `json:"tdx_api_url"`
	AIConfig      AIConfig           `json:"ai_config"`
	Stocks        []StockItem        `json:"stocks"`
	Notification  NotificationConfig `json:"notification"`
	TradingTime   TradingTimeConfig  `json:"trading_time"`
	APIServerPort      int    `json:"api_server_port"`
	LogDir             string `json:"log_dir"`
	APIToken           string `json:"api_token,omitempty"`           // API认证Token，用于前端重启后端等功能。默认：1122334455667788（为了安全，强烈建议修改！）
	AnalysisHistoryLimit int  `json:"analysis_history_limit"`       // 分析历史记录数量（最小3条，最大100条，默认20条）
	AnalysisMode        string `json:"analysis_mode,omitempty"`      // 分析模式："smart"（智能模式，推荐）、"concurrent"（并发模式）、"polling"（轮询模式），默认："smart"
	MaxConcurrentAnalysis int  `json:"max_concurrent_analysis,omitempty"` // 最大并发分析数（1-4，默认3），仅并发模式和智能模式有效
}

// TradingTimeConfig 交易时间配置
type TradingTimeConfig struct {
	EnableCheck  bool     `json:"enable_check"`  // 是否启用交易时间检查
	TradingHours []string `json:"trading_hours"` // 交易时段（如：["09:30-11:30", "13:00-15:00"]）
	Timezone     string   `json:"timezone"`      // 时区（如：Asia/Shanghai）
}

// AIConfig AI配置
type AIConfig struct {
	Provider        string `json:"provider"` // "deepseek", "qwen", "custom"
	DeepSeekKey     string `json:"deepseek_key"`
	QwenKey         string `json:"qwen_key"`
	CustomAPIURL    string `json:"custom_api_url"`
	CustomAPIKey    string `json:"custom_api_key"`
	CustomModelName string `json:"custom_model_name"`
}

// StockItem 股票配置项
type StockItem struct {
	Code                string  `json:"code"`
	Name                string  `json:"name"`
	Enabled             bool    `json:"enabled"`
	ScanIntervalMinutes int     `json:"scan_interval_minutes"`
	MinConfidence       int     `json:"min_confidence"` // 最小信心度阈值
	
	// 新增：持仓模式相关字段（可选）
	PositionQuantity    int     `json:"position_quantity,omitempty"` // 持仓数量（股）
	BuyPrice            float64 `json:"buy_price,omitempty"` // 购买价格（元/股）
	BuyDate             string  `json:"buy_date,omitempty"` // 购买日期（YYYY-MM-DD，可选）
}

// NotificationConfig 通知配置
type NotificationConfig struct {
	Enabled  bool           `json:"enabled"`
	DingTalk DingTalkConfig `json:"dingtalk"`
	Feishu   FeishuConfig   `json:"feishu"`
}

// DingTalkConfig 钉钉配置
type DingTalkConfig struct {
	Enabled    bool   `json:"enabled"`
	WebhookURL string `json:"webhook_url"`
	Secret     string `json:"secret"`
}

// FeishuConfig 飞书配置
type FeishuConfig struct {
	Enabled    bool   `json:"enabled"`
	WebhookURL string `json:"webhook_url"`
	Secret     string `json:"secret"`
}

// LoadStockConfig 加载股票分析配置
func LoadStockConfig(filename string) (*StockConfig, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	var config StockConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	// 验证配置
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("配置验证失败: %w", err)
	}

	return &config, nil
}

// Validate 验证配置
func (c *StockConfig) Validate() error {
	// 验证TDX API URL
	if c.TDXAPIUrl == "" {
		return fmt.Errorf("tdx_api_url不能为空")
	}

	// 验证AI配置
	if c.AIConfig.Provider == "" {
		return fmt.Errorf("ai_config.provider不能为空")
	}
	if c.AIConfig.Provider != "deepseek" && c.AIConfig.Provider != "qwen" && c.AIConfig.Provider != "custom" {
		return fmt.Errorf("ai_config.provider必须是 'deepseek', 'qwen' 或 'custom'")
	}

	// 验证对应的API密钥
	if c.AIConfig.Provider == "deepseek" && c.AIConfig.DeepSeekKey == "" {
		return fmt.Errorf("使用DeepSeek时必须配置deepseek_key")
	}
	if c.AIConfig.Provider == "qwen" && c.AIConfig.QwenKey == "" {
		return fmt.Errorf("使用Qwen时必须配置qwen_key")
	}
	if c.AIConfig.Provider == "custom" {
		if c.AIConfig.CustomAPIURL == "" || c.AIConfig.CustomAPIKey == "" || c.AIConfig.CustomModelName == "" {
			return fmt.Errorf("使用自定义API时必须配置custom_api_url, custom_api_key和custom_model_name")
		}
	}

	// 验证股票列表
	if len(c.Stocks) == 0 {
		return fmt.Errorf("至少需要配置一只股票")
	}

	stockCodes := make(map[string]bool)
	enabledCount := 0
	for i, stock := range c.Stocks {
		// 设置默认值
		c.Stocks[i].SetDefaults()

		if stock.Code == "" {
			return fmt.Errorf("stocks[%d]: code不能为空", i)
		}
		if stock.Name == "" {
			return fmt.Errorf("stocks[%d]: name不能为空", i)
		}
		if stockCodes[stock.Code] {
			return fmt.Errorf("stocks[%d]: 股票代码 '%s' 重复", i, stock.Code)
		}
		stockCodes[stock.Code] = true

		if stock.Enabled {
			enabledCount++
		}

		// 验证持仓模式配置
		// 如果填写了持仓数量或购买价格，必须两者都填写
		if (stock.PositionQuantity > 0 && stock.BuyPrice <= 0) ||
			(stock.PositionQuantity <= 0 && stock.BuyPrice > 0) {
			return fmt.Errorf("stocks[%d]: 持仓数量和购买价格必须同时填写", i)
		}

		if stock.PositionQuantity < 0 {
			return fmt.Errorf("stocks[%d]: 持仓数量不能为负数", i)
		}

		if stock.BuyPrice < 0 {
			return fmt.Errorf("stocks[%d]: 购买价格不能为负数", i)
		}
	}

	if enabledCount == 0 {
		return fmt.Errorf("至少需要启用一只股票")
	}

	// 设置默认API端口
	if c.APIServerPort <= 0 {
		c.APIServerPort = 9090
	}

	// 设置默认日志目录
	if c.LogDir == "" {
		c.LogDir = "stock_analysis_logs"
	}

	// 设置默认分析历史记录数量
	if c.AnalysisHistoryLimit <= 0 {
		c.AnalysisHistoryLimit = 20 // 默认保存20条记录
	} else if c.AnalysisHistoryLimit < 3 {
		c.AnalysisHistoryLimit = 3 // 最小3条
	} else if c.AnalysisHistoryLimit > 100 {
		c.AnalysisHistoryLimit = 100 // 最大100条
	}

	// 设置默认分析模式
	if c.AnalysisMode == "" {
		c.AnalysisMode = "smart" // 默认智能模式
	} else if c.AnalysisMode != "smart" && c.AnalysisMode != "concurrent" && c.AnalysisMode != "polling" {
		log.Printf("⚠️  无效的分析模式 '%s'，将使用默认模式 'smart'", c.AnalysisMode)
		c.AnalysisMode = "smart"
	}

	// 设置默认最大并发分析数
	if c.MaxConcurrentAnalysis <= 0 {
		c.MaxConcurrentAnalysis = 3 // 默认3个并发
	} else if c.MaxConcurrentAnalysis < 1 {
		c.MaxConcurrentAnalysis = 1 // 最小1个
	} else if c.MaxConcurrentAnalysis > 4 {
		c.MaxConcurrentAnalysis = 4 // 最大4个（避免触发AI模型的RPM/TPM限制）
	}

	// 设置默认交易时间配置
	if c.TradingTime.Timezone == "" {
		c.TradingTime.Timezone = "Asia/Shanghai"
	}
	if len(c.TradingTime.TradingHours) == 0 {
		c.TradingTime.TradingHours = []string{"09:30-11:30", "13:00-15:00"} // A股默认交易时段
	}

	// 从环境变量读取API Token（如果配置文件中没有）
	if c.APIToken == "" {
		if envToken := os.Getenv("API_TOKEN"); envToken != "" {
			c.APIToken = envToken
		}
	}

	// 如果还是没有Token，使用默认Token（仅在首次运行时）
	// 注意：为了安全，强烈建议在生产环境中修改此默认Token！
	// 默认Token仅用于测试，生产环境应该：
	// 1. 在配置文件中设置强密码Token
	// 2. 或通过环境变量 API_TOKEN 设置
	// 3. 或通过Web界面重新设置
	if c.APIToken == "" {
		// 默认Token：1122334455667788（为了安全，强烈建议修改！）
		c.APIToken = "1122334455667788"
		log.Printf("⚠️  使用默认API Token，为了安全，请在生产环境中修改！")
	}

	// 验证通知配置
	if c.Notification.Enabled {
		if !c.Notification.DingTalk.Enabled && !c.Notification.Feishu.Enabled {
			return fmt.Errorf("启用通知时至少需要配置一个通知渠道（钉钉或飞书）")
		}
		if c.Notification.DingTalk.Enabled && c.Notification.DingTalk.WebhookURL == "" {
			return fmt.Errorf("启用钉钉通知时必须配置webhook_url")
		}
		if c.Notification.Feishu.Enabled && c.Notification.Feishu.WebhookURL == "" {
			return fmt.Errorf("启用飞书通知时必须配置webhook_url")
		}
	}

	return nil
}

// GetScanInterval 获取扫描间隔
func (s *StockItem) GetScanInterval() time.Duration {
	return time.Duration(s.ScanIntervalMinutes) * time.Minute
}

// IsPositionMode 判断是否为持仓模式
// 有持仓数量且购买价格>0时，判定为持仓模式
func (s *StockItem) IsPositionMode() bool {
	return s.PositionQuantity > 0 && s.BuyPrice > 0
}

// SetDefaults 设置默认值
func (s *StockItem) SetDefaults() {
	if s.ScanIntervalMinutes <= 0 {
		s.ScanIntervalMinutes = 5
	}
	if s.MinConfidence <= 0 {
		s.MinConfidence = 70
	}
}
