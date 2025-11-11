package analytics

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"time"
)

// MonteCarloConfig contains simulation parameters
type MonteCarloConfig struct {
	InitialBalance float64 `json:"initial_balance"`
	NumSimulations int     `json:"num_simulations"`
	NumPeriods     int     `json:"num_periods"`
	MeanReturn     float64 `json:"mean_return"`      // Average return per period
	StdDevReturn   float64 `json:"std_dev_return"`   // Standard deviation of returns
	WinRate        float64 `json:"win_rate"`         // Win rate (0-1)
	AvgWin         float64 `json:"avg_win"`          // Average win amount
	AvgLoss        float64 `json:"avg_loss"`         // Average loss amount
	RiskFreeRate   float64 `json:"risk_free_rate"`   // For Sharpe calculation
	UseHistorical  bool    `json:"use_historical"`   // Use historical returns
	Seed           int64   `json:"seed,omitempty"`   // Random seed for reproducibility
}

// MonteCarloResult contains simulation results
type MonteCarloResult struct {
	Paths              [][]float64           `json:"paths"`                 // All simulation paths
	FinalValues        []float64             `json:"final_values"`          // Final equity for each simulation
	Percentiles        map[string]float64    `json:"percentiles"`           // 5th, 25th, 50th, 75th, 95th percentiles
	Mean               float64               `json:"mean"`                  // Mean final equity
	StdDev             float64               `json:"std_dev"`               // Standard deviation
	MedianReturn       float64               `json:"median_return"`         // Median return percentage
	ProbabilityProfit  float64               `json:"probability_profit"`    // Probability of profit
	ProbabilityLoss    float64               `json:"probability_loss"`      // Probability of loss
	ValueAtRisk95      float64               `json:"var_95"`                // 95% VaR
	ValueAtRisk99      float64               `json:"var_99"`                // 99% VaR
	ConditionalVaR95   float64               `json:"cvar_95"`               // Expected shortfall at 95%
	MaxDrawdownDist    []float64             `json:"max_drawdown_dist"`     // Distribution of max drawdowns
	AvgMaxDrawdown     float64               `json:"avg_max_drawdown"`      // Average max drawdown across paths
	WorstCaseScenario  []float64             `json:"worst_case_scenario"`   // Worst path
	BestCaseScenario   []float64             `json:"best_case_scenario"`    // Best path
	ConfidenceIntervals map[string][]float64 `json:"confidence_intervals"` // CI bounds per period
}

// RunMonteCarloSimulation runs Monte Carlo simulation
func RunMonteCarloSimulation(config *MonteCarloConfig) (*MonteCarloResult, error) {
	if config.NumSimulations <= 0 {
		return nil, fmt.Errorf("num_simulations must be positive")
	}
	if config.NumPeriods <= 0 {
		return nil, fmt.Errorf("num_periods must be positive")
	}
	if config.InitialBalance <= 0 {
		return nil, fmt.Errorf("initial_balance must be positive")
	}

	// Initialize random number generator
	var rng *rand.Rand
	if config.Seed != 0 {
		rng = rand.New(rand.NewSource(config.Seed))
	} else {
		rng = rand.New(rand.NewSource(time.Now().UnixNano()))
	}

	result := &MonteCarloResult{
		Paths:       make([][]float64, config.NumSimulations),
		FinalValues: make([]float64, config.NumSimulations),
		Percentiles: make(map[string]float64),
		MaxDrawdownDist: make([]float64, config.NumSimulations),
		ConfidenceIntervals: make(map[string][]float64),
	}

	// Run simulations
	for sim := 0; sim < config.NumSimulations; sim++ {
		path := make([]float64, config.NumPeriods+1)
		path[0] = config.InitialBalance

		equity := config.InitialBalance
		peak := equity
		maxDD := 0.0

		for period := 1; period <= config.NumPeriods; period++ {
			// Generate return based on strategy
			var periodReturn float64

			if config.UseHistorical {
				// Use normal distribution with mean and std dev
				periodReturn = rng.NormFloat64()*config.StdDevReturn + config.MeanReturn
			} else {
				// Use win/loss distribution
				if rng.Float64() < config.WinRate {
					periodReturn = config.AvgWin
				} else {
					periodReturn = -config.AvgLoss
				}
			}

			// Apply return to equity
			equity = equity * (1 + periodReturn)

			// Prevent negative equity
			if equity < 0 {
				equity = 0
			}

			path[period] = equity

			// Track drawdown
			if equity > peak {
				peak = equity
			}
			if peak > 0 {
				dd := ((peak - equity) / peak) * 100
				if dd > maxDD {
					maxDD = dd
				}
			}
		}

		result.Paths[sim] = path
		result.FinalValues[sim] = equity
		result.MaxDrawdownDist[sim] = maxDD
	}

	// Calculate statistics
	result.calculateStatistics(config)

	return result, nil
}

// calculateStatistics computes summary statistics from simulation results
func (r *MonteCarloResult) calculateStatistics(config *MonteCarloConfig) {
	if len(r.FinalValues) == 0 {
		return
	}

	// Sort final values for percentile calculations
	sortedValues := make([]float64, len(r.FinalValues))
	copy(sortedValues, r.FinalValues)
	sort.Float64s(sortedValues)

	// Calculate percentiles
	r.Percentiles["5"] = percentile(sortedValues, 5)
	r.Percentiles["25"] = percentile(sortedValues, 25)
	r.Percentiles["50"] = percentile(sortedValues, 50)
	r.Percentiles["75"] = percentile(sortedValues, 75)
	r.Percentiles["95"] = percentile(sortedValues, 95)

	// Calculate mean and std dev
	sum := 0.0
	for _, val := range r.FinalValues {
		sum += val
	}
	r.Mean = sum / float64(len(r.FinalValues))

	variance := 0.0
	for _, val := range r.FinalValues {
		diff := val - r.Mean
		variance += diff * diff
	}
	r.StdDev = math.Sqrt(variance / float64(len(r.FinalValues)))

	// Median return percentage
	if config.InitialBalance > 0 {
		r.MedianReturn = ((r.Percentiles["50"] - config.InitialBalance) / config.InitialBalance) * 100
	}

	// Probability of profit/loss
	profitCount := 0
	for _, val := range r.FinalValues {
		if val > config.InitialBalance {
			profitCount++
		}
	}
	r.ProbabilityProfit = float64(profitCount) / float64(len(r.FinalValues))
	r.ProbabilityLoss = 1.0 - r.ProbabilityProfit

	// Value at Risk (VaR)
	r.ValueAtRisk95 = config.InitialBalance - percentile(sortedValues, 5)
	r.ValueAtRisk99 = config.InitialBalance - percentile(sortedValues, 1)

	// Conditional VaR (Expected Shortfall) at 95%
	idx95 := int(0.05 * float64(len(sortedValues)))
	if idx95 > 0 {
		cvarSum := 0.0
		for i := 0; i < idx95; i++ {
			cvarSum += sortedValues[i]
		}
		r.ConditionalVaR95 = config.InitialBalance - (cvarSum / float64(idx95))
	}

	// Average max drawdown
	ddSum := 0.0
	for _, dd := range r.MaxDrawdownDist {
		ddSum += dd
	}
	if len(r.MaxDrawdownDist) > 0 {
		r.AvgMaxDrawdown = ddSum / float64(len(r.MaxDrawdownDist))
	}

	// Find best and worst paths
	bestIdx := 0
	worstIdx := 0
	for i, val := range r.FinalValues {
		if val > r.FinalValues[bestIdx] {
			bestIdx = i
		}
		if val < r.FinalValues[worstIdx] {
			worstIdx = i
		}
	}
	r.BestCaseScenario = r.Paths[bestIdx]
	r.WorstCaseScenario = r.Paths[worstIdx]

	// Calculate confidence intervals per period
	r.calculateConfidenceIntervals()
}

// calculateConfidenceIntervals calculates CI bounds for each period
func (r *MonteCarloResult) calculateConfidenceIntervals() {
	if len(r.Paths) == 0 || len(r.Paths[0]) == 0 {
		return
	}

	numPeriods := len(r.Paths[0])
	lower := make([]float64, numPeriods)
	upper := make([]float64, numPeriods)
	median := make([]float64, numPeriods)

	for period := 0; period < numPeriods; period++ {
		// Collect all values at this period
		values := make([]float64, len(r.Paths))
		for sim := 0; sim < len(r.Paths); sim++ {
			values[sim] = r.Paths[sim][period]
		}

		sort.Float64s(values)

		lower[period] = percentile(values, 5)
		median[period] = percentile(values, 50)
		upper[period] = percentile(values, 95)
	}

	r.ConfidenceIntervals["lower_5"] = lower
	r.ConfidenceIntervals["median"] = median
	r.ConfidenceIntervals["upper_95"] = upper
}

// percentile calculates the nth percentile of sorted data
func percentile(sortedData []float64, p float64) float64 {
	if len(sortedData) == 0 {
		return 0
	}

	if p <= 0 {
		return sortedData[0]
	}
	if p >= 100 {
		return sortedData[len(sortedData)-1]
	}

	// Linear interpolation
	rank := (p / 100) * float64(len(sortedData)-1)
	lower := int(math.Floor(rank))
	upper := int(math.Ceil(rank))

	if lower == upper {
		return sortedData[lower]
	}

	weight := rank - float64(lower)
	return sortedData[lower]*(1-weight) + sortedData[upper]*weight
}

// EstimateParametersFromHistory estimates MC parameters from historical data
func EstimateParametersFromHistory(returns []float64) *MonteCarloConfig {
	if len(returns) == 0 {
		return &MonteCarloConfig{
			MeanReturn:   0,
			StdDevReturn: 0,
			WinRate:      0.5,
		}
	}

	// Calculate mean return
	sum := 0.0
	wins := 0
	winSum := 0.0
	lossSum := 0.0
	lossCount := 0

	for _, ret := range returns {
		sum += ret
		if ret > 0 {
			wins++
			winSum += ret
		} else if ret < 0 {
			lossSum += math.Abs(ret)
			lossCount++
		}
	}

	mean := sum / float64(len(returns))

	// Calculate standard deviation
	variance := 0.0
	for _, ret := range returns {
		diff := ret - mean
		variance += diff * diff
	}
	stdDev := math.Sqrt(variance / float64(len(returns)))

	// Calculate win/loss statistics
	winRate := float64(wins) / float64(len(returns))
	avgWin := 0.0
	avgLoss := 0.0

	if wins > 0 {
		avgWin = winSum / float64(wins)
	}
	if lossCount > 0 {
		avgLoss = lossSum / float64(lossCount)
	}

	return &MonteCarloConfig{
		MeanReturn:   mean,
		StdDevReturn: stdDev,
		WinRate:      winRate,
		AvgWin:       avgWin,
		AvgLoss:      avgLoss,
	}
}

// GetSummary returns a simplified summary for display
func (r *MonteCarloResult) GetSummary() map[string]interface{} {
	return map[string]interface{}{
		"median_final":      r.Percentiles["50"],
		"median_return_pct": fmt.Sprintf("%.2f%%", r.MedianReturn),
		"prob_profit":       fmt.Sprintf("%.1f%%", r.ProbabilityProfit*100),
		"avg_max_dd":        fmt.Sprintf("%.2f%%", r.AvgMaxDrawdown),
		"var_95":            fmt.Sprintf("%.2f", r.ValueAtRisk95),
	}
}

// CalculateSharpeRatio calculates Sharpe ratio from simulation results
func CalculateSharpeRatio(returns []float64, riskFreeRate float64) float64 {
	if len(returns) == 0 {
		return 0
	}

	// Calculate excess returns
	excessReturns := make([]float64, len(returns))
	sum := 0.0
	for i, ret := range returns {
		excessReturns[i] = ret - riskFreeRate
		sum += excessReturns[i]
	}

	mean := sum / float64(len(excessReturns))

	// Calculate standard deviation of excess returns
	variance := 0.0
	for _, ret := range excessReturns {
		diff := ret - mean
		variance += diff * diff
	}
	stdDev := math.Sqrt(variance / float64(len(excessReturns)))

	if stdDev == 0 {
		return 0
	}

	return mean / stdDev
}

// CalculateSortinoRatio calculates Sortino ratio (downside deviation)
func CalculateSortinoRatio(returns []float64, targetReturn float64) float64 {
	if len(returns) == 0 {
		return 0
	}

	// Calculate mean return
	sum := 0.0
	for _, ret := range returns {
		sum += ret
	}
	mean := sum / float64(len(returns))

	// Calculate downside deviation (only negative deviations)
	downsideVariance := 0.0
	count := 0
	for _, ret := range returns {
		if ret < targetReturn {
			diff := ret - targetReturn
			downsideVariance += diff * diff
			count++
		}
	}

	if count == 0 {
		return math.Inf(1)
	}

	downsideStdDev := math.Sqrt(downsideVariance / float64(count))

	if downsideStdDev == 0 {
		return 0
	}

	return (mean - targetReturn) / downsideStdDev
}
