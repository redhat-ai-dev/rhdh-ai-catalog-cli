package cli

import (
	"context"
	routev1 "github.com/openshift/api/route/v1"
	"github.com/redhat-ai-dev/rhdh-ai-catalog-cli/pkg/cmd/cli/backstage"
	"github.com/redhat-ai-dev/rhdh-ai-catalog-cli/pkg/cmd/cli/kserve"
	"github.com/redhat-ai-dev/rhdh-ai-catalog-cli/pkg/cmd/cli/kubeflowmodelregistry"
	"github.com/redhat-ai-dev/rhdh-ai-catalog-cli/pkg/config"
	"github.com/redhat-ai-dev/rhdh-ai-catalog-cli/pkg/util"
	"github.com/spf13/cobra"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/klog/v2"
	"net/url"
	"os"
	"strconv"
	"strings"
)

const (
	bkstgAIExample = `
# The 'new-model' command will access a supported backend for AI Model metadata: 
# - kserve, for inspecting active Kserve Inferences Services in a Kubernetes cluster
# - kubeflow, for querying a Kubeflow Model Registry instance for Model information
# 
# and from the data retrieved from those sources, produce YAML formatted output that corresponds
# to the Backstage catalog entities:
# - Components
# - Resources
# - APIs
# 
# After possibly reviewing the output to the screen, the user will (re)run the command and redirect it
# to a 'catalog-info.yaml' file and push the contents of that file to an HTTP accessible location (most likely
# a Git repository.  Afterward, use if 'import-model' will complete the creation flow.
$ %s new-model <kserve|kubeflow> <owner> <lifecycle> <args...>

# The 'import-model' command takes the 'catalog-info.yaml' file produced by 'new-model', and stored in an HTTP accessible 
# location (where the <url> parameter is the retrieval address for the file), and imports the contents of the 
# 'catalog-info.yaml' file into the Catalog of a running Backstage instance.
$ %s import-model <url>

# The 'get' command allows for the retrieval of YAML formatted representations of various entities from the Backstage Catalog.
$ %s get [location|components|resources|apis|entities] [args...]

# The 'delete-model' command will remove the Backstage Catalog Location entity for the provided Location ID, which in turn
# will remove any Components, Resources, or APIs imported from that location.  The <location id> is a generated key that
# is associated with the URL provided when the user runs the 'import-model' command.  You also can see this ID when you
# view the locations from the Backstage UI.
$ %s delete-model <location id>
`

	newModelExample = `
# Access a supported backend for AI Model metadata and generate Backstage Catalog Entity YAML for that metadata
$ %s new-model kserve [args]
`

	getExample = `
# Access the Backstage Catalog for Entities related to AI Models
$ %s get <locations|components|resources|apis|entities> [args...]
`

	deleteModelExample = `
# Remove from the Backstage Catalog the Location entity for the provided Location ID, using the dynamically generated 
# hash ID from when the location was imported.  There is not support in Backstage currently for specifying
# the URL used to import the model as a query parameter.
$ %s delete-model <location id>

# Set the URL for the Backstage instance, the authentication token, and Skip-TLS settings 
$ %s delete-model <location id> --backstage-url=https://my-rhdh.com --backstage-token=my-token --backstage-skip-tls=true
`

	importModelExample = `
# Import from an accessible URL Backstage Catalog entities
$ %s import-model <url>

# Set the additional URL for the Backstage instance, the authentication token, and Skip-TLS settings 
$ %s import-model <url> --backstage-url=https://my-rhdh.com --backstage-token=my-token --backstage-skip-tls=true
`

	getEntitiesExample = `
# Access the Backstage Catalog for all entities, regardless if AI related
$ %s get entities

# Set the URL for the Backstage, the authentication token, and Skip-TLS settings
$ %s get entities --backstage-url=https://my-rhdh.com --backstage-token=my-token --backstage-skip-tls=true
`

	getLocationsExample = `
# Access the Backstage Catalog for locations, regardless if AI related
$ %s get locations [args...]

# Access the Backstage Catatlog for a specific location using the dynamically generated 
# hash ID from when the location was imported.  There is not support in Backstage currently for specifying
# the URL used to import the model as a query parameter.
$ %s get locations my-big-long-id-for-location

# Set the URL for the Backstage, the authentication token, and Skip-TLS settings
$ %s get locations --backstage-url=https://my-rhdh.com --backstage-token=my-token --backstage-skip-tls=true
`

	getComponentsExample = `
# Retrieve the Backstage Catalog for resources related to AI Models, where being AI related is determined by the 
# 'type' being set to 'model-server'
$ %s get components [args...]

# Set the URL for the Backstage, the authentication token, and Skip-TLS settings
$ %s get components --backstage-url=https://my-rhdh.com --backstage-token=my-token --backstage-skip-tls=true

# Retrieve a specific set of AI related Components by namespace:name
$ %s get components default:my-component default:your-component

# Retrieve a set of AI Components where the provided list of tags match (order of tags disregarded)
$ %s get components genai vllm --use-params-as-tags=true

# Retrieve a set of Components which have any of the provided list of tags
$ %s get components gen-ai --use-params-as-tags=true --use-any-subset=true
`

	getResourcesExample = `
# Retrieve the Backstage Catalog for resources related to AI Models, where being AI related is determined by the 
# 'type' being set to 'ai-model'
$ %s get resources [args...]

# Set the URL for the Backstage, the authentication token, and Skip-TLS settings
$ %s get resources --backstage-url=https://my-rhdh.com --backstage-token=my-token --backstage-skip-tls=true

# Retrieve a specific set of AI related Resources by namespace:name
$ %s get resources default:my-component default:your-component

# Retrieve a set of AI Resources where the provided list of tags match (order of tags disregarded)
$ %s get resources genai vllm --use-params-as-tags=true

# Retrieve a set of AI Resources which have any of the provided list of tags
$ %s get resources gen-ai --use-params-as-tags=true --use-any-subset=true
`

	getApisExample = `
# Retrieve the Backstage Catalog for APIs related to AI Models, where being AI related is determined by the 
# 'type' being set to 'model-service-api'
$ %s get apis [args...]

# Set the URL for the Backstage, the authentication token, and Skip-TLS settings
$ %s get locations --backstage-url=https://my-rhdh.com --backstage-token=my-token --backstage-skip-tls=true

# Retrieve a specific set of AI related APIs by namespace:name
$ %s get apis default:my-component default:your-component

# Retrieve a set of AI APIs where the provided list of tags match (order of tags disregarded)
$ %s get apis genai vllm --use-params-as-tags=true

# Retrieve a set of AI APIs which have any of the provided list of tags
$ %s get apis gen-ai --use-params-as-tags=true --use-any-subset=true
`
)

// NewCmd create a new root command, linking together all sub-commands organized by groups.
func NewCmd() *cobra.Command {
	cfg := &config.Config{}
	bkstgAI := &cobra.Command{
		Use:     util.ApplicationName,
		Long:    "Backstage AI is a command line tool that facilitates management of AI related Entities in the Backstage Catalog.",
		Example: strings.ReplaceAll(bkstgAIExample, "%s", util.ApplicationName),
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	cfg.Kubeconfig = os.Getenv("KUBECONFIG")
	cfg.BackstageURL = os.Getenv("BACKSTAGE_URL")
	cfg.BackstageToken = os.Getenv("BACKSTAGE_TOKEN")
	cfg.BackstageSkipTLS, _ = strconv.ParseBool(os.Getenv("BACKSTAGE_SKIP_TLS"))
	cfg.StoreURL = os.Getenv("MODEL_METADATA_URL")
	cfg.StoreToken = os.Getenv("MODEL_METADATA_TOKEN")
	cfg.StoreSkipTLS, _ = strconv.ParseBool(os.Getenv("METADATA_MODEL_SKIP_TLS"))
	cfg.Namespace = util.GetCurrentProject()

	bkstgAI.PersistentFlags().StringVar(&(cfg.Kubeconfig), "kubeconfig", cfg.Kubeconfig,
		"Path to the kubeconfig file to use for CLI requests.")
	bkstgAI.PersistentFlags().StringVar(&(cfg.Namespace), "namespace", cfg.Namespace,
		"The name of the Kubernetes namespace to use for CLI requests.")
	bkstgAI.PersistentFlags().StringVar(&(cfg.BackstageURL), "backstage-url", cfg.BackstageURL,
		"The URL used for accessing the Backstage Catalog REST API.")
	bkstgAI.PersistentFlags().StringVar(&(cfg.BackstageToken), "backstage-token", cfg.BackstageToken,
		"The bearer authorization token used for accessing the Backstage Catalog REST API.")
	bkstgAI.PersistentFlags().BoolVar(&(cfg.BackstageSkipTLS), "backstage-skip-tls", cfg.StoreSkipTLS,
		"Whether to skip use of TLS when accessing the Backstage Catalog REST API.")
	bkstgAI.PersistentFlags().StringVar(&(cfg.StoreURL), "model-metadata-url", cfg.StoreURL,
		"The URL used for accessing the external source for Model Metadata.")
	bkstgAI.PersistentFlags().StringVar(&(cfg.StoreToken), "model-metadata-token", cfg.StoreToken,
		"The bearer authorization token used for accessing the external source for Model Metadata.")
	bkstgAI.PersistentFlags().BoolVar(&(cfg.StoreSkipTLS), "model-metadata-skip-tls", cfg.StoreSkipTLS,
		"Whether to skip use of TLS when accessing the external source for Model Metadata.")

	newModel := &cobra.Command{
		Use:     "new-model",
		Long:    "new-model accesses one of the supported backends and builds Backstage Catalog Entity YAML with available Model metadata",
		Aliases: []string{"create", "c", "nm", "new-models", "export-model"},
		Example: strings.ReplaceAll(newModelExample, "%s", util.ApplicationName),
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	newModel.AddCommand(kserve.NewCmd(cfg))
	newModel.AddCommand(kubeflowmodelregistry.NewCmd(cfg))

	queryModel := &cobra.Command{
		Use:     "get",
		Long:    "get accesses the Backstage Catalog for Entities related to AI Models",
		Aliases: []string{"g"},
		Example: strings.ReplaceAll(getExample, "%s", util.ApplicationName),
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}
	deleteModel := &cobra.Command{
		Use:     "delete-model",
		Long:    "delete-model removes the Backstage Catalog for Entities corresponding to the provided location ID",
		Aliases: []string{"delete", "dm", "del", "d", "delete-models"},
		Example: strings.ReplaceAll(deleteModelExample, "%s", util.ApplicationName),
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				klog.Error("ERROR: delete-model requires a location ID")
			}
			util.ProcessOutput(backstage.SetupBackstageRESTClient(cfg).DeleteLocation(args[0]))
		},
	}
	importModel := &cobra.Command{
		Use:     "import-model",
		Long:    "import-model updates the Backstage Catalog with Entities contained in the provided location URL",
		Aliases: []string{"post", "im", "p", "i", "import-models"},
		Example: strings.ReplaceAll(importModelExample, "%s", util.ApplicationName),
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				klog.Error("ERROR: import-model requires a location URL")
				klog.Flush()
				return
			}
			u, uerr := url.Parse(args[0])
			if uerr != nil {
				klog.Errorf("ERROR: import-model requires a valid location URL: %s", uerr.Error())
				klog.Flush()
				return
			}
			switch u.Scheme {
			case "http":
				fallthrough
			case "https":
				util.ProcessOutput(backstage.SetupBackstageRESTClient(cfg).ImportLocation(args[0]))
				return
			default:
				filePath := u.Path
				var content []byte
				var fileErr error
				content, fileErr = os.ReadFile(filePath)
				if fileErr != nil {
					klog.Errorf("ERROR: import-model problem reading file %s: %s", filePath, fileErr.Error())
					klog.Flush()
					return
				}
				restCfg, err := util.GetK8sConfig(cfg)
				if err != nil {
					klog.Errorf("ERROR: import-model problem getting k8s rest config: %s", err.Error())
					klog.Flush()
					return
				}
				ctx := context.Background()
				coreClient := util.GetCoreClient(restCfg)
				cm := &corev1.ConfigMap{}
				cm.Namespace = cfg.Namespace
				cm.ObjectMeta.GenerateName = "bac-import-model-"
				cm.BinaryData = map[string][]byte{}
				cm.BinaryData["catalog-info-yaml"] = content
				cm, err = coreClient.ConfigMaps(cfg.Namespace).Create(ctx, cm, metav1.CreateOptions{})
				if err != nil {
					klog.Errorf("ERROR: import-model problem creating configmap: %s", err.Error())
					klog.Flush()
					return
				}
				svc := &corev1.Service{}
				svc.Namespace = cfg.Namespace
				svc.Name = cm.Name
				svc.ObjectMeta.Labels = map[string]string{}
				svc.ObjectMeta.Labels["app"] = cm.Name
				svc.Spec.Selector = map[string]string{}
				svc.Spec.Selector["app"] = cm.Name
				svc.Spec.Ports = []corev1.ServicePort{
					{
						Name:     "location",
						Protocol: corev1.ProtocolTCP,
						Port:     8080,
						//TargetPort: intstr.FromString("location"),
						TargetPort: intstr.FromInt32(8080),
					},
				}
				svc, err = coreClient.Services(cfg.Namespace).Create(ctx, svc, metav1.CreateOptions{})
				if err != nil {
					klog.Errorf("ERROR: import-model problem creating service: %s", err.Error())
					klog.Flush()
					return
				}
				route := &routev1.Route{}
				route.Namespace = cfg.Namespace
				route.Name = cm.Name
				route.Spec = routev1.RouteSpec{
					To: routev1.RouteTargetReference{Kind: "Service", Name: svc.Name},
					//Port: &routev1.RoutePort{TargetPort: intstr.FromInt32(80)},
					//Port: &routev1.RoutePort{TargetPort: intstr.FromString("location")},
					//TLS:  &routev1.TLSConfig{Termination: routev1.TLSTerminationEdge, InsecureEdgeTerminationPolicy: routev1.InsecureEdgeTerminationPolicyAllow},
				}
				routeClient := util.GetRouteClient(restCfg)
				route, err = routeClient.Routes(cfg.Namespace).Create(ctx, route, metav1.CreateOptions{})
				if err != nil {
					klog.Errorf("ERROR: import-model problem creating route: %s", err.Error())
					klog.Flush()
					return
				}
				deployment := &appv1.Deployment{}
				deployment.Namespace = cfg.Namespace
				deployment.Name = cm.Name
				deployment.ObjectMeta.Labels = map[string]string{}
				deployment.ObjectMeta.Labels["app.kubernetes.io/name"] = cm.Name
				replicas := int32(1)
				defaultMode := int32(420)
				//pid0 := int64(0)
				readOnlyFSnonRoot := true
				deployment.Spec = appv1.DeploymentSpec{
					Replicas: &replicas,
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"app": cm.Name},
					},
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{"app": cm.Name},
						},
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:  "location",
									Image: "quay.io/gabemontero/import-location:v2",
									Ports: []corev1.ContainerPort{
										{
											Name:          "location",
											ContainerPort: 8080,
										},
									},
									Resources: corev1.ResourceRequirements{
										Limits: corev1.ResourceList{
											corev1.ResourceCPU:    resource.Quantity{Format: resource.Format("500m")},
											corev1.ResourceMemory: resource.Quantity{Format: resource.Format("384Mi")},
										},
										Requests: corev1.ResourceList{
											corev1.ResourceCPU:    resource.Quantity{Format: resource.Format("5m")},
											corev1.ResourceMemory: resource.Quantity{Format: resource.Format("64Mi")},
										},
									},
									SecurityContext: &corev1.SecurityContext{ReadOnlyRootFilesystem: &readOnlyFSnonRoot},
									VolumeMounts: []corev1.VolumeMount{
										{
											Name:      "location",
											MountPath: "/data",
											ReadOnly:  true,
										},
									},
								},
							},
							SecurityContext: &corev1.PodSecurityContext{RunAsNonRoot: &readOnlyFSnonRoot},
							Volumes: []corev1.Volume{
								{
									Name: "location",
									VolumeSource: corev1.VolumeSource{ConfigMap: &corev1.ConfigMapVolumeSource{
										LocalObjectReference: corev1.LocalObjectReference{Name: cm.Name},
										DefaultMode:          &defaultMode,
									}},
								},
							},
						},
					},
					Strategy: appv1.DeploymentStrategy{},
				}
				appClient := util.GetAppClient(restCfg)
				deployment, err = appClient.Deployments(cfg.Namespace).Create(ctx, deployment, metav1.CreateOptions{})
				if err != nil {
					klog.Errorf("ERROR: import-model problem creating deployment: %s", err.Error())
					klog.Flush()
					return
				}
				watchTimeOut := int64(120)
				routeWatch, routeWatchErr := routeClient.Routes(cfg.Namespace).Watch(ctx, metav1.ListOptions{TimeoutSeconds: &watchTimeOut})
				if routeWatchErr != nil {
					klog.Errorf("ERROR: import-model problem creating route watch: %s", routeWatchErr.Error())
					klog.Flush()
					return
				}
				routeURL := ""
				for event := range routeWatch.ResultChan() {
					item := event.Object.(*routev1.Route)
					if item.Name != route.Name {
						continue
					}

					switch event.Type {
					case watch.Error:
						fallthrough
					case watch.Deleted:
						routeWatch.Stop()
						break
					}

					if len(item.Status.Ingress) == 0 {
						continue
					}
					if len(item.Status.Ingress[0].Conditions) == 0 {
						continue
					}

					if item.Status.Ingress[0].Conditions[0].Status == corev1.ConditionTrue {
						routeURL = "http://" + item.Status.Ingress[0].Host
						break
					}

				}
				util.ProcessOutput(backstage.SetupBackstageRESTClient(cfg).ImportLocation(routeURL))
			}

		},
	}

	bkstgAI.AddCommand(newModel)
	bkstgAI.AddCommand(queryModel)
	bkstgAI.AddCommand(deleteModel)
	bkstgAI.AddCommand(importModel)

	queryModel.AddCommand(&cobra.Command{
		Use:     "entities",
		Long:    "entities retrieves the AI related Backstage Catalog Entities",
		Aliases: []string{"e", "entity"},
		Example: strings.ReplaceAll(getEntitiesExample, "%s", util.ApplicationName),
		RunE: func(cmd *cobra.Command, args []string) error {

			str, err := backstage.SetupBackstageRESTClient(cfg).ListEntities()
			util.ProcessOutput(str, err)
			return err

		},
	})

	queryModel.AddCommand(&cobra.Command{
		Use:     "locations",
		Long:    "locations retrieves the AI related Backstage Catalog Locations",
		Aliases: []string{"l", "location"},
		Example: strings.ReplaceAll(getLocationsExample, "%s", util.ApplicationName),
		RunE: func(cmd *cobra.Command, args []string) error {
			str, err := backstage.SetupBackstageRESTClient(cfg).GetLocation(args...)
			util.ProcessOutput(str, err)
			return err
		},
	})

	queryModel.AddCommand(&cobra.Command{
		Use:     "components",
		Long:    "components retrieves the AI related Backstage Catalog Components",
		Aliases: []string{"c", "component"},
		Example: strings.ReplaceAll(getComponentsExample, "%s", util.ApplicationName),
		RunE: func(cmd *cobra.Command, args []string) error {
			str, err := backstage.SetupBackstageRESTClient(cfg).GetComponent(args...)
			util.ProcessOutput(str, err)
			return err
		},
	})

	queryModel.AddCommand(&cobra.Command{
		Use:     "resources",
		Long:    "resources retrieves the AI related Backstage Catalog Resources",
		Aliases: []string{"r", "resource"},
		Example: strings.ReplaceAll(getResourcesExample, "%s", util.ApplicationName),
		RunE: func(cmd *cobra.Command, args []string) error {
			str, err := backstage.SetupBackstageRESTClient(cfg).GetResource(args...)
			util.ProcessOutput(str, err)
			return err
		},
	})

	queryModel.AddCommand(&cobra.Command{
		Use:     "apis",
		Long:    "apis retrieves the AI related Backstage Catalog APIS",
		Aliases: []string{"a", "api"},
		Example: strings.ReplaceAll(getApisExample, "%s", util.ApplicationName),
		RunE: func(cmd *cobra.Command, args []string) error {
			str, err := backstage.SetupBackstageRESTClient(cfg).GetAPI(args...)
			util.ProcessOutput(str, err)
			return err
		},
	})

	queryModel.PersistentFlags().BoolVar(&(cfg.ParamsAsTags), "use-params-as-tags", cfg.ParamsAsTags,
		"Use any additional parameters as tag identifiers")
	queryModel.PersistentFlags().BoolVar(&(cfg.AnySubsetWorks), "allow-tags-subset", cfg.AnySubsetWorks,
		"When set with 'use-params-as-tags', this just requires the tags provided to be set, but allows for additional tags to be set")

	return bkstgAI
}
