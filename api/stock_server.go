package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"nofx/stock"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// StockAPIServer 股票分析API服务器
type StockAPIServer struct {
	router      *gin.Engine
	manager     AnalyzerManagerInterface
	port        int
	apiToken    string // API认证Token
	restartFunc func() // 重启函数（由main函数提供）
}

// AnalyzerManagerInterface 分析器管理器接口
type AnalyzerManagerInterface interface {
	GetAnalyzer(code string) interface{}
	GetAllAnalyzers() map[string]interface{}
	TriggerAnalysis(code string) (interface{}, error) // 手动触发分析
	GetAnalysisHistory(code string, limit int) interface{} // 获取分析历史
	GetAllRecentAnalysis(limit int) interface{} // 获取所有股票的最近分析记录
}

// NewStockAPIServer 创建股票API服务器
func NewStockAPIServer(manager AnalyzerManagerInterface, port int, apiToken string) *StockAPIServer {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	// 配置CORS
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-API-Token"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	server := &StockAPIServer{
		router:   router,
		manager:  manager,
		port:     port,
		apiToken: apiToken,
	}

	server.setupRoutes()
	return server
}

// SetRestartFunc 设置重启函数（由main函数提供）
func (s *StockAPIServer) SetRestartFunc(fn func()) {
	s.restartFunc = fn
}

// setupRoutes 设置路由
func (s *StockAPIServer) setupRoutes() {
	// 健康检查（兼容两种路径）
	s.router.GET("/health", s.handleHealth)
	s.router.GET("/api/health", s.handleHealth)

	// Favicon处理（避免404）
	s.router.GET("/favicon.ico", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	// 静态文件服务
	s.router.Static("/static", "./web/static")
	s.router.StaticFile("/", "./web/config.html")
	s.router.StaticFile("/config", "./web/config.html")

	// API路由组
	api := s.router.Group("/api")
	{
		// 配置管理接口
		api.GET("/config", s.handleGetConfig)
		api.POST("/config", s.handleSaveConfig)

		// 获取所有监控股票列表
		api.GET("/stocks", s.handleGetStocks)

		// 获取单个股票的最新分析结果
		api.GET("/stock/:code/latest", s.handleGetLatestAnalysis)

		// 获取单个股票的历史分析记录
		api.GET("/stock/:code/history", s.handleGetAnalysisHistory)

		// 获取所有股票的最近分析记录
		api.GET("/analysis/recent", s.handleGetRecentAnalysis)

		// 手动触发分析
		api.POST("/stock/:code/analyze", s.handleTriggerAnalysis)

		// 获取系统统计信息
		api.GET("/statistics", s.handleGetStatistics)
		
		// 系统测试接口
		api.POST("/test", s.handleSystemTest)
		api.POST("/test/tdx", s.handleTestTDX)
		api.POST("/test/ai", s.handleTestAI)
		api.POST("/test/stock/:code", s.handleTestStock)

		// 系统控制接口（需要Token认证）
		api.POST("/system/restart", s.handleRestart)
	}
}

// handleHealth 健康检查
func (s *StockAPIServer) handleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"time":   time.Now().Format("2006-01-02 15:04:05"),
	})
}

// handleGetStocks 获取所有监控股票
func (s *StockAPIServer) handleGetStocks(c *gin.Context) {
	analyzers := s.manager.GetAllAnalyzers()

	stocks := []gin.H{}
	for code := range analyzers {
		// TODO: 获取每个分析器的配置信息
		stocks = append(stocks, gin.H{
			"code":    code,
			"name":    "", // 需要从analyzer获取
			"enabled": true,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"total":  len(stocks),
			"stocks": stocks,
		},
	})
}

// handleGetLatestAnalysis 获取最新分析结果
func (s *StockAPIServer) handleGetLatestAnalysis(c *gin.Context) {
	code := c.Param("code")

	analyzer := s.manager.GetAnalyzer(code)
	if analyzer == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    -1,
			"message": "未找到该股票的分析器",
		})
		return
	}

	// 获取该股票的最新分析结果
	historyInterface := s.manager.GetAnalysisHistory(code, 1)
	history, ok := historyInterface.([]*stock.AnalysisResult)
	if !ok || len(history) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": "暂无分析结果",
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    history[0],
	})
}

// handleGetAnalysisHistory 获取历史分析记录
func (s *StockAPIServer) handleGetAnalysisHistory(c *gin.Context) {
	code := c.Param("code")
	limit := 20 // 默认返回最近20条

	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := fmt.Sscanf(limitStr, "%d", &limit); err == nil && parsedLimit == 1 && limit > 0 && limit <= 100 {
			// 成功解析且在合理范围内
		} else {
			limit = 20 // 解析失败或超出范围，使用默认值
		}
	}

	analyzer := s.manager.GetAnalyzer(code)
	if analyzer == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    -1,
			"message": "未找到该股票的分析器",
		})
		return
	}

	historyInterface := s.manager.GetAnalysisHistory(code, limit)
	history, ok := historyInterface.([]*stock.AnalysisResult)
	if !ok {
		history = []*stock.AnalysisResult{}
	}
	
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"stock_code": code,
			"count":      len(history),
			"limit":      limit,
			"records":    history,
		},
	})
}

// handleGetRecentAnalysis 获取所有股票的最近分析记录
func (s *StockAPIServer) handleGetRecentAnalysis(c *gin.Context) {
	limit := 10 // 默认返回最近10条

	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := fmt.Sscanf(limitStr, "%d", &limit); err == nil && parsedLimit == 1 && limit > 0 && limit <= 50 {
			// 成功解析且在合理范围内
		} else {
			limit = 10 // 解析失败或超出范围，使用默认值
		}
	}

	recentAnalysisInterface := s.manager.GetAllRecentAnalysis(limit)
	recentAnalysis, ok := recentAnalysisInterface.([]*stock.AnalysisResult)
	if !ok {
		recentAnalysis = []*stock.AnalysisResult{}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"count":   len(recentAnalysis),
			"limit":   limit,
			"records": recentAnalysis,
		},
	})
}

// handleTriggerAnalysis 手动触发分析
func (s *StockAPIServer) handleTriggerAnalysis(c *gin.Context) {
	code := c.Param("code")

	result, err := s.manager.TriggerAnalysis(code)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -1,
			"message": fmt.Sprintf("触发分析失败: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "分析完成",
		"data":    result,
	})
}

// handleGetStatistics 获取系统统计
func (s *StockAPIServer) handleGetStatistics(c *gin.Context) {
	analyzers := s.manager.GetAllAnalyzers()

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"total_stocks":   len(analyzers),
			"system_uptime":  "", // TODO: 计算运行时间
			"total_analysis": 0,  // TODO: 统计总分析次数
		},
	})
}

// handleGetConfig 获取配置
func (s *StockAPIServer) handleGetConfig(c *gin.Context) {
	// 读取配置文件
	configFile := "config_stock.json"
	data, err := os.ReadFile(configFile)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    -1,
			"message": fmt.Sprintf("读取配置文件失败: %v", err),
		})
		return
	}

	// 解析为JSON对象
	var config map[string]interface{}
	if err := json.Unmarshal(data, &config); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    -1,
			"message": fmt.Sprintf("解析配置文件失败: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    config,
	})
}

// handleSaveConfig 保存配置
func (s *StockAPIServer) handleSaveConfig(c *gin.Context) {
	var config map[string]interface{}
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -1,
			"message": fmt.Sprintf("请求数据格式错误: %v", err),
		})
		return
	}

	// 转换为格式化的JSON
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    -1,
			"message": fmt.Sprintf("序列化配置失败: %v", err),
		})
		return
	}

	// 备份原配置文件
	configFile := "config_stock.json"
	backupFile := fmt.Sprintf("config_stock.json.backup.%s", time.Now().Format("20060102150405"))
	if err := os.Rename(configFile, backupFile); err != nil {
		log.Printf("⚠️  备份配置文件失败: %v", err)
	} else {
		log.Printf("✓ 配置文件已备份: %s", backupFile)
	}

	// 写入新配置
	if err := os.WriteFile(configFile, data, 0644); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    -1,
			"message": fmt.Sprintf("保存配置文件失败: %v", err),
		})
		return
	}

	log.Printf("✓ 配置文件已更新: %s", configFile)

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "配置保存成功，请重启程序使配置生效",
		"data": gin.H{
			"backup_file": backupFile,
		},
	})
}

// Start 启动服务器
func (s *StockAPIServer) Start() error {
	addr := fmt.Sprintf(":%d", s.port)
	log.Printf("🚀 股票分析API服务器启动在端口 %d", s.port)
	return s.router.Run(addr)
}

// handleSystemTest 系统测试（完整测试）
func (s *StockAPIServer) handleSystemTest(c *gin.Context) {
	var testResult = gin.H{
		"timestamp": time.Now().Format("2006-01-02 15:04:05"),
		"tests":     []gin.H{},
		"passed":    0,
		"failed":    0,
		"total":     0,
	}

	tests := testResult["tests"].([]gin.H)

	// 1. 测试配置文件
	testResult["total"] = testResult["total"].(int) + 1
	configFile := "config_stock.json"
	if _, err := os.ReadFile(configFile); err == nil {
		tests = append(tests, gin.H{
			"name":    "配置文件检查",
			"status":  "passed",
			"message": "配置文件存在且可读",
		})
		testResult["passed"] = testResult["passed"].(int) + 1
	} else {
		tests = append(tests, gin.H{
			"name":    "配置文件检查",
			"status":  "failed",
			"message": fmt.Sprintf("配置文件不存在或无法读取: %v", err),
		})
		testResult["failed"] = testResult["failed"].(int) + 1
	}

	// 2. 测试TDX API连接
	testResult["total"] = testResult["total"].(int) + 1
	tdxResult := s.testTDXConnection()
	tests = append(tests, tdxResult)
	if tdxResult["status"] == "passed" {
		testResult["passed"] = testResult["passed"].(int) + 1
	} else {
		testResult["failed"] = testResult["failed"].(int) + 1
	}

	// 3. 测试AI配置
	testResult["total"] = testResult["total"].(int) + 1
	aiResult := s.testAIConfig()
	tests = append(tests, aiResult)
	if aiResult["status"] == "passed" {
		testResult["passed"] = testResult["passed"].(int) + 1
	} else {
		testResult["failed"] = testResult["failed"].(int) + 1
	}

	// 4. 测试分析器状态
	testResult["total"] = testResult["total"].(int) + 1
	analyzers := s.manager.GetAllAnalyzers()
	if len(analyzers) > 0 {
		tests = append(tests, gin.H{
			"name":    "分析器状态",
			"status":  "passed",
			"message": fmt.Sprintf("共有 %d 个分析器正在运行", len(analyzers)),
			"data":    gin.H{"count": len(analyzers)},
		})
		testResult["passed"] = testResult["passed"].(int) + 1
	} else {
		tests = append(tests, gin.H{
			"name":    "分析器状态",
			"status":  "failed",
			"message": "没有正在运行的分析器",
		})
		testResult["failed"] = testResult["failed"].(int) + 1
	}

	testResult["tests"] = tests
	testResult["success"] = testResult["failed"].(int) == 0

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "系统测试完成",
		"data":    testResult,
	})
}

// testTDXConnection 测试TDX连接
func (s *StockAPIServer) testTDXConnection() gin.H {
	configFile := "config_stock.json"
	data, err := os.ReadFile(configFile)
	if err != nil {
		return gin.H{
			"name":    "TDX API连接",
			"status":  "failed",
			"message": fmt.Sprintf("无法读取配置文件: %v", err),
		}
	}

	var config map[string]interface{}
	if err := json.Unmarshal(data, &config); err != nil {
		return gin.H{
			"name":    "TDX API连接",
			"status":  "failed",
			"message": fmt.Sprintf("配置文件格式错误: %v", err),
		}
	}

	tdxURL, ok := config["tdx_api_url"].(string)
	if !ok || tdxURL == "" {
		return gin.H{
			"name":    "TDX API连接",
			"status":  "failed",
			"message": "TDX API地址未配置",
		}
	}

	// 尝试连接TDX API
	resp, err := http.Get(fmt.Sprintf("%s/api/quote?code=000001", tdxURL))
	if err != nil {
		return gin.H{
			"name":    "TDX API连接",
			"status":  "failed",
			"message": fmt.Sprintf("无法连接到TDX API (%s): %v", tdxURL, err),
			"data":    gin.H{"url": tdxURL},
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return gin.H{
			"name":    "TDX API连接",
			"status":  "passed",
			"message": "TDX API连接正常",
			"data":    gin.H{"url": tdxURL, "status_code": resp.StatusCode},
		}
	}

	return gin.H{
		"name":    "TDX API连接",
		"status":  "failed",
		"message": fmt.Sprintf("TDX API返回错误状态码: %d", resp.StatusCode),
		"data":    gin.H{"url": tdxURL, "status_code": resp.StatusCode},
	}
}

// testAIConfig 测试AI配置
func (s *StockAPIServer) testAIConfig() gin.H {
	configFile := "config_stock.json"
	data, err := os.ReadFile(configFile)
	if err != nil {
		return gin.H{
			"name":    "AI配置检查",
			"status":  "failed",
			"message": fmt.Sprintf("无法读取配置文件: %v", err),
		}
	}

	var config map[string]interface{}
	if err := json.Unmarshal(data, &config); err != nil {
		return gin.H{
			"name":    "AI配置检查",
			"status":  "failed",
			"message": fmt.Sprintf("配置文件格式错误: %v", err),
		}
	}

	aiConfig, ok := config["ai_config"].(map[string]interface{})
	if !ok {
		return gin.H{
			"name":    "AI配置检查",
			"status":  "failed",
			"message": "AI配置项不存在",
		}
	}

	provider, _ := aiConfig["provider"].(string)
	provider = fmt.Sprintf("%v", provider) // 转换为字符串

	if provider == "" {
		return gin.H{
			"name":    "AI配置检查",
			"status":  "failed",
			"message": "AI提供商未配置",
		}
	}

	// 检查对应提供商的密钥
	hasKey := false
	var keyField string
	switch provider {
	case "deepseek":
		keyField = "deepseek_key"
		key, _ := aiConfig[keyField].(string)
		hasKey = key != "" && key != "sk-test-key-placeholder"
	case "qwen":
		keyField = "qwen_key"
		key, _ := aiConfig[keyField].(string)
		hasKey = key != ""
	case "custom":
		url, _ := aiConfig["custom_api_url"].(string)
		key, _ := aiConfig["custom_api_key"].(string)
		model, _ := aiConfig["custom_model_name"].(string)
		hasKey = url != "" && key != "" && model != ""
	}

	if !hasKey {
		return gin.H{
			"name":    "AI配置检查",
			"status":  "warning",
			"message": fmt.Sprintf("AI提供商已配置 (%s)，但API密钥未配置或为测试值", provider),
			"data":    gin.H{"provider": provider},
		}
	}

	return gin.H{
		"name":    "AI配置检查",
		"status":  "passed",
		"message": fmt.Sprintf("AI配置正常 (%s)", provider),
		"data":    gin.H{"provider": provider},
	}
}

// handleTestTDX 测试TDX连接
func (s *StockAPIServer) handleTestTDX(c *gin.Context) {
	result := s.testTDXConnection()
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "TDX连接测试完成",
		"data":    result,
	})
}

// handleTestAI 测试AI配置
func (s *StockAPIServer) handleTestAI(c *gin.Context) {
	result := s.testAIConfig()
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "AI配置测试完成",
		"data":    result,
	})
}

// handleTestStock 测试单个股票分析
func (s *StockAPIServer) handleTestStock(c *gin.Context) {
	code := c.Param("code")

	result, err := s.manager.TriggerAnalysis(code)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -1,
			"message": fmt.Sprintf("测试分析失败: %v", err),
			"data": gin.H{
				"stock_code": code,
				"error":      err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "股票分析测试完成",
		"data": gin.H{
			"stock_code": code,
			"result":     result,
			"success":    true,
		},
	})
}

// handleRestart 重启后端服务（需要Token认证）
func (s *StockAPIServer) handleRestart(c *gin.Context) {
	// 验证Token
	token := c.GetHeader("X-API-Token")
	if token == "" {
		// 尝试从请求体获取
		var body map[string]string
		if err := c.ShouldBindJSON(&body); err == nil {
			token = body["token"]
		}
	}

	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    -1,
			"message": "未提供API Token，请在请求头中添加 'X-API-Token' 或在请求体中提供 'token' 字段",
		})
		return
	}

	// 验证Token是否正确
	if s.apiToken != "" && token != s.apiToken {
		c.JSON(http.StatusForbidden, gin.H{
			"code":    -1,
			"message": "API Token验证失败",
		})
		return
	}

	// 如果Token为空或匹配，执行重启
	if s.restartFunc != nil {
		log.Printf("🔄 收到重启请求，准备重启服务...")
		
		// 先返回响应，再执行重启（避免客户端等待）
		c.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": "重启指令已接收，服务将在3秒后重启",
		})

		// 延迟执行重启，给响应返回时间
		go func() {
			time.Sleep(3 * time.Second)
			log.Printf("🔄 开始执行重启...")
			s.restartFunc()
		}()

		return
	}

	// 如果没有设置重启函数，返回错误
	c.JSON(http.StatusServiceUnavailable, gin.H{
		"code":    -1,
		"message": "重启功能未启用，请通过系统服务管理器重启",
	})
}
