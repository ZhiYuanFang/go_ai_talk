package controller

import "testing"

func TestShouldProxyHistoryRequest(t *testing.T) {
	cfgLocal := historyRouteProxyConfig{mode: historyRouteModeLocal, targetURL: "http://history:9801"}
	if shouldProxyHistoryRequest(cfgLocal, "dev-1") {
		t.Fatal("local mode should not proxy")
	}

	cfgProxy := historyRouteProxyConfig{mode: historyRouteModeProxy, targetURL: "http://history:9801"}
	if !shouldProxyHistoryRequest(cfgProxy, "dev-1") {
		t.Fatal("proxy mode should always proxy")
	}

	cfgCanary0 := historyRouteProxyConfig{mode: historyRouteModeCanary, targetURL: "http://history:9801", canaryPercent: 0}
	if shouldProxyHistoryRequest(cfgCanary0, "dev-1") {
		t.Fatal("canary 0 should not proxy")
	}

	cfgCanary100 := historyRouteProxyConfig{mode: historyRouteModeCanary, targetURL: "http://history:9801", canaryPercent: 100}
	if !shouldProxyHistoryRequest(cfgCanary100, "dev-1") {
		t.Fatal("canary 100 should proxy")
	}

	cfgNoTarget := historyRouteProxyConfig{mode: historyRouteModeProxy, targetURL: ""}
	if shouldProxyHistoryRequest(cfgNoTarget, "dev-1") {
		t.Fatal("empty target should not proxy")
	}
}
