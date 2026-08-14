package api

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

const maxUploadBytes = 10 << 20 // 10MB

var allowedUploadTypes = map[string]string{
	"image/png":  ".png",
	"image/jpeg": ".jpg",
}

var errUnsupportedFileType = errors.New("unsupported file type: only PNG and JPEG images are allowed")

// saveUpload validates an uploaded file by sniffing its actual content
// (never trusting the client-supplied filename or Content-Type header for
// either validation or the on-disk path) and, if valid, writes it under
// uploadsDir with a generated, non-guessable filename. It returns the
// filename (not a full path) to store in the DB.
func saveUpload(uploadsDir string, file multipart.File, header *multipart.FileHeader) (string, error) {
	if header.Size > maxUploadBytes {
		return "", fmt.Errorf("file too large: %d bytes (max %d)", header.Size, maxUploadBytes)
	}

	sniff := make([]byte, 512)
	n, err := file.Read(sniff)
	if err != nil && err != io.EOF {
		return "", err
	}
	contentType := http.DetectContentType(sniff[:n])

	ext, ok := allowedUploadTypes[contentType]
	if !ok {
		return "", errUnsupportedFileType
	}

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", err
	}

	filename := uuid.NewString() + ext
	destPath := filepath.Join(uploadsDir, filename)

	dest, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return "", err
	}
	defer dest.Close()

	// http.MaxBytesReader on the request body already caps total request
	// size upstream; this LimitReader is a second, per-file backstop.
	if _, err := io.Copy(dest, io.LimitReader(file, maxUploadBytes+1)); err != nil {
		os.Remove(destPath)
		return "", err
	}

	return filename, nil
}
