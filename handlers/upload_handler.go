package handlers

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"go-api/services"
	"go-api/utils"

	"github.com/gin-gonic/gin"
)

// ===============================
// UPLOAD HANDLERS - Cloudinary
// ===============================

const (
	maxFileSize = 5 * 1024 * 1024 // 5MB
)

// UploadImage upload ảnh lên Cloudinary
// @Summary Upload ảnh
// @Description Upload ảnh lên Cloudinary cloud storage (menu, category, restaurant, avatar)
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

	// Upload lên Cloudinary với folder structure: qr-restaurant/{user_id}/{folder}
	userIDStr := fmt.Sprintf("%v", userID)
	cloudinaryFolder := fmt.Sprintf("qr-restaurant/%s/%s", userIDStr, folder)

	uploadResult, err := services.UploadImage(file, header.Filename, cloudinaryFolder)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Không thể upload ảnh lên Cloudinary", "CLOUDINARY_UPLOAD_ERROR", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, gin.H{
		"url":       uploadResult.SecureURL,
		"public_id": uploadResult.PublicID,
		"filename":  header.Filename,
		"folder":    folder,
		"width":     uploadResult.Width,
		"height":    uploadResult.Height,
		"format":    uploadResult.Format,
		"size":      uploadResult.Bytes,
	}, "Upload ảnh thành công")
}

// UploadMultipleImages upload nhiều ảnh
// @Summary Upload nhiều ảnh
// @Description Upload nhiều ảnh vào server (tối đa 10 ảnh)
// UploadMultipleImages upload nhiều ảnh lên Cloudinary
// @Summary Upload nhiều ảnh
// @Description Upload nhiều ảnh lên Cloudinary (tối đa 10 ảnh)
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
	cloudinaryFolder := fmt.Sprintf("qr-restaurant/%s/%s", userIDStr, folder)

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

		// Upload lên Cloudinary
		uploadResult, err := services.UploadImage(file, fileHeader.Filename, cloudinaryFolder)
		file.Close()

		if err != nil {
			errors = append(errors, fileHeader.Filename+": không thể upload lên Cloudinary")
			continue
		}

		results = append(results, gin.H{
			"original_name": fileHeader.Filename,
			"url":           uploadResult.SecureURL,
			"public_id":     uploadResult.PublicID,
			"format":        uploadResult.Format,
			"width":         uploadResult.Width,
			"height":        uploadResult.Height,
			"size":          uploadResult.Bytes,
		})
	}

	utils.SuccessResponse(c, http.StatusOK, gin.H{
		"uploaded": results,
		"errors":   errors,
		"total":    len(results),
	}, "Upload hoàn tất")
}

// DeleteImageHandler xóa ảnh từ Cloudinary
// @Summary Xóa ảnh
// @Description Xóa ảnh từ Cloudinary bằng URL
// @Tags Upload
// @Accept json
// @Produce json
// @Param body body object true "URL ảnh Cloudinary" example({"url": "https://res.cloudinary.com/exe2/image/upload/v123456/qr-restaurant/1/menu/abc.jpg"})
// @Success 200 {object} map[string]interface{}
// @Security BearerAuth
// @Router /upload/image [delete]
func DeleteImageHandler(c *gin.Context) {
	// Lấy user_id từ context
	_, exists := c.Get("user_id")
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

	// Xóa ảnh từ Cloudinary
	if err := services.DeleteImageByURL(input.URL); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Không thể xóa ảnh từ Cloudinary", "DELETE_ERROR", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, nil, "Xóa ảnh thành công")
}

// UploadImageFromURL upload ảnh từ URL lên Cloudinary
// @Summary Upload ảnh từ URL
// @Description Upload ảnh từ URL lên Cloudinary
// @Tags Upload
// @Accept json
// @Produce json
// @Param body body object true "URL ảnh" example({"url": "https://example.com/image.jpg", "folder": "menu"})
// @Success 200 {object} map[string]interface{}
// @Security BearerAuth
// @Router /upload/url [post]
func UploadImageFromURL(c *gin.Context) {
	// Lấy user_id từ context
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Vui lòng đăng nhập", "UNAUTHORIZED", "")
		return
	}

	var input struct {
		URL    string `json:"url" binding:"required"`
		Folder string `json:"folder"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Vui lòng cung cấp URL của ảnh", "VALIDATION_ERROR", err.Error())
		return
	}

	if input.Folder == "" {
		input.Folder = "menu"
	}

	// Upload từ URL lên Cloudinary
	userIDStr := fmt.Sprintf("%v", userID)
	cloudinaryFolder := fmt.Sprintf("qr-restaurant/%s/%s", userIDStr, input.Folder)

	uploadResult, err := services.UploadFromURL(input.URL, cloudinaryFolder)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Không thể upload ảnh từ URL", "UPLOAD_ERROR", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, gin.H{
		"url":       uploadResult.SecureURL,
		"public_id": uploadResult.PublicID,
		"width":     uploadResult.Width,
		"height":    uploadResult.Height,
		"format":    uploadResult.Format,
		"size":      uploadResult.Bytes,
	}, "Upload ảnh từ URL thành công")
}

// TestCloudinaryUpload test Cloudinary connection
// @Summary Test Cloudinary
// @Description Test upload một ảnh lên Cloudinary để kiểm tra kết nối
// @Tags Upload
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /upload/test [get]
func TestCloudinaryUpload(c *gin.Context) {
	testURL := "https://images.unsplash.com/photo-1546069901-ba9599a7e63c?w=300"

	uploadResult, err := services.UploadFromURL(testURL, "qr-restaurant/test")
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Cloudinary test thất bại", "TEST_FAILED", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, gin.H{
		"status":    "success",
		"url":       uploadResult.SecureURL,
		"public_id": uploadResult.PublicID,
		"message":   "Cloudinary đã được cấu hình thành công!",
	}, "Test thành công")
}
