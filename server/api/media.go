package api

import (
	"mime/multipart"
	"strconv"

	"github.com/gin-gonic/gin"
	"zmessage/server/models"
	"zmessage/server/modules/media"
)

// RegisterMediaRoutes 注册媒体路由
func RegisterMediaRoutes(r *gin.Engine, svc media.Service) {
	media := r.Group("/api/media")
	// 简单的token验证，实际应该注入user.Service
	media.Use(func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			c.JSON(401, ErrorResponse{Error: "USER_UNAUTHORIZED"})
			c.Abort()
			return
		}
		// TODO: 实际项目中应该调用user.Service.ValidateToken
		c.Set("auth", &AuthContext{UserID: 1}) // 临时使用固定用户ID
		c.Next()
	})
	{
		media.POST("/upload", handleUpload(svc))
		media.GET("/:id", handleGetMedia(svc))
		media.GET("/:id/thumb", handleGetThumbnail(svc))
		media.DELETE("/:id", handleDeleteMedia(svc))
	}
}

// handleUpload 处理文件上传
func handleUpload(svc media.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := GetAuthContext(c)
		if auth == nil {
			return
		}

		// 显式使用导入的类型
		var _ multipart.FileHeader
		var _ models.Media

		// 获取文件
		fileHeader, err := c.FormFile("file")
		if err != nil {
			BadRequest(c, "未找到文件")
			return
		}

		// 打开文件
		file, err := fileHeader.Open()
		if err != nil {
			InternalError(c, err)
			return
		}
		defer file.Close()

		// 获取文件类型
		fileType := c.PostForm("type")
		if fileType == "" {
			BadRequest(c, "未指定文件类型")
			return
		}

		// 验证文件类型
		mediaType, err := svc.ValidateType(fileHeader.Filename, fileHeader.Header.Get("Content-Type"))
		if err != nil {
			c.JSON(400, ErrorResponse{Error: err.Error()})
			return
		}

		// 获取文件大小（从FileHeader）
		size := fileHeader.Size

		// 验证文件大小
		if err := svc.ValidateSize(size, mediaType); err != nil {
			c.JSON(413, ErrorResponse{Error: err.Error()})
			return
		}

		// 调用媒体服务上传
		media, err := svc.Upload(&media.UploadRequest{
			File:     file,
			Header:   fileHeader,
			Type:     mediaType,
			OwnerID:  auth.UserID,
		})
		if err != nil {
			handleMediaError(c, err)
			return
		}

		c.JSON(200, MediaResponse{
			ID:         media.ID,
			Type:       media.Type,
			OriginalURL: "/api/media/" + strconv.FormatInt(media.ID, 10),
			ThumbnailURL: "/api/media/" + strconv.FormatInt(media.ID, 10) + "/thumb",
			Size:        media.Size,
			MimeType:    media.MimeType,
			Width:       getIntOrZero(media.Width),
			Height:      getIntOrZero(media.Height),
			CreatedAt:    media.CreatedAt,
		})
	}
}

// handleGetMedia 处理获取媒体文件
func handleGetMedia(svc media.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取媒体ID
		idStr := c.Param("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			BadRequest(c, "无效的媒体ID")
			return
		}

		// 获取文件路径
		path, _, err := svc.GetFile(id)
		if err != nil {
			handleMediaError(c, err)
			return
		}

		// 设置响应头并发送文件
		c.File(path)
	}
}

// handleGetThumbnail 处理获取缩略图
func handleGetThumbnail(svc media.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取媒体ID
		idStr := c.Param("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			BadRequest(c, "无效的媒体ID")
			return
		}

		// 获取缩略图路径
		path, _, err := svc.GetThumbnail(id)
		if err != nil {
			handleMediaError(c, err)
			return
		}

		// 设置响应头并发送文件
		c.File(path)
	}
}

// handleDeleteMedia 处理删除媒体文件
func handleDeleteMedia(svc media.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := GetAuthContext(c)
		if auth == nil {
			return
		}

		// 获取媒体ID
		idStr := c.Param("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			BadRequest(c, "无效的媒体ID")
			return
		}

		// 调用媒体服务删除
		err = svc.Delete(id, auth.UserID)
		if err != nil {
			handleMediaError(c, err)
			return
		}

		Success(c, map[string]bool{"success": true})
	}
}

// handleMediaError 处理媒体服务错误
func handleMediaError(c *gin.Context, err error) {
	switch err.Error() {
	case "MEDIA_INVALID_TYPE":
		c.JSON(400, ErrorResponse{Error: err.Error()})
	case "MEDIA_INVALID_SIZE":
		c.JSON(413, ErrorResponse{Error: err.Error()})
	case "MEDIA_NOT_FOUND":
		NotFound(c, "文件不存在")
	default:
		InternalError(c, err)
	}
}

// MediaResponse 媒体响应
type MediaResponse struct {
	ID           int64  `json:"id"`
	Type         string `json:"type"`
	OriginalURL  string `json:"original_url"`
	ThumbnailURL  string `json:"thumbnail_url,omitempty"`
	Size         int64  `json:"size"`
	MimeType     string `json:"mime_type"`
	Width        int    `json:"width,omitempty"`
	Height       int    `json:"height,omitempty"`
	CreatedAt    int64  `json:"created_at"`
}

// getIntOrZero 安全获取int值
func getIntOrZero(p *int) int {
	if p == nil {
		return 0
	}
	return *p
}
