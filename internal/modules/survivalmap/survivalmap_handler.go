package survivalmap

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

// --- Handlers ---

// CreatePOI godoc
// @Summary      Tạo POI mới
// @Description  Tài xế đóng góp một điểm POI (quán ăn, sửa xe, cà phê...) lên bản đồ sinh tồn
// @Tags         Map
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      CreatePOIRequest  true  "Thông tin POI"
// @Success      201   {object}  POI
// @Failure      400   {object}  map[string]string
// @Failure      401   {object}  map[string]string
// @Router       /map/pois [post]
func (h *Handler) CreatePOI(w http.ResponseWriter, r *http.Request) {
	httputil.LimitBody(r)
	userID, ok := r.Context().Value(user.ContextKeyUserID).(uint)
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req CreatePOIRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	poi, err := h.service.CreatePOI(r.Context(), userID, req)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	httputil.WriteJSON(w, http.StatusCreated, poi)
}

// SearchNearby godoc
// @Summary      Tìm POI xung quanh
// @Description  Truy vấn các điểm POI trong bán kính xung quanh vị trí tài xế (PostGIS)
// @Tags         Map
// @Produce      json
// @Security     BearerAuth
// @Param        lat        query    number  true   "Vĩ độ tài xế"
// @Param        lng        query    number  true   "Kinh độ tài xế"
// @Param        radius_km  query    number  false  "Bán kính km (mặc định 3)"
// @Param        category   query    string  false  "Lọc loại: food, repair, cafe, wc, gas, rest, other"
// @Param        limit      query    int     false  "Số kết quả (mặc định 50)"
// @Param        offset     query    int     false  "Offset phân trang"
// @Success      200        {array}  POIWithDistance
// @Failure      400        {object} map[string]string
// @Router       /map/pois/nearby [get]
func (h *Handler) SearchNearby(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	lat, err := strconv.ParseFloat(q.Get("lat"), 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "lat is required")
		return
	}
	lng, err := strconv.ParseFloat(q.Get("lng"), 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "lng is required")
		return
	}
	if !httputil.ValidCoords(lat, lng) {
		httputil.WriteError(w, http.StatusBadRequest, "invalid coordinates")
		return
	}

	radiusKm, _ := strconv.ParseFloat(q.Get("radius_km"), 64)
	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))
	limit, offset = httputil.ClampPagination(limit, offset, 50, 100)

	results, err := h.service.SearchNearby(r.Context(), NearbyQuery{
		Latitude:  lat,
		Longitude: lng,
		RadiusKm:  radiusKm,
		Category:  Category(q.Get("category")),
		Limit:     limit,
		Offset:    offset,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	httputil.WriteJSON(w, http.StatusOK, results)
}

// GetPOI godoc
// @Summary      Xem chi tiết POI
// @Description  Lấy thông tin chi tiết một điểm POI theo ID
// @Tags         Map
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "POI ID"
// @Success      200  {object}  POI
// @Failure      404  {object}  map[string]string
// @Router       /map/pois/{id} [get]
func (h *Handler) GetPOI(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid poi id")
		return
	}

	poi, err := h.service.GetPOI(r.Context(), uint(id))
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			httputil.WriteError(w, http.StatusNotFound, "poi not found")
		} else {
			httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	httputil.WriteJSON(w, http.StatusOK, poi)
}

// UpvotePOI godoc
// @Summary      Upvote POI
// @Description  Tài xế upvote một POI hữu ích
// @Tags         Map
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "POI ID"
// @Success      200  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /map/pois/{id}/upvote [post]
func (h *Handler) UpvotePOI(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid poi id")
		return
	}

	if err := h.service.UpvotePOI(r.Context(), uint(id)); err != nil {
		httputil.WriteError(w, http.StatusNotFound, err.Error())
		return
	}

	httputil.WriteJSON(w, http.StatusOK, map[string]string{"message": "upvoted"})
}
