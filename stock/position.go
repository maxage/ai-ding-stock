package stock

import (
	"fmt"
	"time"
)

// PositionInfo 持仓信息
type PositionInfo struct {
	StockCode         string    `json:"stock_code"`
	StockName         string    `json:"stock_name"`
	Quantity          int       `json:"quantity"`        // 持仓数量（股）
	BuyPrice          float64   `json:"buy_price"`       // 购买价格（元/股）
	BuyDate           time.Time `json:"buy_date"`        // 购买日期
	CurrentPrice      float64   `json:"current_price"`   // 当前价格（元/股）
	TotalCost         float64   `json:"total_cost"`      // 持仓成本（元）
	MarketValue       float64   `json:"market_value"`    // 市值（元）
	ProfitLoss        float64   `json:"profit_loss"`     // 浮动盈亏（元）
	ProfitLossPercent float64   `json:"profit_loss_percent"` // 盈亏比例（%）
}

// CalculatePositionInfo 计算持仓信息
func CalculatePositionInfo(code, name string, quantity int, buyPrice, currentPrice float64, buyDate time.Time) *PositionInfo {
	totalCost := buyPrice * float64(quantity)
	marketValue := currentPrice * float64(quantity)
	profitLoss := marketValue - totalCost
	profitLossPercent := 0.0
	if buyPrice > 0 {
		profitLossPercent = ((currentPrice - buyPrice) / buyPrice) * 100.0
	}

	return &PositionInfo{
		StockCode:         code,
		StockName:         name,
		Quantity:          quantity,
		BuyPrice:          buyPrice,
		BuyDate:           buyDate,
		CurrentPrice:      currentPrice,
		TotalCost:         totalCost,
		MarketValue:       marketValue,
		ProfitLoss:        profitLoss,
		ProfitLossPercent: profitLossPercent,
	}
}

// FormatProfitLoss 格式化盈亏显示
func (p *PositionInfo) FormatProfitLoss() string {
	sign := "+"
	if p.ProfitLoss < 0 {
		sign = ""
	}
	return fmt.Sprintf("%s%.2f元 (%.2f%%)", sign, p.ProfitLoss, p.ProfitLossPercent)
}

