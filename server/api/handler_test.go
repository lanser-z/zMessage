package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"zmessage/server/models"
	"zmessage/server/modules/media"
)

// setupTestRouter 设置测试路由
func setupTestRouter() *gin.Engine {
	r := gin.New()
	RegisterMediaRoutes(r, &mockMediaService{})
	return r
}

// mockMediaService 简单的媒体服务模拟
type mockMediaService struct{}

func (m *mockMediaService) Upload(req *media.UploadRequest) (*models.Media, error) {
	return &models.Media{}, nil
}

func (m *mockMediaService) Get(id int64) (*models.Media, error) {
	return &models.Media{}, nil
}

func (m *mockMediaService) GetFile(id int64) (string, string, error) {
	return "/path/to/file", "image/jpeg", nil
}

func (m *mockMediaService) GetThumbnail(id int64) (string, string, error) {
	return "/path/to/thumb", "image/jpeg", nil
}

func (m *mockMediaService) Delete(id int64, userID int64) error {
	return nil
}

func (m *mockMediaService) ValidateType(filename string, mimeType string) (string, error) {
	return "image", nil
}

func (m *mockMediaService) ValidateSize(size int64, mediaType string) error {
	return nil
}

func (m *mockMediaService) GetWithURL(id int64, baseURL string) (*media.MediaWithURL, error) {
	return &media.MediaWithURL{}, nil
}

func TestHandleUpload_NoToken(t *testing.T) {
	r := setupTestRouter()

	req, _ := http.NewRequest("POST", "/api/media/upload", nil)
	req.Header.Set("Authorization", "")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, 401, w.Code)
}
