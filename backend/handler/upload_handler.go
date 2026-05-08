package handler

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

func UploadRoomLogo(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(10 << 20) // 10MB
	if err != nil {
		http.Error(w, "Failed to parse multipart form", http.StatusBadRequest)
		return
	}

	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Error retrieving file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// 检查文件类型
	ext := strings.ToLower(filepath.Ext(handler.Filename))
	allowedExts := []string{".jpg", ".jpeg", ".png", ".gif", ".webp"}
	isAllowed := false
	for _, e := range allowedExts {
		if ext == e {
			isAllowed = true
			break
		}
	}
	if !isAllowed {
		http.Error(w, "Invalid file type", http.StatusBadRequest)
		return
	}

	// 生成唯一文件名
	fileName := uuid.New().String() + ext

	// 获取当前工作目录
	cwd, err := os.Getwd()
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// 确保目录存在
	uploadDir := filepath.Join(cwd, "../frontend/public/upload/room_logo")
	err = os.MkdirAll(uploadDir, 0755)
	if err != nil {
		http.Error(w, "Failed to create upload directory", http.StatusInternalServerError)
		return
	}

	// 创建目标文件
	dstPath := filepath.Join(uploadDir, fileName)
	dst, err := os.Create(dstPath)
	if err != nil {
		http.Error(w, "Failed to create file", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	// 复制文件内容
	_, err = dst.ReadFrom(file)
	if err != nil {
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}

	// 返回文件URL
	response := fmt.Sprintf(`{"url": "/upload/room_logo/%s"}`, fileName)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(response))
}
