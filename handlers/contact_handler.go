package handlers

import (
	"net/http"
	"regexp"
	"strings"

	"go-api/config"
	"go-api/models"
	"go-api/utils"

	"github.com/gin-gonic/gin"
)

// ===============================
// CONTACT MESSAGE HANDLERS
// ===============================

// CreateContactInput là input cho việc gửi tin nhắn liên hệ
type CreateContactInput struct {
	Name    string `json:"name" binding:"required"`
	Phone   string `json:"phone" binding:"required"`
	Email   string `json:"email" binding:"required,email"`
	Subject string `json:"subject"`
	Message string `json:"message" binding:"required"`
}

// CreateContactMessage xử lý gửi tin nhắn liên hệ
// @Summary Gửi tin nhắn liên hệ
// @Description Khách hàng gửi tin nhắn liên hệ đến hệ thống
// @Tags Contact
// @Accept json
// @Produce json
// @Param contact body CreateContactInput true "Thông tin liên hệ"
// @Success 200 {object} map[string]interface{}
// @Router /contact [post]
func CreateContactMessage(c *gin.Context) {
	var input CreateContactInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Vui lòng điền đầy đủ thông tin", "VALIDATION_ERROR", err.Error())
		return
	}

	// Validate phone number (Vietnam format)
	phoneRegex := regexp.MustCompile(`^(0[35789])[0-9]{8}$`)
	phone := strings.ReplaceAll(input.Phone, " ", "")
	phone = strings.ReplaceAll(phone, "-", "")
	if !phoneRegex.MatchString(phone) {
		utils.ErrorResponse(c, http.StatusBadRequest, "Số điện thoại không hợp lệ", "INVALID_PHONE", "")
		return
	}

	// Validate message length
	if len(strings.TrimSpace(input.Message)) < 10 {
		utils.ErrorResponse(c, http.StatusBadRequest, "Tin nhắn phải có ít nhất 10 ký tự", "MESSAGE_TOO_SHORT", "")
		return
	}

	// Default subject
	subject := input.Subject
	if subject == "" {
		subject = "Đăng ký dùng thử"
	}

	// Create contact message
	contactMessage := models.ContactMessage{
		Name:    strings.TrimSpace(input.Name),
		Phone:   phone,
		Email:   strings.ToLower(strings.TrimSpace(input.Email)),
		Subject: subject,
		Message: strings.TrimSpace(input.Message),
		Status:  "new",
	}

	if err := config.GetDB().Create(&contactMessage).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Không thể gửi tin nhắn", "CREATE_ERROR", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, gin.H{
		"id":      contactMessage.ID,
		"message": "Cảm ơn bạn đã liên hệ! Chúng tôi sẽ phản hồi trong thời gian sớm nhất.",
	}, "Gửi tin nhắn thành công")
}

// GetContactMessages lấy danh sách tin nhắn liên hệ (Admin only)
// @Summary Lấy danh sách tin nhắn liên hệ
// @Description Admin lấy tất cả tin nhắn liên hệ
// @Tags Contact
// @Produce json
// @Param status query string false "Filter theo status" Enums(new, read, replied, closed)
// @Success 200 {object} map[string]interface{}
// @Security BearerAuth
// @Router /admin/contacts [get]
func GetContactMessages(c *gin.Context) {
	var messages []models.ContactMessage

	query := config.GetDB().Order("created_at DESC")

	// Filter by status
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Find(&messages).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Không thể lấy danh sách tin nhắn", "FETCH_ERROR", err.Error())
		return
	}

	var result []gin.H
	for _, msg := range messages {
		result = append(result, gin.H{
			"id":         msg.ID,
			"name":       msg.Name,
			"phone":      msg.Phone,
			"email":      msg.Email,
			"subject":    msg.Subject,
			"message":    msg.Message,
			"status":     msg.Status,
			"note":       msg.Note,
			"replied_at": msg.RepliedAt,
			"created_at": msg.CreatedAt,
		})
	}

	utils.SuccessResponse(c, http.StatusOK, gin.H{
		"messages": result,
		"total":    len(result),
	}, "Lấy danh sách tin nhắn thành công")
}

// UpdateContactMessageStatus cập nhật trạng thái tin nhắn (Admin only)
// @Summary Cập nhật trạng thái tin nhắn
// @Description Admin cập nhật trạng thái tin nhắn liên hệ
// @Tags Contact
// @Accept json
// @Produce json
// @Param id path int true "Contact Message ID"
// @Param body body object true "Status update"
// @Success 200 {object} map[string]interface{}
// @Security BearerAuth
// @Router /admin/contacts/{id} [put]
func UpdateContactMessageStatus(c *gin.Context) {
	id := c.Param("id")

	var msg models.ContactMessage
	if err := config.GetDB().First(&msg, id).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Không tìm thấy tin nhắn", "NOT_FOUND", "")
		return
	}

	var input struct {
		Status string  `json:"status"`
		Note   *string `json:"note"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Dữ liệu không hợp lệ", "VALIDATION_ERROR", err.Error())
		return
	}

	updates := make(map[string]interface{})
	if input.Status != "" {
		validStatuses := []string{"new", "read", "replied", "closed"}
		isValid := false
		for _, s := range validStatuses {
			if s == input.Status {
				isValid = true
				break
			}
		}
		if !isValid {
			utils.ErrorResponse(c, http.StatusBadRequest, "Trạng thái không hợp lệ", "INVALID_STATUS", "")
			return
		}
		updates["status"] = input.Status
	}

	if input.Note != nil {
		updates["note"] = *input.Note
	}

	if err := config.GetDB().Model(&msg).Updates(updates).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Không thể cập nhật", "UPDATE_ERROR", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, nil, "Cập nhật thành công")
}

// DeleteContactMessage xóa tin nhắn (Admin only)
// @Summary Xóa tin nhắn
// @Description Admin xóa tin nhắn liên hệ
// @Tags Contact
// @Produce json
// @Param id path int true "Contact Message ID"
// @Success 200 {object} map[string]interface{}
// @Security BearerAuth
// @Router /admin/contacts/{id} [delete]
func DeleteContactMessage(c *gin.Context) {
	id := c.Param("id")

	if err := config.GetDB().Delete(&models.ContactMessage{}, id).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Không thể xóa tin nhắn", "DELETE_ERROR", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, nil, "Xóa tin nhắn thành công")
}
