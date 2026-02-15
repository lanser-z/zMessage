package media

import (
	"fmt"
)

var (
	// ErrInvalidType 不支持的文件类型
	ErrInvalidType = fmt.Errorf("invalid file type")

	// ErrInvalidSize 文件大小超出限制
	ErrInvalidSize = fmt.Errorf("file size exceeds limit")

	// ErrInvalidFormat 文件格式无效
	ErrInvalidFormat = fmt.Errorf("invalid file format")

	// ErrUploadFailed 上传失败
	ErrUploadFailed = fmt.Errorf("upload failed")

	// ErrNotFound 文件不存在
	ErrNotFound = fmt.Errorf("media not found")

	// ErrAccessDenied 无权访问
	ErrAccessDenied = fmt.Errorf("access denied")

	// ErrDeleteFailed 删除失败
	ErrDeleteFailed = fmt.Errorf("delete failed")

	// ErrInvalidMediaType 无效的媒体类型
	ErrInvalidMediaType = fmt.Errorf("invalid media type")

	// ErrFileRequired 文件必须
	ErrFileRequired = fmt.Errorf("file is required")

	// ErrProcessFailed 处理失败
	ErrProcessFailed = fmt.Errorf("process failed")
)

// 文件大小限制（字节）
const (
	MaxImageSize = 5 * 1024 * 1024  // 5MB
	MaxVoiceSize = 10 * 1024 * 1024 // 10MB
)

// 缩略图配置
const (
	ThumbnailSize  = 300  // 宽度300px
	ThumbnailQuality = 85   // JPEG质量
)

// 允许的文件扩展名
var AllowedImageExts = map[string]bool{
	".jpg":  true, ".jpeg": true, ".png": true,
	".gif": true, ".webp": true,
}

var AllowedVoiceExts = map[string]bool{
	".mp3": true, ".m4a": true, ".ogg": true, ".wav": true, ".webm": true,
}

// MIME类型映射
var MimeTypes = map[string]string{
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".png":  "image/png",
	".gif":  "image/gif",
	".webp": "image/webp",
	".mp3":  "audio/mpeg",
	".m4a":  "audio/mp4",
	".ogg":  "audio/ogg",
	".wav":  "audio/wav",
}
