package server

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"k8s.io/klog/v2"
	"net/http"
)

type ImportLocationServer struct {
	router  *gin.Engine
	content []byte
}

func NewImportLocationServer(content []byte) *ImportLocationServer {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	i := &ImportLocationServer{
		router:  r,
		content: content,
	}
	r.SetTrustedProxies(nil)
	r.TrustedPlatform = "X-Forwarded-For"
	r.Use(addRequestId())
	r.GET("/", i.handleCatalogInfoGet)
	r.GET("/catalog-info.yaml", i.handleCatalogInfoGet)
	return i
}

// Middleware adding request ID to gin context.
// Note that this is a simple unique ID that can be used for debugging purposes.
// In the future, this might be replaced with OpenTelemetry IDs/tooling.
func addRequestId() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("requestId", uuid.New().String())
		c.Next()
	}
}

func (i *ImportLocationServer) Run(stopCh <-chan struct{}) {
	ch := make(chan int)
	go func() {
		for {
			select {
			case <-ch:
				return
			default:
				err := i.router.Run(":8080")
				if err != nil {
					klog.Errorf("ERROR: gin-gonic run error %s", err.Error())
				}
			}
		}
	}()
	<-stopCh
	close(ch)
}

func (i *ImportLocationServer) handleCatalogInfoGet(c *gin.Context) {
	c.Data(http.StatusOK, "Content-Type: application/json", i.content)
}
