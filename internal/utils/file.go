package utils

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

var allowExts = map[string]bool{
	".png":  true,
	".jpg":  true,
	".jpeg": true,
}
var allowMimeTypes = map[string]bool{
	"image/png":  true,
	"image/jpeg": true,
	"image/jpg":  true,
}

const maxSize = 5 << 20

func ValiDateAndSaveFile(fileheader *multipart.FileHeader, uploadDir string) (string, error) {
	ext := strings.ToLower(filepath.Ext(fileheader.Filename))
	if !allowExts[ext] {
		return "", errors.New("file extention not support")
	}
	if fileheader.Size > maxSize {
		return "", errors.New("file too large (max 5MB)")
	}
	file, err := fileheader.Open()
	if err != nil {
		return "", errors.New("can't open file")
	}
	defer file.Close()

	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil {
		return "", errors.New("can't read file")
	}
	mimeType := http.DetectContentType(buffer)
	if !allowMimeTypes[mimeType] {
		return "", fmt.Errorf("file mime type not support %s", mimeType)
	}
	filename := fmt.Sprintf("%s%s", uuid.New().String(), ext)

	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		return "", errors.New("can't create directory")
	}

	savePath := filepath.Join(uploadDir, filename)
	if err := saveFile(fileheader, savePath); err != nil {
		return "", err
	}

	return filename, nil
}
func saveFile(fileheader *multipart.FileHeader, destination string) error {
	src, err := fileheader.Open()
	if err != nil {
		return err
	}
	defer src.Close()
	out, err := os.Create(destination)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, src)

	return err
}
