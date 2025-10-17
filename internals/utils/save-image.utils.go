package utils

import (
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"github.com/disintegration/imaging"
	"github.com/gin-gonic/gin"
)

func SaveUploadedFile(ctx *gin.Context, file *multipart.FileHeader, destDir string, filename string) (string, error) {
	// 1. Validasi ekstensi
	ext := strings.ToLower(strings.TrimSpace(filepath.Ext(file.Filename)))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
		return "", fmt.Errorf("invalid file type: only jpg, jpeg, png allowed")
	}

	// 2. Validasi ukuran file (max 1 MB)
	const maxSize = 1 << 20 // 1 MB
	if file.Size > maxSize {
		return "", fmt.Errorf("file too large: maximum size is 1MB")
	}

	// 3. Baca isi file ke buffer supaya bisa di-decode ulang
	srcFile, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer srcFile.Close()

	// Decode image dari buffer
	img, format, err := image.Decode(srcFile)
	if err != nil {
		return "", fmt.Errorf("invalid image: %w", err)
	}

	// 4. Resize otomatis ke 512x512 (force square)
	resizedImg := imaging.Resize(img, 512, 512, imaging.Lanczos)

	// Buat folder jika belum ada
	if err := os.MkdirAll(destDir, os.ModePerm); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	// Nama final file
	finalName := fmt.Sprintf("%s%s", filename, ext)
	fullPath := filepath.Join(destDir, finalName)

	// 5. Simpan hasil resize
	out, err := os.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer out.Close()

	switch format {
	case "jpeg":
		err = jpeg.Encode(out, resizedImg, &jpeg.Options{Quality: 90})
	case "png":
		err = png.Encode(out, resizedImg)
	default:
		return "", fmt.Errorf("unsupported image format: %s", format)
	}
	if err != nil {
		return "", fmt.Errorf("failed to save image: %w", err)
	}

	return finalName, nil
}
