// 文件位置：handlers/registry.go
package handlers

import (
	"bytes"
	"github.com/gin-gonic/gin"
	"go-xxl-admin/core"
	"go-xxl-admin/models"
	"net/http"
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

	core.RegsC.StoreNode(param.RegistryKey, addr)
	c.JSON(http.StatusOK, models.XxlResponse{Code: 200})
}
