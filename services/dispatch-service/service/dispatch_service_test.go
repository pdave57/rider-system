package service

import (
	"math"
	"testing"
)

func TestHaversine(t *testing.T) {
	// Test identical coordinates
	distSame := haversine(6.5244, 3.3792, 6.5244, 3.3792)
	if distSame != 0 {
		t.Errorf("expected distance 0 for same coords, got %f", distSame)
	}

	// Test point ~0.55km away (1km radius test)
	dist05 := haversine(6.5244, 3.3792, 6.5294, 3.3792)
	if dist05 < 0.5 || dist05 > 0.6 {
		t.Errorf("expected distance ~0.55km, got %f", dist05)
	}
	if dist05 > 1.0 {
		t.Errorf("distance %f should be within 1km radius", dist05)
	}

	// Test point ~1.66km away (2km radius test)
	dist15 := haversine(6.5244, 3.3792, 6.5394, 3.3792)
	if dist15 < 1.5 || dist15 > 1.8 {
		t.Errorf("expected distance ~1.66km, got %f", dist15)
	}
	if dist15 <= 1.0 || dist15 > 2.0 {
		t.Errorf("distance %f should be between 1km and 2km radius", dist15)
	}

	// Test point ~3.33km away (>2km radius test)
	dist30 := haversine(6.5244, 3.3792, 6.5544, 3.3792)
	if dist30 <= 2.0 {
		t.Errorf("distance %f should be outside 2km radius", dist30)
	}
}

func TestRadiusBounds(t *testing.T) {
	// Verify latitude offset calculation approximation for 1km and 2km
	// 1 degree of latitude is ~111km
	latOffset1km := 1.0 / 111.0
	latOffset2km := 2.0 / 111.0

	dist1km := haversine(6.5244, 3.3792, 6.5244+latOffset1km, 3.3792)
	if math.Abs(dist1km-1.0) > 0.05 {
		t.Errorf("expected ~1.0km, got %f", dist1km)
	}

	dist2km := haversine(6.5244, 3.3792, 6.5244+latOffset2km, 3.3792)
	if math.Abs(dist2km-2.0) > 0.05 {
		t.Errorf("expected ~2.0km, got %f", dist2km)
	}
}
