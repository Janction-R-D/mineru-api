package pkg

import (
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

type ExtractTaskRequest struct {
	Url string `json:"url"`
}

func ExtractTask(c *gin.Context) {
	var request ExtractTaskRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp := ExecExtractTask(request)
	c.JSON(200, resp)
}

func GetExtractTask(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}
	resp, mdPath, fname := GetExtractTaskDetail(id)
	if mdPath != "" {
		c.Header("Content-Type", "application/octet-stream")
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", url.QueryEscape(fname)))
		c.Header("Content-Transfer-Encoding", "binary")
		c.File(mdPath)
		return
	}
	c.JSON(200, resp)
}

func GetFileNameWithoutExt(fileName string) string {
	base := filepath.Base(fileName)
	ext := filepath.Ext(base)
	return base[:len(base)-len(ext)]
}
