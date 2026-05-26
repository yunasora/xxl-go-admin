// 文件位置：handlers/registry.go
package handlers

import (
	"bytes"
	"go-xxl-admin/core"
	"go-xxl-admin/global"
	"go-xxl-admin/models"
	"net/http"
	"time"

	"gorm.io/gorm/clause"

	"github.com/gin-gonic/gin"
)

func HandlerRegistry(c *gin.Context) {
	var param models.RegistryParam
	if err := c.ShouldBindJSON(&param); err != nil {
		c.JSON(http.StatusOK, models.XxlResponse{Code: 500, Msg: "Invalid Message"})
		return
	}
	if param.RegistryKey == "" || param.RegistryValue == "" {
		c.JSON(http.StatusOK, models.XxlResponse{Code: 500, Msg: "Key or Value empty"})
		return
	}

	addr := param.RegistryValue
	if !bytes.HasSuffix([]byte(addr), []byte("/")) {
		addr = addr + "/"
	}

	registryRecord := models.JobRegistry{
		RegistryGroup: param.RegistryKey,
		RegistryKey:   param.RegistryKey,
		RegistryValue: addr,
		UpdateTime:    time.Now(),
	}

	err := global.DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "registry_value"}},
		DoUpdates: clause.AssignmentColumns([]string{"update_time"}),
	}).Create(&registryRecord).Error

	if err != nil {
		c.JSON(http.StatusOK, models.XxlResponse{Code: 500, Msg: "Database save registry failed"})
		return
	}

	core.RegsC.StoreNode(param.RegistryKey, addr)
	c.JSON(http.StatusOK, models.XxlResponse{Code: 200})
}
