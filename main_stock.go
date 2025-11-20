package main

import (
	"fmt"
	"log"
	"nofx/api"
	"nofx/config"
	"nofx/mcp"
	"nofx/notifier"
	"nofx/stock"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

func main() {
	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║    📈 AI股票分析系统 - 实时分析与信号通知               ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")
	fmt.Println()

	// 加载配置文件
	configFile := "config_stock.json"
	if len(os.Args) > 1 {
		configFile = os.Args[1]
	}

	log.Printf("📋 加载配置文件: %s", configFile)
	cfg, err := config.LoadStockConfig(configFile)
	if err != nil {
		log.Fatalf("❌ 加载配置失败: %v", err)
	}

	log.Printf("✓ 配置加载成功")
	fmt.Println()

	// 创建TDX客户端
	tdxClient := stock.NewTDXClient(cfg.TDXAPIUrl)
	log.Printf("✓ TDX API客户端已初始化: %s", cfg.TDXAPIUrl)

	// 创建AI客户端
	mcpClient, err := createMCPClient(&cfg.AIConfig)
	if err != nil {
		log.Fatalf("❌ 创建AI客户端失败: %v", err)
	}
	log.Printf("✓ AI客户端已初始化 (%s)", strings.ToUpper(cfg.AIConfig.Provider))

	// 创建通知器
	var notif notifier.Notifier
	if cfg.Notification.Enabled {
		notif = createNotifier(&cfg.Notification)
		log.Printf("✓ 通知系统已初始化")
	} else {
		log.Printf("⏭️  通知系统未启用")
	}

	// 创建交易时间检查器
	tradingTimeConfig := stock.TradingTimeConfig{
		EnableTradingTimeCheck: cfg.TradingTime.EnableCheck,
		TradingHours:           cfg.TradingTime.TradingHours,
		Timezone:               cfg.TradingTime.Timezone,
	}
	tradingTimeChecker, err := stock.NewTradingTimeChecker(tradingTimeConfig)
	if err != nil {
		log.Printf("⚠️  创建交易时间检查器失败: %v, 将禁用交易时间检查", err)
		tradingTimeChecker = nil
	} else if cfg.TradingTime.EnableCheck {
		log.Printf("✓ 交易时间检查已启用")
		log.Printf("  交易时段: %v", cfg.TradingTime.TradingHours)
		status := tradingTimeChecker.GetTradingTimeStatus(time.Now())
		log.Printf("  当前状态: 交易日=%v, 交易时段=%v",
			status["is_trading_day"], status["is_trading_time"])
	} else {
		log.Printf("⏭️  交易时间检查未启用（将持续分析）")
	}

	// 创建日志目录
	if err := os.MkdirAll(cfg.LogDir, 0755); err != nil {
		log.Printf("⚠️  创建日志目录失败: %v", err)
	}

	fmt.Println()
	fmt.Println("📊 监控股票列表:")
	enabledStocks := []config.StockItem{}
	for _, stockItem := range cfg.Stocks {
		if stockItem.Enabled {
			enabledStocks = append(enabledStocks, stockItem)
			fmt.Printf("  • %s(%s) - 扫描间隔: %d分钟, 信心阈值: %d%%\n",
				stockItem.Name, stockItem.Code, stockItem.ScanIntervalMinutes, stockItem.MinConfidence)
		}
	}

	fmt.Println()
	fmt.Println("🤖 AI分析模式:")
	fmt.Println("  • AI将基于实时行情、K线、技术指标进行全面分析")
	fmt.Println("  • 提供BUY/SELL/HOLD明确信号")
	fmt.Println("  • 给出目标价位和止损建议")
	fmt.Println("  • 信心度≥阈值时发送通知")
	fmt.Println()
	fmt.Println("⚠️  风险提示: AI分析仅供参考，投资有风险，决策需谨慎！")
	fmt.Println()
	fmt.Println("按 Ctrl+C 停止运行")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println()

	// 创建分析器管理器
	// 使用配置文件中的分析历史记录数量限制（最小3，最大100，默认20）
	maxHistorySize := cfg.AnalysisHistoryLimit
	if maxHistorySize < 3 {
		maxHistorySize = 3
	} else if maxHistorySize > 100 {
		maxHistorySize = 100
	}
	analyzerManager := &AnalyzerManager{
		analyzers:           make(map[string]*stock.StockAnalyzer),
		stopChans:           make(map[string]chan struct{}),
		analysisHistory:     make(map[string][]*stock.AnalysisResult),
		maxHistorySize:      maxHistorySize,      // 从配置文件读取，每个股票最多保存的分析记录数
		analysisMode:        cfg.AnalysisMode,    // 分析模式：smart/concurrent/polling
		maxConcurrent:       cfg.MaxConcurrentAnalysis, // 最大并发分析数
		stockCount:          len(enabledStocks),  // 启用的股票数量
	}
	log.Printf("✓ 分析历史记录配置: 每个股票最多保存 %d 条记录", maxHistorySize)

	// 为每只启用的股票创建分析器
	for _, stockItem := range enabledStocks {
		analysisConfig := &stock.AnalysisConfig{
			StockCode:          stockItem.Code,
			StockName:          stockItem.Name,
			ScanInterval:       stockItem.GetScanInterval(),
			EnableNotification: cfg.Notification.Enabled,
			MinConfidence:      stockItem.MinConfidence,
			
			// 新增：持仓信息（如果填写了）
			PositionQuantity: stockItem.PositionQuantity,
			BuyPrice:         stockItem.BuyPrice,
			BuyDate:          parseBuyDate(stockItem.BuyDate),
		}

		analyzer := stock.NewStockAnalyzer(tdxClient, mcpClient, notif, analysisConfig, tradingTimeChecker)
		analyzerManager.AddAnalyzer(stockItem.Code, analyzer)
	}

	// 创建并启动API服务器
	apiServer := api.NewStockAPIServer(analyzerManager, cfg.APIServerPort, cfg.APIToken)
	
	// 设置重启函数（优雅重启）
	apiServer.SetRestartFunc(func() {
		log.Printf("🔄 收到重启指令，开始优雅关闭...")
		analyzerManager.StopAll()
		log.Printf("✅ 所有分析器已停止")
		
		// 尝试通过管理脚本自动重启
		// 获取当前工作目录或可执行文件所在目录
		workDir := "."
		if exePath, err := os.Executable(); err == nil {
			if absPath, err := os.Readlink(exePath); err == nil {
				exePath = absPath
			}
			if exeDir := fmt.Sprintf("%s/../", exePath); exeDir != "" {
				workDir = exeDir
			}
		}
		
		// 尝试多个可能的脚本路径（相对路径优先）
		scriptPaths := []string{
			"./manage_backend.sh",
			fmt.Sprintf("%s/manage_backend.sh", workDir),
		}
		
		// 如果当前目录就是脚本目录，添加绝对路径
		if cwd, err := os.Getwd(); err == nil {
			scriptPaths = append(scriptPaths, fmt.Sprintf("%s/manage_backend.sh", cwd))
		}
		
		scriptFound := false
		for _, scriptPath := range scriptPaths {
			if _, err := os.Stat(scriptPath); err == nil {
				log.Printf("📜 检测到管理脚本: %s，尝试自动重启...", scriptPath)
				// 在后台执行重启脚本（分离进程，避免阻塞）
				cmd := exec.Command("bash", scriptPath, "restart")
				cmd.Dir = workDir
				cmd.Env = os.Environ()
				// 分离标准输入输出，让脚本在后台执行
				cmd.Stdin = nil
				cmd.Stdout = nil
				cmd.Stderr = nil
				
				if err := cmd.Start(); err == nil {
					log.Printf("✅ 已触发重启脚本，服务将在后台重启")
					// 不等待命令完成，让脚本独立运行
					_ = cmd.Process.Release()
					scriptFound = true
					// 等待一小段时间让脚本开始执行
					time.Sleep(2 * time.Second)
					break
				} else {
					log.Printf("⚠️  执行重启脚本失败: %v", err)
				}
			}
		}
		
		if !scriptFound {
			log.Printf("⚠️  未找到管理脚本，程序将退出")
			log.Printf("💡 提示：请手动执行 './manage_backend.sh restart' 或使用 systemd/supervisor 管理，服务将自动重启")
		}
		
		log.Printf("👋 程序退出")
		os.Exit(0) // 退出程序，由脚本或外部进程管理器重启
	})
	
	go func() {
		if err := apiServer.Start(); err != nil {
			log.Printf("❌ API服务器错误: %v", err)
		}
	}()
	log.Printf("✓ API服务器已启动: http://localhost:%d", cfg.APIServerPort)
	if cfg.APIToken != "" {
		log.Printf("✓ API Token已配置（可用于重启等功能）")
	}
	fmt.Println()

	// 设置优雅退出
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// 启动所有分析器
	analyzerManager.StartAll()

	// 等待退出信号
	<-sigChan
	fmt.Println()
	fmt.Println()
	log.Println("📛 收到退出信号，正在停止所有分析器...")
	analyzerManager.StopAll()

	fmt.Println()
	fmt.Println("👋 感谢使用AI股票分析系统！")
}

// createMCPClient 创建MCP客户端
func createMCPClient(aiConfig *config.AIConfig) (*mcp.Client, error) {
	client := mcp.New()

	switch aiConfig.Provider {
	case "deepseek":
		client.SetDeepSeekAPIKey(aiConfig.DeepSeekKey)
	case "qwen":
		client.SetQwenAPIKey(aiConfig.QwenKey, "")
	case "custom":
		client.SetCustomAPI(aiConfig.CustomAPIURL, aiConfig.CustomAPIKey, aiConfig.CustomModelName)
	default:
		return nil, fmt.Errorf("不支持的AI提供商: %s", aiConfig.Provider)
	}

	return client, nil
}

// createNotifier 创建通知器
func createNotifier(notifConfig *config.NotificationConfig) notifier.Notifier {
	var notifiers []notifier.Notifier

	if notifConfig.DingTalk.Enabled {
		ding := notifier.NewDingTalkNotifier(
			notifConfig.DingTalk.WebhookURL,
			notifConfig.DingTalk.Secret,
		)
		notifiers = append(notifiers, ding)
		log.Printf("  ✓ 钉钉通知已启用")
	}

	if notifConfig.Feishu.Enabled {
		feishu := notifier.NewFeishuNotifier(
			notifConfig.Feishu.WebhookURL,
			notifConfig.Feishu.Secret,
		)
		notifiers = append(notifiers, feishu)
		log.Printf("  ✓ 飞书通知已启用")
	}

	if len(notifiers) == 0 {
		return nil
	}

	if len(notifiers) == 1 {
		return notifiers[0]
	}

	return notifier.NewMultiNotifier(notifiers...)
}

// parseBuyDate 解析购买日期字符串为time.Time
func parseBuyDate(dateStr string) time.Time {
	if dateStr == "" {
		return time.Time{} // 零值
	}
	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		log.Printf("⚠️  解析购买日期失败: %v，将忽略该字段", err)
		return time.Time{}
	}
	return t
}

// AnalyzerManager 分析器管理器
type AnalyzerManager struct {
	analyzers        map[string]*stock.StockAnalyzer
	stopChans        map[string]chan struct{}
	analysisHistory  map[string][]*stock.AnalysisResult // 存储最近的分析结果（每个股票代码对应一个结果列表）
	maxHistorySize   int                                  // 每个股票最多保存的分析记录数
	analysisMode     string                               // 分析模式：smart/concurrent/polling
	maxConcurrent    int                                  // 最大并发分析数
	stockCount       int                                  // 启用的股票数量
	mutex            sync.RWMutex
	semaphore        chan struct{}                        // 并发控制信号量（用于限制并发数）
}

// AddAnalyzer 添加分析器
func (m *AnalyzerManager) AddAnalyzer(code string, analyzer *stock.StockAnalyzer) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.analyzers[code] = analyzer
	m.stopChans[code] = make(chan struct{})
}

// GetAnalyzer 获取分析器
func (m *AnalyzerManager) GetAnalyzer(code string) interface{} {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.analyzers[code]
}

// TriggerAnalysis 手动触发分析
func (m *AnalyzerManager) TriggerAnalysis(code string) (interface{}, error) {
	m.mutex.RLock()
	analyzer, exists := m.analyzers[code]
	m.mutex.RUnlock()
	
	if !exists {
		return nil, fmt.Errorf("股票代码 %s 的分析器不存在", code)
	}
	
	result, err := analyzer.Analyze()
	if err != nil {
		return nil, err
	}
	
	// 保存分析结果到历史记录
	if result != nil {
		m.saveAnalysisResult(code, result)
	}
	
	return result, nil
}

// saveAnalysisResult 保存分析结果到历史记录
func (m *AnalyzerManager) saveAnalysisResult(code string, result *stock.AnalysisResult) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.analysisHistory == nil {
		m.analysisHistory = make(map[string][]*stock.AnalysisResult)
	}

	history := m.analysisHistory[code]
	if history == nil {
		history = []*stock.AnalysisResult{}
	}

	// 添加到列表开头（最新的在前面）
	history = append([]*stock.AnalysisResult{result}, history...)

	// 限制历史记录数量
	if len(history) > m.maxHistorySize {
		history = history[:m.maxHistorySize]
	}

	m.analysisHistory[code] = history
}

// GetAnalysisHistory 获取分析历史记录
func (m *AnalyzerManager) GetAnalysisHistory(code string, limit int) interface{} {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if limit <= 0 {
		limit = 20 // 默认20条
	}

	history := m.analysisHistory[code]
	if history == nil {
		return []*stock.AnalysisResult{}
	}

	if len(history) > limit {
		return history[:limit]
	}

	return history
}

// GetAllRecentAnalysis 获取所有股票的最远分析记录（最近N条）
func (m *AnalyzerManager) GetAllRecentAnalysis(limit int) interface{} {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if limit <= 0 {
		limit = 10 // 默认10条
	}

	var allResults []*stock.AnalysisResult

	// 收集所有股票的最新分析结果
	for _, history := range m.analysisHistory {
		if len(history) > 0 {
			// 只取每个股票的最新一条
			allResults = append(allResults, history[0])
		}
	}

	// 按时间排序（最新的在前）
	for i := 0; i < len(allResults)-1; i++ {
		for j := i + 1; j < len(allResults); j++ {
			if allResults[i].Timestamp.Before(allResults[j].Timestamp) {
				allResults[i], allResults[j] = allResults[j], allResults[i]
			}
		}
	}

	// 限制返回数量
	if len(allResults) > limit {
		return allResults[:limit]
	}

	return allResults
}

// StartAll 启动所有分析器
func (m *AnalyzerManager) StartAll() {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// 确定实际使用的分析模式和并发数
	actualMode, actualMaxConcurrent := m.determineAnalysisMode()

	log.Printf("📊 分析模式: %s，最大并发数: %d，股票总数: %d", actualMode, actualMaxConcurrent, m.stockCount)

	// 初始化并发控制信号量
	if actualMode == "concurrent" || actualMode == "smart" {
		m.semaphore = make(chan struct{}, actualMaxConcurrent)
	}

	// 如果是轮询模式，使用轮询方式启动
	if actualMode == "polling" {
		m.startPollingMode()
		return
	}

	// 并发模式或智能模式，使用并发方式启动
	for code, analyzer := range m.analyzers {
		stopChan := m.stopChans[code]
		go func(code string, analyzer *stock.StockAnalyzer, stopChan chan struct{}) {
			// 包装监控函数，在分析完成后保存结果
			ticker := time.NewTicker(analyzer.AnalysisConfig.ScanInterval)
			defer ticker.Stop()

			log.Printf("🚀 开始监控股票 %s，扫描间隔: %v",
				code,
				analyzer.AnalysisConfig.ScanInterval)

			// 立即执行一次分析（带并发控制）
			m.runAnalysisWithSemaphore(code, analyzer)

			for {
				select {
				case <-ticker.C:
					m.runAnalysisWithSemaphore(code, analyzer)
				case <-stopChan:
					log.Printf("⏹️  停止监控股票 %s", code)
					return
				}
			}
		}(code, analyzer, stopChan)
	}
}

// determineAnalysisMode 确定实际使用的分析模式和并发数
func (m *AnalyzerManager) determineAnalysisMode() (string, int) {
	if m.analysisMode == "polling" {
		return "polling", 1
	}

	if m.analysisMode == "concurrent" {
		return "concurrent", m.maxConcurrent
	}

	// 智能模式：根据股票数量自动选择
	if m.stockCount <= 4 {
		// 股票数量 <= 4，使用并发，并发数 = 股票数（最多4个）
		maxConcurrent := m.stockCount
		if maxConcurrent > 4 {
			maxConcurrent = 4
		}
		return "concurrent", maxConcurrent
	}

	// 股票数量 > 4，使用轮询模式
	return "polling", 1
}

// runAnalysisWithSemaphore 带并发控制的分析执行
func (m *AnalyzerManager) runAnalysisWithSemaphore(code string, analyzer *stock.StockAnalyzer) {
	if m.semaphore == nil {
		// 如果没有信号量（轮询模式），直接执行
		if result, err := analyzer.Analyze(); err == nil && result != nil {
			m.saveAnalysisResult(code, result)
		}
		return
	}

	// 获取信号量（控制并发数）
	m.semaphore <- struct{}{}
	defer func() { <-m.semaphore }()

	if result, err := analyzer.Analyze(); err == nil && result != nil {
		m.saveAnalysisResult(code, result)
	}
}

// startPollingMode 启动轮询模式（顺序分析）
func (m *AnalyzerManager) startPollingMode() {
	// 收集所有分析器和对应的停止通道
	type analyzerInfo struct {
		code     string
		analyzer *stock.StockAnalyzer
		stopChan chan struct{}
		interval time.Duration
	}

	var analyzers []analyzerInfo
	for code, analyzer := range m.analyzers {
		analyzers = append(analyzers, analyzerInfo{
			code:     code,
			analyzer: analyzer,
			stopChan: m.stopChans[code],
			interval: analyzer.AnalysisConfig.ScanInterval,
		})
		log.Printf("🚀 准备监控股票 %s，扫描间隔: %v", code, analyzer.AnalysisConfig.ScanInterval)
	}

	// 启动轮询协程（顺序分析）
	go func() {
		log.Printf("🔄 启动轮询模式，顺序分析 %d 只股票", len(analyzers))

		// 立即执行一轮分析（顺序执行）
		for _, info := range analyzers {
			select {
			case <-info.stopChan:
				log.Printf("⏹️  停止监控股票 %s", info.code)
				return
			default:
				log.Printf("📊 [轮询] 开始分析股票 %s", info.code)
				if result, err := info.analyzer.Analyze(); err == nil && result != nil {
					m.saveAnalysisResult(info.code, result)
				}
				log.Printf("✅ [轮询] 完成分析股票 %s", info.code)
			}
		}

		// 记录每个股票的上次分析时间
		lastAnalysis := make(map[string]time.Time)
		for _, info := range analyzers {
			lastAnalysis[info.code] = time.Now()
		}

		// 计算最短间隔（用于主循环）
		minInterval := time.Minute * 5 // 默认5分钟
		for _, info := range analyzers {
			if info.interval < minInterval {
				minInterval = info.interval
			}
		}

		// 主轮询循环
		ticker := time.NewTicker(minInterval / 4) // 每1/4间隔检查一次
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				// 检查每个股票是否需要分析
				for i, info := range analyzers {
					select {
					case <-info.stopChan:
						log.Printf("⏹️  停止监控股票 %s", info.code)
						// 从列表中移除已停止的股票
						analyzers = append(analyzers[:i], analyzers[i+1:]...)
						delete(lastAnalysis, info.code)

						// 如果所有股票都停止了，退出
						if len(analyzers) == 0 {
							log.Printf("⏹️  所有股票监控已停止")
							return
						}
						goto nextCheck // 重新开始检查
					default:
						// 检查是否到了该股票的分析时间
						if time.Since(lastAnalysis[info.code]) >= info.interval {
							log.Printf("📊 [轮询] 开始分析股票 %s（第 %d/%d 只）", info.code, i+1, len(analyzers))
							if result, err := info.analyzer.Analyze(); err == nil && result != nil {
								m.saveAnalysisResult(info.code, result)
							}
							lastAnalysis[info.code] = time.Now()
							log.Printf("✅ [轮询] 完成分析股票 %s", info.code)
						}
					}
				}
			nextCheck:
				// 继续下一轮检查
			}
		}
	}()
}

// StopAll 停止所有分析器
func (m *AnalyzerManager) StopAll() {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	for _, stopChan := range m.stopChans {
		close(stopChan)
	}
}

// GetAllAnalyzers 获取所有分析器
func (m *AnalyzerManager) GetAllAnalyzers() map[string]interface{} {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	result := make(map[string]interface{})
	for code, analyzer := range m.analyzers {
		result[code] = analyzer
	}
	return result
}
