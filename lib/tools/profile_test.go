package tools

import (
	"strings"
	"testing"

	"github.com/hsk-kr/licokit/lib/config"
)

func TestValidProfile(t *testing.T) {
	for _, profile := range []string{"core", "personal", "all"} {
		if !ValidProfile(profile) {
			t.Errorf("expected %q to be valid", profile)
		}
	}
	if ValidProfile("unknown") {
		t.Fatal("unknown profile should be invalid")
	}
}

func TestInstallDescription(t *testing.T) {
	tests := []struct {
		tool config.ToolConfig
		want string
	}{
		{config.ToolConfig{InstallType: "brew", Package: "git"}, "brew install git"},
		{config.ToolConfig{InstallType: "cask", Package: "ghostty"}, "brew install --cask ghostty"},
		{config.ToolConfig{InstallType: "tap", Package: "owner/tap"}, "brew tap owner/tap"},
		{config.ToolConfig{InstallType: "mas", Package: "123"}, "mas install 123"},
	}
	for _, test := range tests {
		if got := installDescription(test.tool); got != test.want {
			t.Errorf("installDescription(%s) = %q, want %q", test.tool.InstallType, got, test.want)
		}
	}
}

func TestCombineFailures(t *testing.T) {
	if combineFailures(nil, "") != nil {
		t.Fatal("no failures should return nil")
	}
	err := combineFailures([]string{"one", "two"}, "")
	if err == nil || !strings.Contains(err.Error(), "2 step(s)") {
		t.Fatalf("unexpected combined error: %v", err)
	}
}
