package gin_gonic_http_srv

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"k8s.io/klog/v2"
	"net/http"
	"time"
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
	//err := i.router.Run(":80")
	//if err != nil {
	//	klog.Errorf("ERROR: gin-gonic run error %s", err.Error())
	//}

	//go func() {
	//	for {
	//		select {
	//		case <-stopCh:
	//			return
	//		default:
	//			err := i.router.Run(":8080")
	//			if err != nil {
	//				klog.Errorf("ERROR: gin-gonic run error %s", err.Error())
	//			}
	//		}
	//	}
	//}()
	//<-stopCh

	srv := &http.Server{
		Addr:    ":80",
		Handler: i.router,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			klog.Errorf("ERROR: import-location http srv error: %s", err.Error())
		}
	}()
	<-stopCh
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		klog.Errorf("Error: import-location http srv shutdown error: %s", err.Error())
	}
}

func (i *ImportLocationServer) handleCatalogInfoGet(c *gin.Context) {
	c.Data(http.StatusOK, "Content-Type: application/json", i.content)
}
