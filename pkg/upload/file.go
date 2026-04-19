package upload

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func prepareUploadedFile(userID uint, originalFilename, dir, routePrefix, filenamePrefix string, allowedExts map[string]struct{}, unsupportedErr error) (savePath, fileURL string, err error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", "", err
	}

	ext := strings.ToLower(filepath.Ext(originalFilename))
	if _, ok := allowedExts[ext]; !ok {
		return "", "", unsupportedErr
	}

	filename := fmt.Sprintf("%s_%d_%d%s", filenamePrefix, userID, time.Now().UnixNano(), ext)
	return filepath.Join(dir, filename), routePrefix + "/" + filename, nil
}

func SaveFile(file *multipart.FileHeader, savePath string) error {
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer func() {
		_ = src.Close()
	}()

	if err := os.MkdirAll(filepath.Dir(savePath), 0o755); err != nil {
		return err
	}

	dst, err := os.Create(savePath)
	if err != nil {
		return err
	}
	defer func() {
		_ = dst.Close()
	}()

	if _, err = io.Copy(dst, src); err != nil {
		return err
	}

	return nil
}

func removeUploadedFile(fileURL, routePrefix, dir string) error {
	if fileURL == "" {
		return nil
	}

	filename := strings.TrimPrefix(fileURL, routePrefix+"/")
	err := os.Remove(filepath.Join(dir, filename))
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}
