package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go-api/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ===============================
// UPLOAD HANDLERS - Local Storage
// ===============================

const (
	uploadBasePath = "./assets/img"
	maxFileSize    = 5 * 1024 * 1024 // 5MB
)

// UploadImage upload ảnh vào local storage
// @Summary Upload ảnh
// @Description Upload ảnh vào server (menu, category, restaurant, avatar)
// @Tags Upload
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "File ảnh (jpg, png, gif, webp)"
// @Param folder formData string false "Thư mục lưu trữ" default(menu)
// @Success 200 {object} map[string]interface{}
// @Security BearerAuth
// @Router /upload/image [post]
func UploadImage(c *gin.Context) {
	// Lấy user_id từ context (đã được set bởi auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Vui lòng đăng nhập", "UNAUTHORIZED", "")
		return
	}

	// Lấy file từ form
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Vui lòng chọn file ảnh", "FILE_REQUIRED", err.Error())
		return
	}
	defer file.Close()

	// Validate file type
	ext := strings.ToLower(filepath.Ext(header.Filename))
	allowedExts := []string{".jpg", ".jpeg", ".png", ".gif", ".webp"}
	isValid := false
	for _, allowed := range allowedExts {
		if ext == allowed {
			isValid = true
			break
		}
	}
	if !isValid {
		utils.ErrorResponse(c, http.StatusBadRequest, "Định dạng file không hợp lệ. Chấp nhận: jpg, png, gif, webp", "INVALID_FILE_TYPE", "")
		return
	}

	// Validate file size (max 5MB)
	if header.Size > maxFileSize {
		utils.ErrorResponse(c, http.StatusBadRequest, "File quá lớn. Tối đa 5MB", "FILE_TOO_LARGE", "")
		return
	}

	// Lấy folder từ form (mặc định là "menu")
	folder := c.DefaultPostForm("folder", "menu")
	allowedFolders := []string{"menu", "category", "restaurant", "avatar", "qr"}
	isValidFolder := false
	for _, f := range allowedFolders {
		if f == folder {
			isValidFolder = true
			break
		}
	}
	if !isValidFolder {
		folder = "menu"
	}

	// Tạo đường dẫn: assets/img/{user_id}/{folder}/
	userIDStr := fmt.Sprintf("%v", userID)
	uploadDir := filepath.Join(uploadBasePath, userIDStr, folder)

	// Tạo thư mục nếu chưa tồn tại
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Không thể tạo thư mục", "CREATE_DIR_ERROR", err.Error())
		return
	}

	// Tạo tên file unique: timestamp_uuid.ext
	filename := fmt.Sprintf("%d_%s%s", time.Now().UnixNano(), uuid.New().String()[:8], ext)
	filePath := filepath.Join(uploadDir, filename)

	// Tạo file đích
	dst, err := os.Create(filePath)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Không thể tạo file", "CREATE_FILE_ERROR", err.Error())
		return
	}
	defer dst.Close()

	// Copy nội dung file
	if _, err := io.Copy(dst, file); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Không thể lưu file", "SAVE_FILE_ERROR", err.Error())
		return
	}

	// Tạo URL để truy cập ảnh
	// URL format: /assets/img/{user_id}/{folder}/{filename}
	relativePath := fmt.Sprintf("/assets/img/%s/%s/%s", userIDStr, folder, filename)

	// Tạo full URL với domain
	baseURL := "https://apiqrcodeexe201-production.up.railway.app"
	fullURL := baseURL + relativePath

	// Lấy thông tin file
	fileInfo, _ := os.Stat(filePath)

	utils.SuccessResponse(c, http.StatusOK, gin.H{
		"url":      fullURL,
		"path":     relativePath,
		"filename": filename,
		"folder":   folder,
		"size":     fileInfo.Size(),
		"format":   strings.TrimPrefix(ext, "."),
	}, "Upload ảnh thành công")
}

// UploadMultipleImages upload nhiều ảnh
// @Summary Upload nhiều ảnh
// @Description Upload nhiều ảnh vào server (tối đa 10 ảnh)
// @Tags Upload
// @Accept multipart/form-data
// @Produce json
// @Param files formData file true "Các file ảnh"
// @Param folder formData string false "Thư mục lưu trữ" default(menu)
// @Success 200 {object} map[string]interface{}
// @Security BearerAuth
// @Router /upload/images [post]
func UploadMultipleImages(c *gin.Context) {
	// Lấy user_id từ context
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Vui lòng đăng nhập", "UNAUTHORIZED", "")
		return
	}

	form, err := c.MultipartForm()
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Vui lòng chọn file ảnh", "FILES_REQUIRED", err.Error())
		return
	}

	files := form.File["files"]
	if len(files) == 0 {
		utils.ErrorResponse(c, http.StatusBadRequest, "Vui lòng chọn ít nhất 1 file ảnh", "FILES_REQUIRED", "")
		return
	}

	if len(files) > 10 {
		utils.ErrorResponse(c, http.StatusBadRequest, "Tối đa 10 ảnh mỗi lần upload", "TOO_MANY_FILES", "")
		return
	}

	folder := c.DefaultPostForm("folder", "menu")
	userIDStr := fmt.Sprintf("%v", userID)
	uploadDir := filepath.Join(uploadBasePath, userIDStr, folder)

	// Tạo thư mục
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Không thể tạo thư mục", "CREATE_DIR_ERROR", err.Error())
		return
	}

	var results []gin.H
	var errors []string

	for _, fileHeader := range files {
		// Validate file type
		ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
		allowedExts := []string{".jpg", ".jpeg", ".png", ".gif", ".webp"}
		isValid := false
		for _, allowed := range allowedExts {
			if ext == allowed {
				isValid = true
				break
			}
		}
		if !isValid {
			errors = append(errors, fileHeader.Filename+": định dạng không hợp lệ")
			continue
		}

		// Validate size
		if fileHeader.Size > maxFileSize {
			errors = append(errors, fileHeader.Filename+": file quá lớn")
			continue
		}

		file, err := fileHeader.Open()
		if err != nil {
			errors = append(errors, fileHeader.Filename+": không thể đọc file")
			continue
		}

		// Tạo tên file unique
		filename := fmt.Sprintf("%d_%s%s", time.Now().UnixNano(), uuid.New().String()[:8], ext)
		filePath := filepath.Join(uploadDir, filename)

		// Tạo và lưu file
		dst, err := os.Create(filePath)
		if err != nil {
			file.Close()
			errors = append(errors, fileHeader.Filename+": không thể tạo file")
			continue
		}

		_, err = io.Copy(dst, file)
		dst.Close()
		file.Close()

		if err != nil {
			errors = append(errors, fileHeader.Filename+": không thể lưu file")
			continue
		}

		baseURL := "https://apiqrcodeexe201-production.up.railway.app"
		relativePath := fmt.Sprintf("/assets/img/%s/%s/%s", userIDStr, folder, filename)
		fullURL := baseURL + relativePath
		results = append(results, gin.H{
			"original_name": fileHeader.Filename,
			"filename":      filename,
			"url":           fullURL,
			"path":          relativePath,
		})
	}

	utils.SuccessResponse(c, http.StatusOK, gin.H{
		"uploaded": results,
		"errors":   errors,
		"total":    len(results),
	}, "Upload hoàn tất")
}

// DeleteImageHandler xóa ảnh từ local storage
// @Summary Xóa ảnh
// @Description Xóa ảnh từ server bằng đường dẫn
// @Tags Upload
// @Accept json
// @Produce json
// @Param body body object true "Đường dẫn ảnh" example({"url": "/assets/img/1/menu/abc123.jpg"})
// @Success 200 {object} map[string]interface{}
// @Security BearerAuth
// @Router /upload/image [delete]
func DeleteImageHandler(c *gin.Context) {
	// Lấy user_id từ context
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Vui lòng đăng nhập", "UNAUTHORIZED", "")
		return
	}

	var input struct {
		URL string `json:"url" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Vui lòng cung cấp url của ảnh", "VALIDATION_ERROR", err.Error())
		return
	}

	// Validate URL thuộc về user hiện tại
	userIDStr := fmt.Sprintf("%v", userID)
	expectedPrefix := fmt.Sprintf("/assets/img/%s/", userIDStr)

	// Admin có thể xóa tất cả
	role, _ := c.Get("role")
	if role != "admin" && !strings.HasPrefix(input.URL, expectedPrefix) {
		utils.ErrorResponse(c, http.StatusForbidden, "Bạn không có quyền xóa ảnh này", "FORBIDDEN", "")
		return
	}

	// Chuyển URL thành đường dẫn file
	// /assets/img/1/menu/abc.jpg -> ./assets/img/1/menu/abc.jpg
	filePath := "." + input.URL

	// Kiểm tra file tồn tại
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		utils.ErrorResponse(c, http.StatusNotFound, "Ảnh không tồn tại", "FILE_NOT_FOUND", "")
		return
	}

	// Xóa file
	if err := os.Remove(filePath); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Không thể xóa ảnh", "DELETE_ERROR", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, nil, "Xóa ảnh thành công")
}

// UploadImageFromURL không còn hỗ trợ (đã bỏ Cloudinary)
// @Summary Upload ảnh từ URL
// @Description Tính năng này đã bị tắt
// @Tags Upload
// @Accept json
// @Produce json
// @Success 501 {object} map[string]interface{}
// @Security BearerAuth
// @Router /upload/url [post]
func UploadImageFromURL(c *gin.Context) {
	utils.ErrorResponse(c, http.StatusNotImplemented, "Tính năng upload từ URL đã bị tắt", "NOT_IMPLEMENTED", "")
}
