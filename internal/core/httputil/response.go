package httputil

import (
	"encoding/json"
	"log"
	"math"
	"net/http"
)

const MaxBodySize = 1 << 20 // 1MB

// WriteJSON ghi JSON response. Log lỗi encoding thay vì bỏ qua.
func WriteJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("[http] JSON encode error: %v", err)
	}
}

// WriteError ghi JSON error response.
func WriteError(w http.ResponseWriter, status int, msg string) {
	WriteJSON(w, status, map[string]string{"error": msg})
}

// LimitBody wrap r.Body bằng MaxBytesReader để chống DoS.
func LimitBody(r *http.Request) {
	r.Body = http.MaxBytesReader(nil, r.Body, MaxBodySize)
}

// ValidCoords kiểm tra tọa độ hợp lệ (bao gồm NaN, Inf).
func ValidCoords(lat, lng float64) bool {
	if math.IsNaN(lat) || math.IsNaN(lng) || math.IsInf(lat, 0) || math.IsInf(lng, 0) {
		return false
	}
	return lat >= -90 && lat <= 90 && lng >= -180 && lng <= 180
}

// ClampPagination đảm bảo limit/offset không âm và có giới hạn hợp lý.
func ClampPagination(limit, offset, defaultLimit, maxLimit int) (int, int) {
	if limit <= 0 {
		limit = defaultLimit
	} else if limit > maxLimit {
		limit = maxLimit
	}
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}
