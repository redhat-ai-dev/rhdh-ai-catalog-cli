package server

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"k8s.io/klog/v2"
	"net/http"
	"strings"
)

type ImportLocationServer struct {
	router  *gin.Engine
	content map[string][]byte
}

func NewImportLocationServer(content map[string][]byte) *ImportLocationServer {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	i := &ImportLocationServer{
		router:  r,
		content: content,
	}
	klog.Infof("NewImportLocationServer content len %d", len(content))
	r.SetTrustedProxies(nil)
	r.TrustedPlatform = "X-Forwarded-For"
	r.Use(addRequestId())
	d := &DicoveryResponse{Uris: []string{}}
	for key, data := range content {
		klog.Infof("NewImportLocationServer looking at key %s and content len %d", key, len(data))
		il := &ImportLocation{content: data}
		segs := strings.Split(key, "_")
		if len(segs) < 2 {
			continue
		}
		uri := fmt.Sprintf("%s/%s/catalog-info.yaml", segs[0], segs[1])
		klog.Infoln("Adding URI " + uri)
		r.GET(uri, il.handleCatalogInfoGet)
		d.Uris = append(d.Uris, uri)
	}
	r.GET("/list", d.handleCatalogDiscoveryGet)
	//TODO can also provide a POST URI for adding ImportLocations
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

type ImportLocation struct {
	content []byte
}

func (i *ImportLocation) handleCatalogInfoGet(c *gin.Context) {
	c.Data(http.StatusOK, "Content-Type: application/json", i.content)
}

type DicoveryResponse struct {
	Uris []string `json:"uris"`
}

func (d *DicoveryResponse) handleCatalogDiscoveryGet(c *gin.Context) {
	content, err := json.Marshal(d)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		c.Error(err)
		return
	}
	c.Data(http.StatusOK, "Content-Type: application/json", content)
}
