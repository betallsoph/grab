package mood

import (
	"encoding/json"
	"net/http"
	"strconv"

	"grab/internal/modules/user"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data) //nolint:errcheck
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// CreatePost godoc
// @Summary      Đăng bài xả mood
// @Description  Tài xế viết tâm trạng lên bảng tin, có thể ẩn danh
// @Tags         Mood
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      CreatePostRequest  true  "Nội dung bài viết"
// @Success      201   {object}  Post
// @Failure      400   {object}  map[string]string
// @Router       /mood/posts [post]
func (h *Handler) CreatePost(w http.ResponseWriter, r *http.Request) {
	uid, ok := r.Context().Value(user.ContextKeyUserID).(uint)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req CreatePostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	post, err := h.service.CreatePost(r.Context(), uid, req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, post)
}

// ListPosts godoc
// @Summary      Feed bài viết
// @Description  Lấy danh sách bài viết xả mood, mới nhất trước
// @Tags         Mood
// @Produce      json
// @Security     BearerAuth
// @Param        limit   query  int  false  "Số bài (mặc định 20)"
// @Param        offset  query  int  false  "Offset phân trang"
// @Success      200     {array}  Post
// @Router       /mood/posts [get]
func (h *Handler) ListPosts(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))

	posts, err := h.service.ListPosts(r.Context(), limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if posts == nil {
		posts = []Post{}
	}
	writeJSON(w, http.StatusOK, posts)
}

// LikePost godoc
// @Summary      Like bài viết
// @Description  Thả like cho bài viết
// @Tags         Mood
// @Produce      json
// @Security     BearerAuth
// @Param        postId  path      string  true  "Post ID"
// @Success      200     {object}  map[string]string
// @Failure      400     {object}  map[string]string
// @Router       /mood/posts/{postId}/like [post]
func (h *Handler) LikePost(w http.ResponseWriter, r *http.Request) {
	postID := r.PathValue("postId")

	if err := h.service.LikePost(r.Context(), postID); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "liked"})
}

// CreateComment godoc
// @Summary      Bình luận bài viết
// @Description  Thêm comment vào bài viết, có thể ẩn danh
// @Tags         Mood
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        postId  path      string                true  "Post ID"
// @Param        body    body      CreateCommentRequest  true  "Nội dung comment"
// @Success      201     {object}  Comment
// @Failure      400     {object}  map[string]string
// @Router       /mood/posts/{postId}/comments [post]
func (h *Handler) CreateComment(w http.ResponseWriter, r *http.Request) {
	uid, ok := r.Context().Value(user.ContextKeyUserID).(uint)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	postID := r.PathValue("postId")

	var req CreateCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	comment, err := h.service.CreateComment(r.Context(), uid, postID, req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, comment)
}

// ListComments godoc
// @Summary      Danh sách comment
// @Description  Lấy comment của bài viết, cũ nhất trước
// @Tags         Mood
// @Produce      json
// @Security     BearerAuth
// @Param        postId  path   string  true   "Post ID"
// @Param        limit   query  int     false  "Số comment (mặc định 50)"
// @Param        offset  query  int     false  "Offset phân trang"
// @Success      200     {array}  Comment
// @Router       /mood/posts/{postId}/comments [get]
func (h *Handler) ListComments(w http.ResponseWriter, r *http.Request) {
	postID := r.PathValue("postId")
	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))

	comments, err := h.service.ListComments(r.Context(), postID, limit, offset)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if comments == nil {
		comments = []Comment{}
	}
	writeJSON(w, http.StatusOK, comments)
}
