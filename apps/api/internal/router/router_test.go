package router

import (
	"testing"

	"github.com/mahadillahm4di-cyber/mh-gdpr-ai.eu-s-plus/internal/proxy"
)

func TestSmartRouter_SelectCheapest(t *testing.T) {
	r := NewSmartRouter([]proxy.Provider{
		proxy.ProviderOpenAI,
		proxy.ProviderAnthropic,
		proxy.ProviderOllama,
	})

	got := r.Select(StrategyCheapest)
	if got != proxy.ProviderOllama {
		t.Errorf("cheapest should be ollama (free), got %s", got)
	}
}

func TestSmartRouter_SelectFastest(t *testing.T) {
	r := NewSmartRouter([]proxy.Provider{
		proxy.ProviderOpenAI,
		proxy.ProviderOllama,
	})

	got := r.Select(StrategyFastest)
	if got != proxy.ProviderOpenAI {
		t.Errorf("fastest should be openai, got %s", got)
	}
}

func TestSmartRouter_SelectBest(t *testing.T) {
	r := NewSmartRouter([]proxy.Provider{
		proxy.ProviderOllama,
		proxy.ProviderAnthropic,
		proxy.ProviderOpenAI,
	})

	got := r.Select(StrategyBest)
	if got != proxy.ProviderOpenAI {
		t.Errorf("best should be openai, got %s", got)
	}
}

func TestSmartRouter_SelectBest_NoOpenAI(t *testing.T) {
	r := NewSmartRouter([]proxy.Provider{
		proxy.ProviderOllama,
		proxy.ProviderAnthropic,
	})

	got := r.Select(StrategyBest)
	if got != proxy.ProviderAnthropic {
		t.Errorf("best without openai should be anthropic, got %s", got)
	}
}

func TestSmartRouter_EmptyProviders(t *testing.T) {
	r := NewSmartRouter([]proxy.Provider{})

	got := r.Select(StrategyCheapest)
	if got != proxy.ProviderOpenAI {
		t.Errorf("empty should fallback to openai, got %s", got)
	}
}

func TestSmartRouter_SingleProvider(t *testing.T) {
	r := NewSmartRouter([]proxy.Provider{proxy.ProviderAnthropic})

	got := r.Select(StrategyCheapest)
	if got != proxy.ProviderAnthropic {
		t.Errorf("single provider should return that provider, got %s", got)
	}
}

func TestSmartRouter_UnknownStrategy(t *testing.T) {
	r := NewSmartRouter([]proxy.Provider{
		proxy.ProviderOllama,
		proxy.ProviderOpenAI,
	})

	got := r.Select(Strategy("unknown"))
	if got != proxy.ProviderOllama {
		t.Errorf("unknown strategy should return first available, got %s", got)
	}
}

func TestStrategyConstants(t *testing.T) {
	if StrategyCheapest != "cheapest" {
		t.Error("unexpected cheapest value")
	}
	if StrategyFastest != "fastest" {
		t.Error("unexpected fastest value")
	}
	if StrategyBest != "best" {
		t.Error("unexpected best value")
	}
}
