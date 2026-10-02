package utils

import (
	"crypto/rand"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func GetEnvOrDefault(key, fallback string) string { return GetEnv(key, fallback) }

func GenerateTrackingCode() string {
	b := make([]byte, 4)
	rand.Read(b)
	return fmt.Sprintf("RYX-%X-%d", b, time.Now().Unix()%10000)
}

func GeneratePaymentRef() string {
	b := make([]byte, 6)
	rand.Read(b)
	return fmt.Sprintf("PAY-%X", b)
}

func HaversineKm(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371.0
	dLat := (lat2 - lat1) * math.Pi / 180
	dLon := (lon2 - lon1) * math.Pi / 180
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*
			math.Sin(dLon/2)*math.Sin(dLon/2)
	return R * 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
}

func DeliveryFee(distanceKm, weightKg float64) float64 {
	fee := 500.0 + distanceKm*80.0 + weightKg*50.0
	return math.Round(fee/10) * 10
}

func AllowMethods(w http.ResponseWriter, r *http.Request, methods ...string) bool {
	for _, m := range methods {
		if r.Method == m {
			return true
		}
	}
	w.Header().Set("Allow", strings.Join(methods, ", "))
	w.WriteHeader(http.StatusMethodNotAllowed)
	return false
}

func ParseUint(s string, dst *uint) (uint, error) {
	v, err := strconv.ParseUint(strings.TrimSpace(s), 10, 64)
	if err != nil {
		return 0, err
	}
	if dst != nil {
		*dst = uint(v)
	}
	return uint(v), nil
}

func PathID(path, prefix string) (uint, error) {
	raw := strings.TrimPrefix(path, prefix)
	raw = strings.Trim(raw, "/")
	return ParseUint(raw, nil)
}

func ParseUintPath(path, suffix string) uint {
	parts := strings.Split(path, "/")
	for i := len(parts) - 1; i >= 0; i-- {
		if parts[i] == suffix && i > 0 {
			v, _ := strconv.ParseUint(parts[i-1], 10, 64)
			return uint(v)
		}
	}
	return 0
}
