package handler

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"patpet-server/model"
)

// UploadHandler 文件上传
type UploadHandler struct {
	DB             *gorm.DB
	SupabaseURL    string
	SupabaseKey    string
	SupabaseBucket string
}

// UploadAvatar 上传头像
func (h *UploadHandler) UploadAvatar(c *gin.Context) {
	userID := c.GetUint("user_id")

	if h.SupabaseURL == "" || h.SupabaseKey == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "message": "未配置图片存储服务"})
		return
	}

	// 1. 从请求中获取文件
	file, header, err := c.Request.FormFile("avatar")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "message": "请选择要上传的图片"})
		return
	}
	defer file.Close()

	// 2. 校验文件类型
	ext := strings.ToLower(filepath.Ext(header.Filename))
	contentTypes := map[string]string{
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".gif":  "image/gif",
		".webp": "image/webp",
	}
	contentType, ok := contentTypes[ext]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "message": "只支持 jpg/png/gif/webp 格式"})
		return
	}

	// 3. 校验文件大小（最大 5MB）
	if header.Size > 5*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "message": "图片大小不能超过 5MB"})
		return
	}

	// 4. 读取文件内容
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "message": "读取图片失败"})
		return
	}

	// 5. 上传到 Supabase Storage
	// 文件路径：avatars/user_1.jpg（每个用户固定一个文件名，覆盖更新）
	objectPath := fmt.Sprintf("user_%d%s", userID, ext)
	uploadURL := fmt.Sprintf("%s/storage/v1/object/%s/%s", h.SupabaseURL, h.SupabaseBucket, objectPath)

	req, err := http.NewRequest("POST", uploadURL, bytes.NewReader(fileBytes))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "message": "创建上传请求失败"})
		return
	}
	req.Header.Set("Authorization", "Bearer "+h.SupabaseKey)
	req.Header.Set("apikey", h.SupabaseKey)
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("x-upsert", "true") // 覆盖已有文件

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "message": "上传图片失败: " + err.Error()})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "message": "上传失败: " + string(body)})
		return
	}

	// 6. 拼接公开访问 URL
	avatarURL := fmt.Sprintf("%s/storage/v1/object/public/%s/%s", h.SupabaseURL, h.SupabaseBucket, objectPath)

	// 7. 更新数据库中的头像 URL
	if err := h.DB.Model(&model.User{}).Where("id = ?", userID).Update("avatar", avatarURL).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "message": "更新头像失败"})
		return
	}

	// 8. 返回结果
	var user model.User
	h.DB.First(&user, userID)

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "头像上传成功",
		"data":    user,
	})
}
