package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
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
		c.JSON(http.StatusNotFound, gin.H{"code": 1, "message": "用户不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": user,
	})
}

// UpdateProfileRequest 修改个人信息请求
type UpdateProfileRequest struct {
	Nickname *string `json:"nickname" binding:"omitempty,min=1,max=20"`
	Avatar   *string `json:"avatar" binding:"omitempty,url"`
}

// UpdateProfile 修改个人信息（昵称、头像）
func (h *ProfileHandler) UpdateProfile(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "message": "参数错误: " + err.Error()})
		return
	}

	// 构建要更新的字段（只更新传了的字段）
	updates := map[string]interface{}{}
	if req.Nickname != nil {
		updates["nickname"] = *req.Nickname
	}
	if req.Avatar != nil {
		updates["avatar"] = *req.Avatar
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "message": "没有需要更新的字段"})
		return
	}

	if err := h.DB.Model(&model.User{}).Where("id = ?", userID).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "message": "更新失败"})
		return
	}

	// 返回更新后的用户信息
	var user model.User
	h.DB.First(&user, userID)

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "更新成功",
		"data":    user,
	})
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

// ChangePassword 修改密码
func (h *ProfileHandler) ChangePassword(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "message": "参数错误: " + err.Error()})
		return
	}

	// 查找用户
	var user model.User
	if err := h.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 1, "message": "用户不存在"})
		return
	}

	// 验证旧密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.OldPassword)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "message": "原密码错误"})
		return
	}

	// 新旧密码不能相同
	if req.OldPassword == req.NewPassword {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "message": "新密码不能与原密码相同"})
		return
	}

	// 加密新密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "message": "服务器内部错误"})
		return
	}

	h.DB.Model(&user).Update("password", string(hashedPassword))

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "密码修改成功",
	})
}
