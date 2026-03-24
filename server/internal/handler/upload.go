package handler

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/neinei960/cat/server/config"
	"github.com/neinei960/cat/server/pkg/response"
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
	if file.Size > maxSize {
		response.Error(c, http.StatusBadRequest, fmt.Sprintf("文件过大，最大允许%dMB", maxSize/1024/1024))
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
	if err := os.Chmod(uploadPath, 0755); err != nil {
		response.Error(c, http.StatusInternalServerError, "设置目录权限失败")
		return
	}

	filename := uuid.New().String() + ext
	dst := filepath.Join(uploadPath, filename)

	if err := c.SaveUploadedFile(file, dst); err != nil {
		response.Error(c, http.StatusInternalServerError, "保存文件失败")
		return
	}
	os.Chmod(dst, 0644)

	url := "/uploads/" + filename
	response.Success(c, gin.H{"url": url})
}
