package net

import (
	"testing"
)

func TestGetSysctl(t *testing.T) {
	got, err := GetSysctl("net.ipv4.ip_forward")
	ex := "1"
	if err != nil {
		t.Fatalf("Could not read key")
	}
	if got != ex {
		t.Errorf("Expected: %s, got %s", ex, got)
	}
}

func TestSetSysctl(t *testing.T) {
	ex := "0"
	err := SetSysctl("net.ipv4.ip_forward", ex)
	if err != nil {
		t.Fatalf("err")
	}
	got, err := GetSysctl("net.ipv4.ip_forward")
	if err != nil {
		t.Fatalf("Could not read key")
	}
	if got != ex {
		t.Errorf("Expected: %s, got %s", ex, got)
	}
}
