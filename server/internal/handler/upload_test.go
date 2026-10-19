package handler

import (
	"bytes"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/neinei960/cat/server/config"
)

func TestUploadPreserveOriginalKeepsLargeImage(t *testing.T) {
	gin.SetMode(gin.TestMode)

	originalPath := config.AppConfig.Upload.Path
	originalMaxSize := config.AppConfig.Upload.MaxSize
	uploadDir := t.TempDir()
	config.AppConfig.Upload.Path = uploadDir
	config.AppConfig.Upload.MaxSize = 10 * 1024 * 1024
	t.Cleanup(func() {
		config.AppConfig.Upload.Path = originalPath
		config.AppConfig.Upload.MaxSize = originalMaxSize
	})

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	if err := writer.WriteField("preserve_original", "1"); err != nil {
		t.Fatalf("write preserve_original field: %v", err)
	}

	part, err := writer.CreateFormFile("file", "care-report.png")
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}

	source := image.NewRGBA(image.Rect(0, 0, 1279, 1809))
	fillTestImage(source)
	if err := png.Encode(part, source); err != nil {
		t.Fatalf("encode source image: %v", err)
	}

	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = req

	Upload(ctx)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", recorder.Code, recorder.Body.String())
	}

	var response struct {
		Code int `json:"code"`
		Data struct {
			URL string `json:"url"`
		} `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.Code != 0 || response.Data.URL == "" {
		t.Fatalf("unexpected response: %s", recorder.Body.String())
	}

	savedPath := filepath.Join(uploadDir, filepath.Base(response.Data.URL))
	file, err := os.Open(savedPath)
	if err != nil {
		t.Fatalf("open saved file: %v", err)
	}
	defer file.Close()

	config, _, err := image.DecodeConfig(file)
	if err != nil {
		t.Fatalf("decode saved image config: %v", err)
	}
	if config.Width != 1279 || config.Height != 1809 {
		t.Fatalf("expected preserved size 1279x1809, got %dx%d", config.Width, config.Height)
	}
}

func fillTestImage(img *image.RGBA) {
	for y := 0; y < img.Bounds().Dy(); y++ {
		for x := 0; x < img.Bounds().Dx(); x++ {
			img.Set(x, y, color.RGBA{
				R: uint8((x * 17) % 255),
				G: uint8((y * 29) % 255),
				B: uint8((x + y) % 255),
				A: 255,
			})
		}
	}
}
