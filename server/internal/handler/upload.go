package handler

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/neinei960/cat/server/config"
	"github.com/neinei960/cat/server/pkg/response"
	"golang.org/x/image/draw"
)

func Upload(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "请选择文件")
		return
	}

	maxSize := config.AppConfig.Upload.MaxSize
	if maxSize <= 0 {
		maxSize = 2 * 1024 * 1024 // 2MB default
	}
	if file.Size > 10*1024*1024 { // 允许上传最大10MB原图，压缩后会变小
		response.Error(c, http.StatusBadRequest, "文件过大，最大允许10MB")
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowed := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true}
	if !allowed[ext] {
		response.Error(c, http.StatusBadRequest, "不支持的文件类型，仅支持 jpg/png/gif/webp")
		return
	}

	uploadPath := config.AppConfig.Upload.Path
	if uploadPath == "" {
		uploadPath = "./uploads"
	}
	if err := os.MkdirAll(uploadPath, 0755); err != nil {
		response.Error(c, http.StatusInternalServerError, "创建目录失败")
		return
	}

	// 打开上传文件
	src, err := file.Open()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "读取文件失败")
		return
	}
	defer src.Close()

	// 读取文件内容
	data, err := io.ReadAll(src)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "读取文件失败")
		return
	}

	// 对 jpg/jpeg/png 做压缩，gif/webp 直接保存
	if ext == ".jpg" || ext == ".jpeg" || ext == ".png" {
		compressed, compErr := compressImage(data, ext)
		if compErr == nil {
			data = compressed
			ext = ".jpg" // 统一输出为 jpeg
		}
		// 压缩失败则保存原图
	}

	filename := uuid.New().String() + ext
	dst := filepath.Join(uploadPath, filename)

	if err := os.WriteFile(dst, data, 0644); err != nil {
		response.Error(c, http.StatusInternalServerError, "保存文件失败")
		return
	}

	url := "/uploads/" + filename
	response.Success(c, gin.H{"url": url})
}

// compressImage 压缩图片：限制最大尺寸 800px，JPEG 质量 80
func compressImage(data []byte, ext string) ([]byte, error) {
	reader := bytes.NewReader(data)

	var img image.Image
	var err error

	switch ext {
	case ".png":
		img, err = png.Decode(reader)
	case ".jpg", ".jpeg":
		img, err = jpeg.Decode(reader)
	default:
		return nil, fmt.Errorf("unsupported format")
	}
	if err != nil {
		return nil, err
	}

	// 缩放到最大 800px
	bounds := img.Bounds()
	w := bounds.Dx()
	h := bounds.Dy()
	maxDim := 800

	if w > maxDim || h > maxDim {
		var newW, newH int
		if w > h {
			newW = maxDim
			newH = h * maxDim / w
		} else {
			newH = maxDim
			newW = w * maxDim / h
		}
		dst := image.NewRGBA(image.Rect(0, 0, newW, newH))
		draw.BiLinear.Scale(dst, dst.Bounds(), img, img.Bounds(), draw.Over, nil)
		img = dst
	}

	// 编码为 JPEG，质量 80
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 80}); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
