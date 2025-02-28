package storage

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/redhat-ai-dev/rhdh-ai-catalog-cli/pkg/cmd/cli/backstage"
	bridgeclient "github.com/redhat-ai-dev/rhdh-ai-catalog-cli/pkg/cmd/server/location/client"
	"github.com/redhat-ai-dev/rhdh-ai-catalog-cli/pkg/config"
	"github.com/redhat-ai-dev/rhdh-ai-catalog-cli/pkg/rest"
	"github.com/redhat-ai-dev/rhdh-ai-catalog-cli/pkg/types"
	"github.com/redhat-ai-dev/rhdh-ai-catalog-cli/pkg/util"
	"k8s.io/klog/v2"
	"net/http"
	"strings"
	"sync"
)

type StorageRESTServer struct {
	router          *gin.Engine
	st              types.BridgeStorage
	pushedLocations sync.Map
	locations       *bridgeclient.BridgeLocationRESTClient
	bkstg           rest.BackstageImport
}

func NewStorageRESTServer(st types.BridgeStorage, bridgeURL, bridgeToken, bkstgURL, bkstgToken string) *StorageRESTServer {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	cfg := &config.Config{
		BackstageURL:     bkstgURL,
		BackstageToken:   bkstgToken,
		BackstageSkipTLS: true,
	}
	s := &StorageRESTServer{
		router:          r,
		st:              st,
		pushedLocations: sync.Map{},
		locations:       bridgeclient.SetupBridgeLocationRESTClient(bridgeURL, bridgeToken),
		bkstg:           backstage.SetupBackstageRESTClient(cfg),
	}
	klog.Infof("NewStorageRESTServer")
	r.SetTrustedProxies(nil)
	r.TrustedPlatform = "X-Forwarded-For"
	r.Use(addRequestId())
	r.POST("/upsert", s.handleCatalogUpsertPost)
	return s
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

func (s *StorageRESTServer) Run(stopCh <-chan struct{}) {
	ch := make(chan int)
	go func() {
		for {
			select {
			case <-ch:
				return
			default:
				err := s.router.Run(":7070")
				if err != nil {
					klog.Errorf("ERROR: gin-gonic run error %s", err.Error())
				}
			}
		}
	}()
	<-stopCh
	close(ch)
}

func (s *StorageRESTServer) handleCatalogUpsertPost(c *gin.Context) {
	key := c.Query("key")
	if len(key) == 0 {
		c.Status(http.StatusBadRequest)
		c.Error(fmt.Errorf("need a 'key' parameter"))
		return
	}
	var postBody rest.PostBody
	err := c.BindJSON(&postBody)
	if err != nil {
		c.Status(http.StatusBadRequest)
		msg := fmt.Sprintf("error reading POST body: %s", err.Error())
		klog.Errorf(msg)
		c.Error(fmt.Errorf(msg))
		return
	}
	segs := strings.Split(key, "_")
	if len(segs) < 2 {
		c.Status(http.StatusBadRequest)
		c.Error(fmt.Errorf("bad key format: %s", key))
		return
	}
	key, uri := util.BuildImportKeyAndURI(segs[0], segs[1])
	klog.Infof("Upserting URI %s with key %s with data of len %d", uri, key, len(postBody.Body))

	// push to storage, but just the byte array
	sb := types.StorageBody{
		Body:           postBody.Body,
		LocationId:     "",
		LocationTarget: "",
	}
	err = s.st.Upsert(key, sb)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		msg := fmt.Sprintf("error upserting to storage key %s POST body: %s", key, err.Error())
		klog.Errorf(msg)
		c.Error(fmt.Errorf(msg))
		return
	}

	// push update to bridge locations REST endpoint
	var rc int
	var msg string
	rc, msg, _, err = s.locations.UpsertModel(key, postBody.Body)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		msg = fmt.Sprintf("error upserting to bridge uri %s POST body: msg %s error %s", uri, msg, err.Error())
		klog.Errorf(msg)
		c.Error(fmt.Errorf(msg))
		return
	}
	if rc != http.StatusCreated && rc != http.StatusOK {
		c.Status(rc)
		msg = fmt.Sprintf("error upserting to bridge uri %s POST body: msg %s", uri, msg)
		klog.Errorf(msg)
		c.Error(fmt.Errorf(msg))
	}

	// if we have not previously pushed to backstage, do so now;
	// we use a sync map here in case our store implementation does not provide atomic updates
	_, alreadyPushed := s.pushedLocations.LoadOrStore(uri, uri)
	if alreadyPushed {
		klog.Info(fmt.Sprintf("%s already provides location %s", s.locations.UpsertURL, uri))
		c.Status(http.StatusOK)
		return
	}

	impResp := map[string]any{}
	impResp, err = s.bkstg.ImportLocation(s.locations.HostURL + uri)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		msg = fmt.Sprintf("error importing location %s to backstage: %s", s.locations.HostURL+uri, err.Error())
		klog.Errorf(msg)
		c.Error(fmt.Errorf(msg))
		return
	}
	retID, retTarget, rok := rest.ParseImportLocationMap(impResp)
	if !rok {
		c.Status(http.StatusBadRequest)
		msg = fmt.Sprintf("parsing of import location return had an issue: %#v", impResp)
		klog.Errorf(msg)
		c.Error(fmt.Errorf(msg))
		return
	}

	// finally store in our storage layer with the id and cross reference location URL from backstage
	sb.LocationId = retID
	sb.LocationTarget = retTarget
	err = s.st.Upsert(key, sb)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		msg = fmt.Sprintf("error upserting to storage key %s POST body plus backstage ID: %s", key, err.Error())
		klog.Errorf(msg)
		c.Error(fmt.Errorf(msg))
		return
	}

	c.Status(http.StatusCreated)
}
