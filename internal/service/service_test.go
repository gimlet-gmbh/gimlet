package service

import (
	"testing"

	"github.com/gmbh-micro/service/static"
)

func TestNewManagedService(t *testing.T) {
	m, err := NewManagedService("101", "testConfig.yaml")
	if err != nil {
		t.Fatalf("failed to parse example config: " + err.Error())
	}
	if m.Mode != Managed {
		t.Fatalf("new managed service did not produce a managed service")
	}
}

func TestNewPlanetaryService(t *testing.T) {
	s, err := static.ParseData("testConfig.yaml")
	if err != nil {
		t.Fatalf("failed to parse example config: " + err.Error())
	}
	m, err := NewPlanetaryService("101", s)
	if m.Mode != Planetary {
		t.Fatalf("new planetary service did not produce a planetary service")
	}

	if m.GetMode() != "planetary" {
		t.Fatalf("failed to return correct mode string")
	}

	m, err = NewPlanetaryService("101", nil)
	if err == nil {
		t.Fatalf("failed to catch nil static data")
	}

}
