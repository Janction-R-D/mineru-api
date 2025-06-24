package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"mineruapi/pkg"

	"github.com/fvbock/endless"
	"github.com/gin-gonic/gin"
)

var port = "9000"
var debug = false
var token = ""
var clientUrl = ""

func init() {
	flag.BoolVar(&debug, "d", false, "debug mode")
	flag.StringVar(&token, "t", "123456789", "token")
	flag.StringVar(&clientUrl, "u", "http://127.0.0.1:30000", "clientUrl")
	flag.StringVar(&port, "p", "8888", "start port")
}

func main() {
	flag.Parse()
	fmt.Println("debug", debug)
	apiToken := os.Getenv("MINERU_API_TOKEN")
	if apiToken != "" {
		token = apiToken
	}
	apiPort := os.Getenv("MINERU_API_PORT")
	if apiPort != "" {
		port = apiPort
	}
	clientUrlE := os.Getenv("MINERU_CLIENT_URL")
	if clientUrlE != "" {
		clientUrl = clientUrlE
	}
	if !debug {
		gin.SetMode(gin.ReleaseMode)
	}

	go pkg.Init(clientUrl)
	g := initAPIEngine()

	endless.DefaultHammerTime = 5 * time.Second
	_ = endless.ListenAndServe(fmt.Sprintf(":%v", port), g)

}

func initAPIEngine() *gin.Engine {
	g := gin.New()
	g.HandleMethodNotAllowed = true
	g.RedirectTrailingSlash = false
	g.Use(gin.Recovery())

	g.NoRoute(func(c *gin.Context) {
		c.String(http.StatusNotFound, "The incorrect API route.")
	})

	g.NoMethod(func(c *gin.Context) {
		c.String(http.StatusForbidden, "The incorrect Method.")
	})

	g.GET("/health_check", func(c *gin.Context) {
		response := map[string]interface{}{
			"server_name": "mineru api",
			"client_ip":   c.ClientIP(),
			"header":      c.Request.Header,
		}
		c.JSON(200, response)
	})

	v1 := g.Group("/v1").Use(TokenCheck())
	{
		v1.POST("/extract/task", pkg.ExtractTask)
		v1.GET("/extract/task/:id", pkg.GetExtractTask)
	}

	return g
}

func TokenCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		authorization := c.Request.Header.Get("Authorization")
		tokenString := strings.Replace(authorization, "Bearer ", "", 1)
		if tokenString != token {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			c.Next()
			return
		}
	}
}
