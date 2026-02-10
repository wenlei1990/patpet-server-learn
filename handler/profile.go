package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"patpet-server/model"
)

// ProfileHandler 个人信息
type ProfileHandler struct {
	DB *gorm.DB
}

// GetProfile 获取当前用户信息（需 JWT）
func (h *ProfileHandler) GetProfile(c *gin.Context) {
	userID := c.GetUint("user_id")

	var user model.User
	if err := h.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 1, "message": "user not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": user,
	})
}
