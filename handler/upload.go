package handler

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"patpet-server/model"
)

// UploadHandler 文件上传
type UploadHandler struct {
	DB         *gorm.DB
	Cloudinary *cloudinary.Cloudinary
}

// UploadAvatar 上传头像
func (h *UploadHandler) UploadAvatar(c *gin.Context) {
	userID := c.GetUint("user_id")

	// 1. 从请求中获取文件
	file, header, err := c.Request.FormFile("avatar")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "message": "请选择要上传的图片"})
		return
	}
	defer file.Close()

	// 2. 校验文件类型
	ext := filepath.Ext(header.Filename)
	allowed := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true}
	if !allowed[ext] {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "message": "只支持 jpg/png/gif/webp 格式"})
		return
	}

	// 3. 校验文件大小（最大 5MB）
	if header.Size > 5*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "message": "图片大小不能超过 5MB"})
		return
	}

	// 4. 上传到 Cloudinary
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	publicID := fmt.Sprintf("patpet/avatars/user_%d", userID)

	result, err := h.Cloudinary.Upload.Upload(ctx, file, uploader.UploadParams{
		PublicID:       publicID,
		Folder:         "",
		Overwrite:      boolPtr(true),
		Transformation: "c_fill,w_400,h_400,q_80", // 裁剪为 400x400，质量 80%
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "message": "图片上传失败: " + err.Error()})
		return
	}

	// 5. 更新数据库中的头像 URL
	avatarURL := result.SecureURL
	if err := h.DB.Model(&model.User{}).Where("id = ?", userID).Update("avatar", avatarURL).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "message": "更新头像失败"})
		return
	}

	// 6. 返回结果
	var user model.User
	h.DB.First(&user, userID)

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "头像上传成功",
		"data":    user,
	})
}

func boolPtr(b bool) *bool {
	return &b
}
