package media

import (
	"bytes"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"zmessage/server/dal"
	"zmessage/server/models"
)

// NewService 创建媒体服务
func NewService(dalMgr dal.Manager, dataDir string) Service {
	return &service{
		dal:      dalMgr,
		storage:   NewLocalStorage(dataDir),
		processor: NewBasicProcessor(),
	}
}

// service 媒体服务实现
type service struct {
	dal       dal.Manager
	storage   Storage
	processor Processor
}

// fileAdapter 适配器，使io.Reader兼容multipart.File接口
type fileAdapter struct {
	*bytes.Reader
	closeFunc func() error
}

func (f *fileAdapter) ReadAt(p []byte, off int64) (int, error) {
	// 简化实现：不支持随机访问
	return 0, io.ErrUnexpectedEOF
}

func (f *fileAdapter) Seek(offset int64, whence int) (int64, error) {
	// 不支持Seek
	return 0, io.ErrUnexpectedEOF
}

func (f *fileAdapter) Close() error {
	if f.closeFunc != nil {
		return f.closeFunc()
	}
	return nil
}

// Upload 上传媒体文件
func (s *service) Upload(req *UploadRequest) (*models.Media, error) {
	// 验证文件
	if req.File == nil {
		return nil, ErrFileRequired
	}

	// 验证类型
	ext := strings.ToLower(filepath.Ext(req.Header.Filename))
	mediaType, err := s.ValidateType(req.Header.Filename, req.Header.Header.Get("Content-Type"))
	if err != nil {
		return nil, err
	}

	if req.Type != "" && req.Type != mediaType {
		return nil, ErrInvalidMediaType
	}

	// 读取文件内容
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, req.File); err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}
	content := buf.Bytes()
	size := int64(len(content))

	// 验证大小
	if err := s.ValidateSize(size, mediaType); err != nil {
		return nil, err
	}

	// 创建媒体记录
	now := time.Now().Unix()
	media := &models.Media{
		OwnerID:  req.OwnerID,
		Type:      mediaType,
		Size:      size,
		MimeType:  req.Header.Header.Get("Content-Type"),
		CreatedAt: now,
	}

	// 保存原始文件
	contentReader := bytes.NewReader(content)
	adapter := &fileAdapter{
		Reader:    contentReader,
		closeFunc: func() error { return req.File.Close() },
	}
	originalPath, _, err := s.storage.SaveOriginal(0, ext, adapter)
	if err != nil {
		return nil, fmt.Errorf("save original: %w", err)
	}
	media.OriginalPath = originalPath

	// 处理文件（获取尺寸/时长）
	if mediaType == "image" {
		width, height, err := s.processor.GetImageDimensions(bytes.NewReader(content))
		if err == nil {
			media.Width = &width
			media.Height = &height
			// 生成缩略图
			_, _, _ = s.processor.ProcessImage(originalPath, "")
			media.ThumbnailPath = s.storage.GetThumbnailPath(0) // 临时路径
		}
	} else if mediaType == "voice" {
		duration, err := s.processor.ExtractAudioInfo(originalPath)
		if err == nil {
			media.Duration = &duration
		}
	}

	// 保存到数据库
	if err := s.dal.Media().Create(media); err != nil {
		return nil, fmt.Errorf("save to db: %w", err)
	}

	// 重命名文件使用实际ID
	if media.ID != 0 {
		// 构建带扩展名的最终路径
		finalPath := filepath.Join(filepath.Dir(originalPath), fmt.Sprintf("%d%s", media.ID, ext))
		finalThumbPath := s.storage.GetThumbnailPath(media.ID)

		// 重命名原始文件
		if err := s.storage.Rename(originalPath, finalPath); err != nil {
			// 尝试清理
			s.dal.Media().Delete(media.ID)
			return nil, fmt.Errorf("rename file: %w", err)
		}
		media.OriginalPath = finalPath

		// 重命名缩略图（如果有）
		if media.ThumbnailPath != "" {
			if err := s.storage.Rename(media.ThumbnailPath, finalThumbPath); err == nil {
				media.ThumbnailPath = finalThumbPath
			}
		}

		// 更新数据库中的路径
		if err := s.dal.Media().Update(media); err != nil {
			fmt.Printf("[MEDIA] Failed to update paths in DB: %v\n", err)
		}
	}

	return media, nil
}

// Get 获取媒体信息
func (s *service) Get(id int64) (*models.Media, error) {
	media, err := s.dal.Media().GetByID(id)
	if err != nil {
		return nil, ErrNotFound
	}
	return media, nil
}

// GetWithURL 获取带URL的媒体信息
func (s *service) GetWithURL(id int64, baseURL string) (*MediaWithURL, error) {
	media, err := s.Get(id)
	if err != nil {
		return nil, err
	}

	result := &MediaWithURL{
		Media:      media,
		// 返回相对路径，前端通过 <base href> 控制前缀
		OriginalURL: fmt.Sprintf("api/media/%d", id),
	}

	if media.ThumbnailPath != "" {
		result.ThumbnailURL = fmt.Sprintf("api/media/%d/thumb", id)
	}

	return result, nil
}

// GetFile 获取媒体文件
func (s *service) GetFile(id int64) (string, string, error) {
	media, err := s.Get(id)
	if err != nil {
		return "", "", err
	}
	return media.OriginalPath, media.MimeType, nil
}

// GetThumbnail 获取缩略图
func (s *service) GetThumbnail(id int64) (string, string, error) {
	media, err := s.Get(id)
	if err != nil {
		return "", "", err
	}

	if media.ThumbnailPath == "" {
		return "", "", ErrNotFound
	}

	return media.ThumbnailPath, "image/jpeg", nil
}

// Delete 删除媒体文件
func (s *service) Delete(id int64, userID int64) error {
	media, err := s.Get(id)
	if err != nil {
		return err
	}

	// 验证权限
	if media.OwnerID != userID {
		return ErrAccessDenied
	}

	// 删除数据库记录
	if err := s.dal.Media().Delete(id); err != nil {
		return fmt.Errorf("delete from db: %w", err)
	}

	// 删除文件
	if err := s.storage.Delete(id); err != nil {
		return fmt.Errorf("delete files: %w", err)
	}

	return nil
}

// ValidateType 验证文件类型
func (s *service) ValidateType(filename string, mimeType string) (string, error) {
	ext := strings.ToLower(filepath.Ext(filename))

	// 检查图片类型
	for allowedExt := range AllowedImageExts {
		if ext == allowedExt {
			return "image", nil
		}
	}

	// 检查语音类型
	for allowedExt := range AllowedVoiceExts {
		if ext == allowedExt {
			return "voice", nil
		}
	}

	return "", ErrInvalidType
}

// ValidateSize 验证文件大小
func (s *service) ValidateSize(size int64, mediaType string) error {
	switch mediaType {
	case "image":
		if size > MaxImageSize {
			return ErrInvalidSize
		}
	case "voice":
		if size > MaxVoiceSize {
			return ErrInvalidSize
		}
	default:
		return ErrInvalidMediaType
	}
	return nil
}
