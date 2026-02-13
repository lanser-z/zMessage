package media

import (
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"os"
)

// NewBasicProcessor 创建基础处理器实现
func NewBasicProcessor() Processor {
	return &basicProcessor{}
}

// basicProcessor 基础文件处理器
type basicProcessor struct{}

// ProcessImage 处理图片，生成缩略图
func (p *basicProcessor) ProcessImage(src string, dst string) (int, int, error) {
	// 读取原始图片
	file, err := os.Open(src)
	if err != nil {
		return 0, 0, fmt.Errorf("open image: %w", err)
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return 0, 0, fmt.Errorf("decode image: %w", err)
	}

	bounds := img.Bounds()
	origWidth := bounds.Dx()
	origHeight := bounds.Dy()

	// 计算缩略图尺寸（保持宽高比）
	var newWidth, newHeight int
	if origWidth > ThumbnailSize {
		newWidth = ThumbnailSize
		newHeight = int(float64(origHeight) * float64(ThumbnailSize) / float64(origWidth))
	} else {
		newWidth = origWidth
		newHeight = origHeight
	}

	// 创建缩略图
	thumb := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))

	// 简化：使用最近邻插值
	for y := 0; y < newHeight; y++ {
		for x := 0; x < newWidth; x++ {
			srcX := x * origWidth / newWidth
			srcY := y * origHeight / newHeight
			thumb.Set(x, y, img.At(srcX, srcY))
		}
	}

	// 保存缩略图
	thumbPath := dst
	if thumbPath == "" {
		// 如果没有指定目标路径，生成一个临时文件
		thumbPath = src + ".thumb"
	}

	outFile, err := os.Create(thumbPath)
	if err != nil {
		return newWidth, newHeight, fmt.Errorf("create thumbnail: %w", err)
	}
	defer outFile.Close()

	if err := jpeg.Encode(outFile, thumb, &jpeg.Options{Quality: ThumbnailQuality}); err != nil {
		return newWidth, newHeight, fmt.Errorf("encode thumbnail: %w", err)
	}

	return newWidth, newHeight, nil
}

// GetImageDimensions 获取图片尺寸
func (p *basicProcessor) GetImageDimensions(r io.Reader) (int, int, error) {
	// 尝试解码图片以获取尺寸
	img, _, err := image.Decode(r)
	if err != nil {
		return 0, 0, fmt.Errorf("decode image: %w", err)
	}

	bounds := img.Bounds()
	return bounds.Dx(), bounds.Dy(), nil
}

// ExtractAudioInfo 提取音频信息
func (p *basicProcessor) ExtractAudioInfo(path string) (int, error) {
	// 简化实现：返回一个默认值
	// 生产环境建议使用 github.com/dhowden/tag 或其他库

	// 基于文件大小估算时长（不准确但够用）
	fileInfo, err := os.Stat(path)
	if err != nil {
		return 0, fmt.Errorf("stat file: %w", err)
	}

	// 粗略估算：每秒约16KB（128kbps）
	estimatedSeconds := int(fileInfo.Size() / 16000)

	if estimatedSeconds < 1 {
		estimatedSeconds = 1
	} else if estimatedSeconds > 300 { // 最大5分钟
		estimatedSeconds = 300
	}

	return estimatedSeconds, nil
}
