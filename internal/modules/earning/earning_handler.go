package earning

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"grab/internal/core/httputil"
	"grab/internal/modules/user"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func getUserID(r *http.Request) (uint, bool) {
	uid, ok := r.Context().Value(user.ContextKeyUserID).(uint)
	return uid, ok
}

// AddTrip godoc
// @Summary      Ghi nhận cuốc xe
// @Description  Tài xế nhập cuốc xe mới với tổng thu, phí nền tảng, xăng, chi phí khác. Hệ thống tự tính lãi ròng.
// @Tags         Earning
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      AddTripRequest  true  "Thông tin cuốc xe"
// @Success      201   {object}  Trip
// @Failure      400   {object}  map[string]string
// @Router       /earning/trips [post]
func (h *Handler) AddTrip(w http.ResponseWriter, r *http.Request) {
	uid, ok := getUserID(r)
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	httputil.LimitBody(r)
	var req AddTripRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	trip, err := h.service.AddTrip(r.Context(), uid, req)
	if err != nil {
		msg := err.Error()
		if strings.Contains(msg, "must be") || strings.Contains(msg, "format") {
			httputil.WriteError(w, http.StatusBadRequest, msg)
		} else {
			httputil.WriteError(w, http.StatusInternalServerError, msg)
		}
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, trip)
}

// ListTrips godoc
// @Summary      Danh sách cuốc xe
// @Description  Lấy danh sách cuốc xe, lọc theo ngày nếu cần
// @Tags         Earning
// @Produce      json
// @Security     BearerAuth
// @Param        date    query  string  false  "Lọc theo ngày YYYY-MM-DD"
// @Param        limit   query  int     false  "Số kết quả (mặc định 50)"
// @Param        offset  query  int     false  "Offset"
// @Success      200     {array}  Trip
// @Router       /earning/trips [get]
func (h *Handler) ListTrips(w http.ResponseWriter, r *http.Request) {
	uid, ok := getUserID(r)
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))
	limit, offset = httputil.ClampPagination(limit, offset, 50, 200)

	trips, err := h.service.ListTrips(r.Context(), uid, q.Get("date"), limit, offset)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if trips == nil {
		trips = []Trip{}
	}
	httputil.WriteJSON(w, http.StatusOK, trips)
}

// DailySummary godoc
// @Summary      Tổng kết ngày
// @Description  Tổng thu, tổng chi, lãi ròng trong một ngày
// @Tags         Earning
// @Produce      json
// @Security     BearerAuth
// @Param        date  query    string  false  "Ngày YYYY-MM-DD (mặc định hôm nay)"
// @Success      200   {object} DailySummary
// @Router       /earning/summary/daily [get]
func (h *Handler) DailySummary(w http.ResponseWriter, r *http.Request) {
	uid, ok := getUserID(r)
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	summary, err := h.service.DailySummary(r.Context(), uid, r.URL.Query().Get("date"))
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.WriteJSON(w, http.StatusOK, summary)
}

// MonthlySummary godoc
// @Summary      Tổng kết tháng
// @Description  Tổng thu, tổng chi, lãi ròng, trung bình lãi/ngày trong tháng
// @Tags         Earning
// @Produce      json
// @Security     BearerAuth
// @Param        month  query    string  false  "Tháng YYYY-MM (mặc định tháng này)"
// @Success      200    {object} MonthlySummary
// @Router       /earning/summary/monthly [get]
func (h *Handler) MonthlySummary(w http.ResponseWriter, r *http.Request) {
	uid, ok := getUserID(r)
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	summary, err := h.service.MonthlySummary(r.Context(), uid, r.URL.Query().Get("month"))
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.WriteJSON(w, http.StatusOK, summary)
}
