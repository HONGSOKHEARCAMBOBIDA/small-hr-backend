package controller

import (
	"mysql/backup"
	"mysql/constant/share"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

type BackupController struct{}

func NewBackupController() *BackupController {
	return &BackupController{}
}

func (bc *BackupController) TriggerBackup(c *gin.Context) {
	result, err := backup.Run()
	if err != nil {
		share.ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"filename":   result.Filename,
		"size_bytes": result.SizeBytes,
		"created_at": result.CreatedAt,
	})
}

func (bc *BackupController) ListBackups(c *gin.Context) {
	entries, err := os.ReadDir("./backups")
	if err != nil {
		share.ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	var files []gin.H
	for _, e := range entries {
		if !e.IsDir() {
			info, _ := e.Info()
			files = append(files, gin.H{
				"filename":   e.Name(),
				"size_bytes": info.Size(),
				"created_at": info.ModTime(),
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{"backups": files})
}

func (bc *BackupController) DownloadBackup(c *gin.Context) {
	filename := c.Query("file")
	if filename == "" {
		share.ResponseError(c, http.StatusBadRequest, "missing file param")
		return
	}
	filePath := "./backups/" + filename
	c.FileAttachment(filePath, filename)
}

func (bc *BackupController) DeleteBackup(c *gin.Context) {
	filename := c.Query("file")
	if filename == "" {
		share.ResponseError(c, http.StatusBadRequest, "missing file param")
		return
	}

	filePath := "./backups/" + filename
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		share.ResponseError(c, http.StatusNotFound, "file not found")
		return
	}

	if err := os.Remove(filePath); err != nil {
		share.ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"message":  "file deleted successfully",
		"filename": filename,
	})
}
