package utils

import (
	"fmt"
	"image"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

// SaveUploadedFile menyimpan file upload ke direktori tujuan dengan validasi
func SaveUploadedFile(ctx *gin.Context, file *multipart.FileHeader, destDir string, filename string) (string, error) {
	// 1. Validasi ekstensi
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
		return "", fmt.Errorf("invalid file type: only jpg, jpeg, png allowed")
	}

	// 2. Validasi ukuran file (max 1 MB)
	const maxSize = 1 << 20 // 1 MB
	if file.Size > maxSize {
		return "", fmt.Errorf("file too large: maximum size is 1MB")
	}

	// 3. Validasi resolusi gambar (max 512x512)
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer src.Close()

	imgCfg, _, err := image.DecodeConfig(src)
	if err != nil {
		return "", fmt.Errorf("invalid image: %w", err)
	}
	if imgCfg.Width > 512 || imgCfg.Height > 512 {
		return "", fmt.Errorf("invalid resolution: max allowed is 512x512")
	}

	// Buat folder jika belum ada
	if err := os.MkdirAll(destDir, os.ModePerm); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	// Nama final file (contoh: profile_12.png)
	finalName := fmt.Sprintf("%s%s", filename, ext)
	fullPath := filepath.Join(destDir, finalName)

	// Simpan file
	if err := ctx.SaveUploadedFile(file, fullPath); err != nil {
		return "", fmt.Errorf("failed to save file: %w", err)
	}

	return fullPath, nil
}

// ukuran resolusi 512*512
// extensi jpeg, jpg, png
// ukuran maksimal 1 mb
