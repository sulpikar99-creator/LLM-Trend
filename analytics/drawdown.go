package analytics

import (
	"fmt"
	"math"
	"time"
)

// EquityPoint represents a point in the equity curve
type EquityPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Equity    float64   `json:"equity"`
	Balance   float64   `json:"balance"`
	UnrealizedPnL float64 `json:"unrealized_pnl"`
}

// DrawdownPeriod represents a drawdown period
type DrawdownPeriod struct {
	Start      time.Time `json:"start"`
	End        time.Time `json:"end,omitempty"`
	Bottom     time.Time `json:"bottom"`
	Peak       float64   `json:"peak"`
	Trough     float64   `json:"trough"`
	Drawdown   float64   `json:"drawdown"`
	DrawdownPct float64  `json:"drawdown_pct"`
	Recovery   *time.Time `json:"recovery,omitempty"`
	Duration   float64    `json:"duration_hours"`
	Recovered  bool       `json:"recovered"`
}

// DrawdownAnalysis contains comprehensive drawdown metrics
type DrawdownAnalysis struct {
	MaxDrawdown        float64           `json:"max_drawdown"`
	MaxDrawdownPct     float64           `json:"max_drawdown_pct"`
	CurrentDrawdown    float64           `json:"current_drawdown"`
	CurrentDrawdownPct float64           `json:"current_drawdown_pct"`
	AvgDrawdown        float64           `json:"avg_drawdown"`
	AvgDrawdownPct     float64           `json:"avg_drawdown_pct"`
	DrawdownPeriods    []*DrawdownPeriod `json:"drawdown_periods"`
	TotalPeriods       int               `json:"total_periods"`
	AvgRecoveryTime    float64           `json:"avg_recovery_time_hours"`
	LongestDrawdown    float64           `json:"longest_drawdown_hours"`
	CurrentPeak        float64           `json:"current_peak"`
	InDrawdown         bool              `json:"in_drawdown"`
}

// CalculateDrawdown calculates comprehensive drawdown analysis
func CalculateDrawdown(points []EquityPoint) (*DrawdownAnalysis, error) {
	if len(points) < 2 {
		return &DrawdownAnalysis{
			MaxDrawdown:     0,
			MaxDrawdownPct:  0,
			CurrentDrawdown: 0,
			DrawdownPeriods: []*DrawdownPeriod{},
			TotalPeriods:    0,
		}, nil
	}

	var (
		peak           = points[0].Equity
		peakTime       = points[0].Timestamp
		maxDrawdown    = 0.0
		maxDrawdownPct = 0.0
		currentDD      = 0.0
		currentDDPct   = 0.0
		periods        []*DrawdownPeriod
		currentPeriod  *DrawdownPeriod
	)

	for i, point := range points {
		equity := point.Equity

		// Update peak
		if equity > peak {
			// If we were in drawdown, mark recovery
			if currentPeriod != nil && !currentPeriod.Recovered {
				currentPeriod.Recovery = &point.Timestamp
				currentPeriod.End = point.Timestamp
				currentPeriod.Duration = point.Timestamp.Sub(currentPeriod.Start).Hours()
				currentPeriod.Recovered = true
				currentPeriod = nil
			}

			peak = equity
			peakTime = point.Timestamp
		}

		// Calculate current drawdown
		if peak > 0 {
			dd := peak - equity
			ddPct := (dd / peak) * 100

			// Start new drawdown period if we dropped below peak
			if dd > 0 && currentPeriod == nil {
				currentPeriod = &DrawdownPeriod{
					Start:      peakTime,
					Peak:       peak,
					Trough:     equity,
					Bottom:     point.Timestamp,
					Drawdown:   dd,
					DrawdownPct: ddPct,
					Recovered:  false,
				}
				periods = append(periods, currentPeriod)
			}

			// Update current period
			if currentPeriod != nil && equity < currentPeriod.Trough {
				currentPeriod.Trough = equity
				currentPeriod.Bottom = point.Timestamp
				currentPeriod.Drawdown = peak - equity
				if peak > 0 {
					currentPeriod.DrawdownPct = ((peak - equity) / peak) * 100
				}
			}

			// Update max drawdown
			if dd > maxDrawdown {
				maxDrawdown = dd
				maxDrawdownPct = ddPct
			}

			// Current drawdown (for last point)
			if i == len(points)-1 {
				currentDD = dd
				currentDDPct = ddPct
			}
		}
	}

	// Mark unrecovered periods
	if currentPeriod != nil && !currentPeriod.Recovered {
		currentPeriod.End = points[len(points)-1].Timestamp
		currentPeriod.Duration = currentPeriod.End.Sub(currentPeriod.Start).Hours()
	}

	// Calculate statistics
	totalDD := 0.0
	totalDDPct := 0.0
	totalRecoveryTime := 0.0
	recoveredCount := 0
	longestDD := 0.0

	for _, period := range periods {
		totalDD += period.Drawdown
		totalDDPct += period.DrawdownPct

		if period.Recovered && period.Recovery != nil {
			recoveryTime := period.Recovery.Sub(period.Start).Hours()
			totalRecoveryTime += recoveryTime
			recoveredCount++
		}

		if period.Duration > longestDD {
			longestDD = period.Duration
		}
	}

	avgDD := 0.0
	avgDDPct := 0.0
	avgRecovery := 0.0

	if len(periods) > 0 {
		avgDD = totalDD / float64(len(periods))
		avgDDPct = totalDDPct / float64(len(periods))
	}

	if recoveredCount > 0 {
		avgRecovery = totalRecoveryTime / float64(recoveredCount)
	}

	return &DrawdownAnalysis{
		MaxDrawdown:        maxDrawdown,
		MaxDrawdownPct:     maxDrawdownPct,
		CurrentDrawdown:    currentDD,
		CurrentDrawdownPct: currentDDPct,
		AvgDrawdown:        avgDD,
		AvgDrawdownPct:     avgDDPct,
		DrawdownPeriods:    periods,
		TotalPeriods:       len(periods),
		AvgRecoveryTime:    avgRecovery,
		LongestDrawdown:    longestDD,
		CurrentPeak:        peak,
		InDrawdown:         currentPeriod != nil && !currentPeriod.Recovered,
	}, nil
}

// GetDrawdownSummary returns a simplified summary for quick display
func (da *DrawdownAnalysis) GetDrawdownSummary() map[string]interface{} {
	return map[string]interface{}{
		"max_drawdown_pct":     fmt.Sprintf("%.2f%%", da.MaxDrawdownPct),
		"current_drawdown_pct": fmt.Sprintf("%.2f%%", da.CurrentDrawdownPct),
		"avg_recovery_hours":   fmt.Sprintf("%.1f", da.AvgRecoveryTime),
		"total_periods":        da.TotalPeriods,
		"in_drawdown":          da.InDrawdown,
	}
}

// CalculateDrawdownFromBalances is a convenience function for balance history
func CalculateDrawdownFromBalances(balances []float64, timestamps []time.Time) (*DrawdownAnalysis, error) {
	if len(balances) != len(timestamps) {
		return nil, fmt.Errorf("balances and timestamps must have same length")
	}

	if len(balances) == 0 {
		return nil, fmt.Errorf("no balance data provided")
	}

	points := make([]EquityPoint, len(balances))
	for i := range balances {
		points[i] = EquityPoint{
			Timestamp: timestamps[i],
			Equity:    balances[i],
			Balance:   balances[i],
		}
	}

	return CalculateDrawdown(points)
}

// GetWorstDrawdowns returns the N worst drawdown periods
func (da *DrawdownAnalysis) GetWorstDrawdowns(n int) []*DrawdownPeriod {
	if len(da.DrawdownPeriods) == 0 {
		return []*DrawdownPeriod{}
	}

	// Create a copy and sort by drawdown percentage
	periods := make([]*DrawdownPeriod, len(da.DrawdownPeriods))
	copy(periods, da.DrawdownPeriods)

	// Simple bubble sort (fine for small N)
	for i := 0; i < len(periods)-1; i++ {
		for j := 0; j < len(periods)-i-1; j++ {
			if periods[j].DrawdownPct < periods[j+1].DrawdownPct {
				periods[j], periods[j+1] = periods[j+1], periods[j]
			}
		}
	}

	if n > len(periods) {
		n = len(periods)
	}

	return periods[:n]
}

// CalculateUnderwaterPeriod calculates underwater chart data (distance from peak)
func CalculateUnderwaterPeriod(points []EquityPoint) []float64 {
	if len(points) == 0 {
		return []float64{}
	}

	underwater := make([]float64, len(points))
	peak := points[0].Equity

	for i, point := range points {
		if point.Equity > peak {
			peak = point.Equity
		}

		if peak > 0 {
			underwater[i] = ((point.Equity - peak) / peak) * 100
		} else {
			underwater[i] = 0
		}
	}

	return underwater
}

// CalculateMaxAdverseExcursion calculates MAE for trades
func CalculateMaxAdverseExcursion(entryPrice, lowestPrice float64, isLong bool) float64 {
	if isLong {
		// For long positions, MAE is the maximum drop below entry
		mae := entryPrice - lowestPrice
		if mae < 0 {
			return 0 // No adverse excursion if lowest > entry
		}
		return (mae / entryPrice) * 100
	}

	// For short positions, MAE is the maximum rise above entry
	mae := lowestPrice - entryPrice
	if mae < 0 {
		return 0
	}
	if entryPrice > 0 {
		return (mae / entryPrice) * 100
	}
	return 0
}

// CalculateMaxFavorableExcursion calculates MFE for trades
func CalculateMaxFavorableExcursion(entryPrice, highestPrice float64, isLong bool) float64 {
	if isLong {
		// For long positions, MFE is the maximum gain above entry
		mfe := highestPrice - entryPrice
		if mfe < 0 {
			return 0
		}
		return (mfe / entryPrice) * 100
	}

	// For short positions, MFE is the maximum gain below entry
	mfe := entryPrice - highestPrice
	if mfe < 0 {
		return 0
	}
	if entryPrice > 0 {
		return (mfe / entryPrice) * 100
	}
	return 0
}

// CalculateRecoveryFactor calculates profit/max drawdown ratio
func CalculateRecoveryFactor(totalProfit, maxDrawdown float64) float64 {
	if maxDrawdown == 0 {
		if totalProfit > 0 {
			return math.Inf(1) // Infinite recovery factor (no drawdown)
		}
		return 0
	}

	return totalProfit / maxDrawdown
}
