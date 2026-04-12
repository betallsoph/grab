package sos

// --- WebSocket message types ---

const (
	MsgTypeSOS       = "sos"        // Tài xế gửi lên: bật SOS
	MsgTypeSOSCancel = "sos_cancel" // Tài xế gửi lên: hủy SOS
	MsgTypeSOSAlert  = "sos_alert"  // Server broadcast xuống các tài xế gần
)

// IncomingMessage là cấu trúc JSON mà mobile gửi lên qua WebSocket.
//
//	{ "type": "sos" }
//	{ "type": "sos_cancel" }
type IncomingMessage struct {
	Type string `json:"type"`
}

// SOSAlert là payload server broadcast xuống cho các tài xế lân cận.
type SOSAlert struct {
	Type       string  `json:"type"`
	FromDriver uint    `json:"from_driver"`
	FullName   string  `json:"full_name"`
	Phone      string  `json:"phone"`
	Latitude   float64 `json:"latitude"`
	Longitude  float64 `json:"longitude"`
	DistanceKm float64 `json:"distance_km"` // khoảng cách từ người nhận đến người gửi
	Timestamp  int64   `json:"timestamp"`
}

// NearbyDriver là kết quả trả về từ Redis GEOSEARCH.
type NearbyDriver struct {
	UserID     uint
	Latitude   float64
	Longitude  float64
	DistanceKm float64
}
