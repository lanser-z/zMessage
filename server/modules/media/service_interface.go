package media

import (
	"io"
	"mime/multipart"
	"zmessage/server/models"
)

// UploadRequest 上传请求
type UploadRequest struct {
	File     multipart.File
	Header   *multipart.FileHeader
	Type      string // "image" or "voice"
	OwnerID   int64
}

// MediaWithURL 带URL的媒体信息
type MediaWithURL struct {
	*models.Media
	OriginalURL  string `json:"original_url"`
	ThumbnailURL string `json:"thumbnail_url,omitempty"`
}

// Service 媒体服务接口
type Service interface {
	// Upload 上传媒体文件
	Upload(req *UploadRequest) (*models.Media, error)

	// Get 获取媒体信息
	Get(id int64) (*models.Media, error)

	// GetWithURL 获取带URL的媒体信息
	GetWithURL(id int64, baseURL string) (*MediaWithURL, error)

	// GetFile 获取媒体文件路径和MIME类型
	GetFile(id int64) (string, string, error)

	// GetThumbnail 获取缩略图路径和MIME类型
	GetThumbnail(id int64) (string, string, error)

	// Delete 删除媒体文件
	Delete(id int64, userID int64) error

	// ValidateType 验证文件类型
	ValidateType(filename string, mimeType string) (string, error)

	// ValidateSize 验证文件大小
	ValidateSize(size int64, mediaType string) error
}

// Processor 文件处理器接口
type Processor interface {
	// ProcessImage 处理图片，生成缩略图
	ProcessImage(src string, dst string) (width, height int, err error)

	// GetImageDimensions 获取图片尺寸
	GetImageDimensions(file io.Reader) (width, height int, err error)

	// ExtractAudioInfo 提取音频信息
	ExtractAudioInfo(path string) (duration int, err error)
}

// Storage 文件存储接口
type Storage interface {
	// SaveOriginal 保存原始文件
	SaveOriginal(id int64, ext string, file multipart.File) (string, int64, error)

	// SaveThumbnail 保存缩略图
	SaveThumbnail(id int64, data []byte) (string, error)

	// Delete 删除文件
	Delete(id int64) error

	// GetOriginalPath 获取原始文件路径
	GetOriginalPath(id int64) string

	// GetThumbnailPath 获取缩略图路径
	GetThumbnailPath(id int64) string

	// Rename 重命名文件
	Rename(oldPath, newPath string) error
}
