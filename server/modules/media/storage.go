package media

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
)

// NewLocalStorage 创建本地存储实现
func NewLocalStorage(dataDir string) Storage {
	uploadDir := filepath.Join(dataDir, "uploads")
	originalDir := filepath.Join(uploadDir, "original")
	thumbnailDir := filepath.Join(uploadDir, "thumbnail")

	// 确保目录存在
	os.MkdirAll(originalDir, 0755)
	os.MkdirAll(thumbnailDir, 0755)

	return &localStorage{
		originalDir:   originalDir,
		thumbnailDir: thumbnailDir,
	}
}

// localStorage 本地文件存储实现
type localStorage struct {
	originalDir  string
	thumbnailDir string
}

// SaveOriginal 保存原始文件
func (s *localStorage) SaveOriginal(id int64, ext string, file multipart.File) (string, int64, error) {
	filename := fmt.Sprintf("temp_%d%s", id, ext)
	filePath := filepath.Join(s.originalDir, filename)

	dst, err := os.Create(filePath)
	if err != nil {
		return "", 0, fmt.Errorf("create file: %w", err)
	}
	defer dst.Close()

	written, err := io.Copy(dst, file)
	if err != nil {
		return "", 0, fmt.Errorf("copy file: %w", err)
	}

	return filePath, written, nil
}

// SaveThumbnail 保存缩略图
func (s *localStorage) SaveThumbnail(id int64, data []byte) (string, error) {
	filename := fmt.Sprintf("%d.jpg", id)
	filePath := filepath.Join(s.thumbnailDir, filename)

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return "", fmt.Errorf("write thumbnail: %w", err)
	}

	return filePath, nil
}

// Delete 删除文件
func (s *localStorage) Delete(id int64) error {
	// 删除所有可能扩展名的原始文件
	for _, ext := range []string{".jpg", ".jpeg", ".png", ".gif", ".webp", ".mp3", ".wav", ".ogg", ".webm"} {
		originalPath := filepath.Join(s.originalDir, fmt.Sprintf("%d%s", id, ext))
		if _, err := os.Stat(originalPath); err == nil {
			if err := os.Remove(originalPath); err != nil {
				return fmt.Errorf("delete original: %w", err)
			}
		}
	}

	// 删除缩略图
	thumbnailPath := s.GetThumbnailPath(id)
	if _, err := os.Stat(thumbnailPath); err == nil {
		if err := os.Remove(thumbnailPath); err != nil {
			return fmt.Errorf("delete thumbnail: %w", err)
		}
	}

	return nil
}

// GetOriginalPath 获取原始文件路径（带扩展名）
func (s *localStorage) GetOriginalPath(id int64) string {
	// 尝试查找有扩展名的文件
	entries, err := os.ReadDir(s.originalDir)
	if err == nil {
		for _, entry := range entries {
			if strings.HasPrefix(entry.Name(), fmt.Sprintf("%d.", id)) {
				return filepath.Join(s.originalDir, entry.Name())
			}
		}
	}
	// 如果没找到，返回不带扩展名的路径（兼容旧数据）
	return filepath.Join(s.originalDir, fmt.Sprintf("%d", id))
}

// GetThumbnailPath 获取缩略图路径
func (s *localStorage) GetThumbnailPath(id int64) string {
	return filepath.Join(s.thumbnailDir, fmt.Sprintf("%d.jpg", id))
}

// Rename 重命名文件
func (s *localStorage) Rename(oldPath, newPath string) error {
	return os.Rename(oldPath, newPath)
}
