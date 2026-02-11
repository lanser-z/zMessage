package models

// Media 媒体模型
type Media struct {
	ID            int64  `json:"id"`
	OwnerID       int64  `json:"owner_id"`
	Type          string `json:"type"` // image, voice
	OriginalPath  string `json:"-"`    // 不对外暴露物理路径
	ThumbnailPath string `json:"-"`    // 不对外暴露物理路径
	Size          int64  `json:"size"`
	MimeType      string `json:"mime_type"`
	Width         *int   `json:"width,omitempty"`
	Height        *int   `json:"height,omitempty"`
	Duration      *int   `json:"duration,omitempty"` // 音频时长(秒)
	CreatedAt     int64  `json:"created_at"`
}

// MediaWithURL 带URL的媒体信息
type MediaWithURL struct {
	*Media
	OriginalURL  string `json:"original_url"`
	ThumbnailURL string `json:"thumbnail_url,omitempty"`
}
