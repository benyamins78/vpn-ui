package service

import "testing"

func TestBBRHelpersUseWholeWordsAndRejectUnsafeRestoreValues(t *testing.T) {
	if !containsWord("cubic bbr", "bbr") || containsWord("notbbr", "bbr") {
		t.Fatal("BBR availability matching is not word-bounded")
	}
	for _, value := range []string{"cubic", "fq_codel", "bbr-legacy"} {
		if !safeBBRValue(value) {
			t.Fatalf("safe value %q rejected", value)
		}
	}
	for _, value := range []string{"cubic; reboot", "$(id)", "fq/codel", ""} {
		if safeBBRValue(value) {
			t.Fatalf("unsafe value %q accepted", value)
		}
	}
}
