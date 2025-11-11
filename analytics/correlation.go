package analytics

import (
	"fmt"
	"math"
	"sort"
)

// PriceData represents price data for a symbol
type PriceData struct {
	Symbol    string    `json:"symbol"`
	Prices    []float64 `json:"prices"`
	Returns   []float64 `json:"returns"`
	Timestamp []int64   `json:"timestamps,omitempty"`
}

// CorrelationMatrix contains correlation data between symbols
type CorrelationMatrix struct {
	Symbols     []string             `json:"symbols"`
	Matrix      [][]float64          `json:"matrix"`       // NxN correlation matrix
	Heatmap     [][]float64          `json:"heatmap"`      // Same as matrix (for frontend)
	Pairs       []*CorrelationPair   `json:"pairs"`        // Pairwise correlations
	AvgCorr     float64              `json:"avg_correlation"`
	MaxCorr     float64              `json:"max_correlation"`
	MinCorr     float64              `json:"min_correlation"`
	Summary     string               `json:"summary"`
}

// CorrelationPair represents correlation between two symbols
type CorrelationPair struct {
	Symbol1     string  `json:"symbol1"`
	Symbol2     string  `json:"symbol2"`
	Correlation float64 `json:"correlation"`
	Strength    string  `json:"strength"` // "strong", "moderate", "weak"
	Direction   string  `json:"direction"` // "positive", "negative", "none"
}

// CalculateCorrelationMatrix computes correlation matrix for multiple symbols
func CalculateCorrelationMatrix(priceData []*PriceData) (*CorrelationMatrix, error) {
	if len(priceData) < 2 {
		return nil, fmt.Errorf("need at least 2 symbols for correlation analysis")
	}

	// Calculate returns if not provided
	for _, data := range priceData {
		if len(data.Returns) == 0 {
			data.Returns = calculateReturns(data.Prices)
		}
	}

	// Find minimum length (align all series)
	minLen := len(priceData[0].Returns)
	for _, data := range priceData {
		if len(data.Returns) < minLen {
			minLen = len(data.Returns)
		}
	}

	if minLen < 2 {
		return nil, fmt.Errorf("insufficient data points for correlation (need at least 2)")
	}

	n := len(priceData)
	matrix := make([][]float64, n)
	for i := range matrix {
		matrix[i] = make([]float64, n)
	}

	// Calculate pairwise correlations
	var pairs []*CorrelationPair
	sumCorr := 0.0
	count := 0
	maxCorr := -1.0
	minCorr := 1.0

	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			if i == j {
				matrix[i][j] = 1.0
				continue
			}

			// Calculate Pearson correlation
			corr := pearsonCorrelation(
				priceData[i].Returns[:minLen],
				priceData[j].Returns[:minLen],
			)

			matrix[i][j] = corr

			// Track statistics (only upper triangle to avoid duplicates)
			if i < j {
				strength, direction := classifyCorrelation(corr)
				pairs = append(pairs, &CorrelationPair{
					Symbol1:     priceData[i].Symbol,
					Symbol2:     priceData[j].Symbol,
					Correlation: corr,
					Strength:    strength,
					Direction:   direction,
				})

				sumCorr += corr
				count++

				if corr > maxCorr {
					maxCorr = corr
				}
				if corr < minCorr {
					minCorr = corr
				}
			}
		}
	}

	avgCorr := 0.0
	if count > 0 {
		avgCorr = sumCorr / float64(count)
	}

	// Extract symbol names
	symbols := make([]string, n)
	for i, data := range priceData {
		symbols[i] = data.Symbol
	}

	// Sort pairs by absolute correlation (strongest first)
	sort.Slice(pairs, func(i, j int) bool {
		return math.Abs(pairs[i].Correlation) > math.Abs(pairs[j].Correlation)
	})

	summary := generateCorrelationSummary(avgCorr, maxCorr, minCorr, len(pairs))

	return &CorrelationMatrix{
		Symbols:  symbols,
		Matrix:   matrix,
		Heatmap:  matrix, // Same data, frontend may expect "heatmap"
		Pairs:    pairs,
		AvgCorr:  avgCorr,
		MaxCorr:  maxCorr,
		MinCorr:  minCorr,
		Summary:  summary,
	}, nil
}

// pearsonCorrelation calculates Pearson correlation coefficient
func pearsonCorrelation(x, y []float64) float64 {
	if len(x) != len(y) || len(x) == 0 {
		return 0
	}

	n := float64(len(x))

	// Calculate means
	sumX := 0.0
	sumY := 0.0
	for i := range x {
		sumX += x[i]
		sumY += y[i]
	}
	meanX := sumX / n
	meanY := sumY / n

	// Calculate correlation
	numerator := 0.0
	sumSqX := 0.0
	sumSqY := 0.0

	for i := range x {
		diffX := x[i] - meanX
		diffY := y[i] - meanY

		numerator += diffX * diffY
		sumSqX += diffX * diffX
		sumSqY += diffY * diffY
	}

	denominator := math.Sqrt(sumSqX * sumSqY)

	if denominator == 0 {
		return 0
	}

	return numerator / denominator
}

// calculateReturns calculates log returns from prices
func calculateReturns(prices []float64) []float64 {
	if len(prices) < 2 {
		return []float64{}
	}

	returns := make([]float64, len(prices)-1)
	for i := 1; i < len(prices); i++ {
		if prices[i-1] > 0 && prices[i] > 0 {
			returns[i-1] = math.Log(prices[i] / prices[i-1])
		} else {
			returns[i-1] = 0
		}
	}

	return returns
}

// classifyCorrelation classifies correlation by strength and direction
func classifyCorrelation(corr float64) (strength, direction string) {
	// Classify direction
	if corr > 0.1 {
		direction = "positive"
	} else if corr < -0.1 {
		direction = "negative"
	} else {
		direction = "none"
	}

	// Classify strength (based on absolute value)
	absCorr := math.Abs(corr)
	if absCorr > 0.7 {
		strength = "strong"
	} else if absCorr > 0.4 {
		strength = "moderate"
	} else {
		strength = "weak"
	}

	return strength, direction
}

// generateCorrelationSummary creates a text summary of correlations
func generateCorrelationSummary(avg, max, min float64, numPairs int) string {
	avgStrength, avgDir := classifyCorrelation(avg)
	return fmt.Sprintf(
		"Analyzed %d symbol pairs. Average correlation: %.3f (%s %s). Range: %.3f to %.3f",
		numPairs, avg, avgStrength, avgDir, min, max,
	)
}

// GetTopCorrelations returns the N strongest correlations
func (cm *CorrelationMatrix) GetTopCorrelations(n int) []*CorrelationPair {
	if n > len(cm.Pairs) {
		n = len(cm.Pairs)
	}
	return cm.Pairs[:n]
}

// GetCorrelation returns correlation between two symbols
func (cm *CorrelationMatrix) GetCorrelation(symbol1, symbol2 string) (float64, error) {
	idx1 := -1
	idx2 := -1

	for i, sym := range cm.Symbols {
		if sym == symbol1 {
			idx1 = i
		}
		if sym == symbol2 {
			idx2 = i
		}
	}

	if idx1 == -1 {
		return 0, fmt.Errorf("symbol %s not found", symbol1)
	}
	if idx2 == -1 {
		return 0, fmt.Errorf("symbol %s not found", symbol2)
	}

	return cm.Matrix[idx1][idx2], nil
}

// CalculateBeta calculates beta of an asset relative to market
func CalculateBeta(assetReturns, marketReturns []float64) (float64, error) {
	if len(assetReturns) != len(marketReturns) {
		return 0, fmt.Errorf("return series must have same length")
	}

	if len(assetReturns) < 2 {
		return 0, fmt.Errorf("need at least 2 data points")
	}

	// Calculate covariance between asset and market
	n := float64(len(assetReturns))

	// Calculate means
	sumAsset := 0.0
	sumMarket := 0.0
	for i := range assetReturns {
		sumAsset += assetReturns[i]
		sumMarket += marketReturns[i]
	}
	meanAsset := sumAsset / n
	meanMarket := sumMarket / n

	// Calculate covariance and market variance
	covariance := 0.0
	marketVariance := 0.0

	for i := range assetReturns {
		diffAsset := assetReturns[i] - meanAsset
		diffMarket := marketReturns[i] - meanMarket

		covariance += diffAsset * diffMarket
		marketVariance += diffMarket * diffMarket
	}

	covariance /= n
	marketVariance /= n

	if marketVariance == 0 {
		return 0, fmt.Errorf("market variance is zero")
	}

	beta := covariance / marketVariance
	return beta, nil
}

// CalculatePortfolioCorrelation calculates weighted portfolio correlation
func CalculatePortfolioCorrelation(weights []float64, corrMatrix [][]float64) float64 {
	if len(weights) == 0 || len(corrMatrix) == 0 {
		return 0
	}

	n := len(weights)
	totalCorr := 0.0
	weightSum := 0.0

	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			if i != j && i < len(corrMatrix) && j < len(corrMatrix[i]) {
				totalCorr += weights[i] * weights[j] * corrMatrix[i][j]
				weightSum += weights[i] * weights[j]
			}
		}
	}

	if weightSum == 0 {
		return 0
	}

	return totalCorr / weightSum
}

// CalculateRollingCorrelation calculates rolling correlation between two series
func CalculateRollingCorrelation(x, y []float64, window int) ([]float64, error) {
	if len(x) != len(y) {
		return nil, fmt.Errorf("series must have same length")
	}

	if window <= 1 || window > len(x) {
		return nil, fmt.Errorf("invalid window size")
	}

	result := make([]float64, len(x)-window+1)

	for i := 0; i <= len(x)-window; i++ {
		windowX := x[i : i+window]
		windowY := y[i : i+window]
		result[i] = pearsonCorrelation(windowX, windowY)
	}

	return result, nil
}

// DetectCorrelationBreakdown identifies periods where correlation breaks down
func DetectCorrelationBreakdown(rollingCorr []float64, threshold float64) []int {
	var breakdowns []int

	for i, corr := range rollingCorr {
		if math.Abs(corr) < threshold {
			breakdowns = append(breakdowns, i)
		}
	}

	return breakdowns
}

// CalculateCovariance calculates covariance between two series
func CalculateCovariance(x, y []float64) (float64, error) {
	if len(x) != len(y) {
		return 0, fmt.Errorf("series must have same length")
	}

	if len(x) == 0 {
		return 0, fmt.Errorf("empty series")
	}

	n := float64(len(x))

	// Calculate means
	sumX := 0.0
	sumY := 0.0
	for i := range x {
		sumX += x[i]
		sumY += y[i]
	}
	meanX := sumX / n
	meanY := sumY / n

	// Calculate covariance
	cov := 0.0
	for i := range x {
		cov += (x[i] - meanX) * (y[i] - meanY)
	}

	return cov / n, nil
}

// CalculateCovarianceMatrix calculates full covariance matrix
func CalculateCovarianceMatrix(data []*PriceData) ([][]float64, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("no data provided")
	}

	// Find minimum length
	minLen := len(data[0].Returns)
	for _, d := range data {
		if len(d.Returns) < minLen {
			minLen = len(d.Returns)
		}
	}

	n := len(data)
	covMatrix := make([][]float64, n)
	for i := range covMatrix {
		covMatrix[i] = make([]float64, n)
	}

	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			if i == j {
				// Variance on diagonal
				variance := calculateVariance(data[i].Returns[:minLen])
				covMatrix[i][j] = variance
			} else {
				// Covariance off-diagonal
				cov, err := CalculateCovariance(
					data[i].Returns[:minLen],
					data[j].Returns[:minLen],
				)
				if err != nil {
					covMatrix[i][j] = 0
				} else {
					covMatrix[i][j] = cov
				}
			}
		}
	}

	return covMatrix, nil
}

// calculateVariance calculates variance of a series
func calculateVariance(data []float64) float64 {
	if len(data) == 0 {
		return 0
	}

	// Calculate mean
	sum := 0.0
	for _, val := range data {
		sum += val
	}
	mean := sum / float64(len(data))

	// Calculate variance
	variance := 0.0
	for _, val := range data {
		diff := val - mean
		variance += diff * diff
	}

	return variance / float64(len(data))
}
