// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0
package main

import (
	"math"
	"math/rand"
	"runtime"
)

// startCPULoad saturates one OS thread from startup with representative checkout
// pricing stack frames for eBPF profiler testing.
func startCPULoad() {
	go func() {
		runtime.LockOSThread()
		checkoutPricingWorker()
	}()
}

func checkoutPricingWorker() {
	rng := rand.New(rand.NewSource(42))
	for {
		processOrderBatch(rng, 32)
	}
}

func processOrderBatch(rng *rand.Rand, batchSize int) float64 {
	var total float64
	for i := 0; i < batchSize; i++ {
		total += computeOrderTotal(rng, i)
	}
	return total
}

func computeOrderTotal(rng *rand.Rand, orderID int) float64 {
	items := generateOrderItems(rng, orderID, 8)
	subtotal := sumItemPrices(items)
	tax := computeTax(subtotal, "US-CA")
	shipping := estimateShippingCost(rng, orderID, len(items))
	discount := applyPromotionalDiscount(rng, subtotal)
	return finalizeTotal(subtotal, tax, shipping, discount)
}

func generateOrderItems(rng *rand.Rand, orderID int, count int) []float64 {
	items := make([]float64, count)
	for i := range items {
		items[i] = computeItemPrice(rng, orderID*1000+i)
	}
	return items
}

func computeItemPrice(rng *rand.Rand, productID int) float64 {
	basePrice := 1.0 + rng.Float64()*999.0
	demandFactor := computeDemandFactor(rng, productID)
	seasonalFactor := computeSeasonalFactor(productID)
	return roundToNearestCent(basePrice * demandFactor * seasonalFactor)
}

func computeDemandFactor(rng *rand.Rand, productID int) float64 {
	sales := make([]float64, 30)
	for i := range sales {
		sales[i] = rng.Float64() * 100
	}
	return 0.8 + 0.4*normalizedVariance(sales)
}

func computeSeasonalFactor(productID int) float64 {
	month := (productID % 12) + 1
	return 0.9 + 0.2*math.Sin(float64(month)*math.Pi/6.0)
}

func normalizedVariance(data []float64) float64 {
	if len(data) == 0 {
		return 0
	}
	mean := sumSlice(data) / float64(len(data))
	var variance float64
	for _, v := range data {
		diff := v - mean
		variance += diff * diff
	}
	variance /= float64(len(data))
	return math.Sqrt(variance) / (mean + 1e-9)
}

func sumSlice(data []float64) float64 {
	var sum float64
	for _, v := range data {
		sum += v
	}
	return sum
}

func sumItemPrices(items []float64) float64 {
	return sumSlice(items)
}

func computeTax(subtotal float64, region string) float64 {
	rate := lookupTaxRate(region)
	return roundToNearestCent(subtotal * rate)
}

func lookupTaxRate(region string) float64 {
	rates := map[string]float64{
		"US-CA": 0.0725, "US-NY": 0.08, "US-TX": 0.0625,
		"US-WA": 0.065, "US-FL": 0.06, "EU-DE": 0.19,
		"EU-FR": 0.20, "EU-UK": 0.20,
	}
	if rate, ok := rates[region]; ok {
		return rate
	}
	return 0.05
}

func estimateShippingCost(rng *rand.Rand, orderID int, itemCount int) float64 {
	baseRate := 4.99
	weightFactor := computePackageWeight(rng, itemCount)
	distanceFactor := computeDistanceFactor(rng, orderID)
	return roundToNearestCent(baseRate + weightFactor*distanceFactor)
}

func computePackageWeight(rng *rand.Rand, itemCount int) float64 {
	var total float64
	for i := 0; i < itemCount; i++ {
		total += 0.1 + rng.Float64()*4.9
	}
	return total
}

func computeDistanceFactor(rng *rand.Rand, orderID int) float64 {
	latA := -90.0 + rng.Float64()*180.0
	lonA := -180.0 + rng.Float64()*360.0
	latB := -90.0 + rng.Float64()*180.0
	lonB := -180.0 + rng.Float64()*360.0
	return haversineDistance(latA, lonA, latB, lonB) / 1000.0
}

func haversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371.0
	dLat := (lat2 - lat1) * math.Pi / 180.0
	dLon := (lon2 - lon1) * math.Pi / 180.0
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180.0)*math.Cos(lat2*math.Pi/180.0)*
			math.Sin(dLon/2)*math.Sin(dLon/2)
	return R * 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
}

func applyPromotionalDiscount(rng *rand.Rand, subtotal float64) float64 {
	tier := classifyCustomerTier(rng)
	rate := tierDiscountRate(tier)
	return roundToNearestCent(subtotal * rate)
}

func classifyCustomerTier(rng *rand.Rand) int {
	score := computeLoyaltyScore(rng)
	switch {
	case score > 0.9:
		return 3
	case score > 0.7:
		return 2
	case score > 0.4:
		return 1
	default:
		return 0
	}
}

func computeLoyaltyScore(rng *rand.Rand) float64 {
	orderCount := rng.Intn(200)
	totalSpend := rng.Float64() * 10000
	recency := rng.Float64()
	return math.Min(1.0,
		(math.Log(float64(orderCount+1))/math.Log(200))*0.4+
			(math.Log(totalSpend+1)/math.Log(10001))*0.4+
			recency*0.2)
}

func tierDiscountRate(tier int) float64 {
	rates := []float64{0.0, 0.05, 0.10, 0.15}
	if tier < len(rates) {
		return rates[tier]
	}
	return 0
}

func finalizeTotal(subtotal, tax, shipping, discount float64) float64 {
	return roundToNearestCent(subtotal + tax + shipping - discount)
}

func roundToNearestCent(amount float64) float64 {
	return math.Round(amount*100) / 100
}
