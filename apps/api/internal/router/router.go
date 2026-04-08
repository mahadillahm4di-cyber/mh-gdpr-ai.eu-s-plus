// Package router provides intelligent provider selection based on cost, speed, and capabilities.
package router

import (
	"github.com/mahadillahm4di-cyber/mh-gdpr-ai.eu-s-plus/internal/proxy"
)

// Strategy defines how the router selects a provider.
type Strategy string

const (
	StrategyCheapest Strategy = "cheapest"
	StrategyFastest  Strategy = "fastest"
	StrategyBest     Strategy = "best"
)

// ProviderCost estimates cost per 1K tokens for each provider.
var ProviderCost = map[proxy.Provider]float64{
	proxy.ProviderOllama:    0.0,    // Free (local)
	proxy.ProviderOpenAI:    0.005,  // ~$5/1M tokens for GPT-4o
	proxy.ProviderAnthropic: 0.003,  // ~$3/1M tokens for Claude Sonnet
}

// ProviderSpeed estimates relative speed (lower = faster).
var ProviderSpeed = map[proxy.Provider]int{
	proxy.ProviderOpenAI:    2,
	proxy.ProviderAnthropic: 2,
	proxy.ProviderOllama:    5, // Local depends on hardware
}

// SmartRouter selects the best provider based on strategy and availability.
type SmartRouter struct {
	available []proxy.Provider
}

// NewSmartRouter creates a router with the given available providers.
func NewSmartRouter(available []proxy.Provider) *SmartRouter {
	return &SmartRouter{available: available}
}

// Select picks the best provider based on strategy.
func (r *SmartRouter) Select(strategy Strategy) proxy.Provider {
	if len(r.available) == 0 {
		return proxy.ProviderOpenAI // fallback
	}

	switch strategy {
	case StrategyCheapest:
		return r.selectCheapest()
	case StrategyFastest:
		return r.selectFastest()
	case StrategyBest:
		return r.selectBest()
	default:
		return r.available[0]
	}
}

func (r *SmartRouter) selectCheapest() proxy.Provider {
	best := r.available[0]
	bestCost := ProviderCost[best]
	for _, p := range r.available[1:] {
		if cost, ok := ProviderCost[p]; ok && cost < bestCost {
			best = p
			bestCost = cost
		}
	}
	return best
}

func (r *SmartRouter) selectFastest() proxy.Provider {
	best := r.available[0]
	bestSpeed := ProviderSpeed[best]
	for _, p := range r.available[1:] {
		if speed, ok := ProviderSpeed[p]; ok && speed < bestSpeed {
			best = p
			bestSpeed = speed
		}
	}
	return best
}

func (r *SmartRouter) selectBest() proxy.Provider {
	// For MVP, "best" = OpenAI if available, then Anthropic, then Ollama
	priority := []proxy.Provider{proxy.ProviderOpenAI, proxy.ProviderAnthropic, proxy.ProviderOllama}
	availSet := make(map[proxy.Provider]bool)
	for _, p := range r.available {
		availSet[p] = true
	}
	for _, p := range priority {
		if availSet[p] {
			return p
		}
	}
	return r.available[0]
}
