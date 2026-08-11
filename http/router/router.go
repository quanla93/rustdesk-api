package router

import (
	"github.com/gin-gonic/gin"
	"github.com/lejianwen/rustdesk-api/v2/global"
	"github.com/lejianwen/rustdesk-api/v2/http/controller/web"
	"net/http"
)

func WebInit(g *gin.Engine) {
	i := &web.Index{}
	g.GET("/", i.Index)

	if global.Config.App.WebClient == 1 {
		g.GET("/webclient-config/index.js", i.ConfigJs)
	}

	if global.Config.App.WebClient == 1 {
		g.StaticFS("/webclient", http.Dir(global.Config.Gin.ResourcesPath+"/web"))
		g.GET("/web", redirectWebClient)
		g.GET("/web/*filepath", redirectWebClient)
		g.GET("/webclient2", redirectWebClient)
		g.GET("/webclient2/*filepath", redirectWebClient)
	}
	g.StaticFS("/_admin", http.Dir(global.Config.Gin.ResourcesPath+"/admin"))
}

func redirectWebClient(c *gin.Context) {
	path := c.Param("filepath")
	if path == "" || path == "/" {
		path = "/"
	}
	c.Redirect(http.StatusMovedPermanently, "/webclient"+path)
}
