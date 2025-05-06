// Code generated from JSON Schema using quicktype. DO NOT EDIT.
// To parse and unparse this JSON data, add this code to your project and do:
//
//    modelCatalog, err := UnmarshalModelCatalog(bytes)
//    bytes, err = modelCatalog.Marshal()

package golang

import "encoding/json"

func UnmarshalModelCatalog(data []byte) (ModelCatalog, error) {
	var r ModelCatalog
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *ModelCatalog) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

// Schema for defining AI models and model servers, for conversion to Backstage catalog
// entities
type ModelCatalog struct {
	// An array of AI models to be imported into the Backstage catalog                     
	Models                                                                    []Model      `json:"models"`
	// A deployed model server running one or more models, exposed over an API             
	ModelServer                                                               *ModelServer `json:"modelServer,omitempty"`
}

// A deployed model server running one or more models, exposed over an API
//
// Schema for defining AI model servers, for conversion to Backstage catalog entities
type ModelServer struct {
	// Annotations relating to the model server, in key-value pair format                  
	Annotations                                                          map[string]string `json:"annotations,omitempty"`
	// The API metadata associated with the model server                                   
	API                                                                  *API              `json:"API,omitempty"`
	// Whether or not the model server requires authentication to access                   
	Authentication                                                       *bool             `json:"authentication,omitempty"`
	// A description of the model server and what it's for                                 
	Description                                                          string            `json:"description"`
	// The URL for the model server's homepage, if present                                 
	HomepageURL                                                          *string           `json:"homepageURL,omitempty"`
	// The lifecycle state of the model server API                                         
	Lifecycle                                                            string            `json:"lifecycle"`
	// The name of the model server                                                        
	Name                                                                 string            `json:"name"`
	// The Backstage user that will be responsible for the model server                    
	Owner                                                                string            `json:"owner"`
	// Descriptive tags for the model server                                               
	Tags                                                                 []string          `json:"tags,omitempty"`
	// How to use and interact with the model server                                       
	Usage                                                                *string           `json:"usage,omitempty"`
}

// The API metadata associated with the model server
//
// Schema for defining the API exposed by model servers, for conversion to Backstage catalog
// entities
type API struct {
	// Annotations relating to the model, in key-value pair format                                                
	Annotations                                                                                 map[string]string `json:"annotations,omitempty"`
	// A link to the schema used by the model server API                                                          
	Spec                                                                                        string            `json:"spec"`
	// Descriptive tags for the model server's API                                                                
	Tags                                                                                        []string          `json:"tags,omitempty"`
	// The type of API that the model server exposes                                                              
	Type                                                                                        Type              `json:"type"`
	// The URL that the model server's REST API is exposed over, how the model(s) are interacted                  
	// with                                                                                                       
	URL                                                                                         string            `json:"url"`
}

// An AI model to be imported into the Backstage catalog
//
// Schema for defining AI models conversion to Backstage catalog entities
type Model struct {
	// Annotations relating to the model, in key-value pair format                                                 
	Annotations                                                                                  map[string]string `json:"annotations,omitempty"`
	// A URL to access the model's artifacts, e.g. on HuggingFace, Minio, Github, etc                              
	ArtifactLocationURL                                                                          *string           `json:"artifactLocationURL,omitempty"`
	// A description of the model and what it's for                                                                
	Description                                                                                  string            `json:"description"`
	// Any ethical considerations for the model                                                                    
	Ethics                                                                                       *string           `json:"ethics,omitempty"`
	// The URL pointing to any specific documentation on how to use the model on the model server                  
	HowToUseURL                                                                                  *string           `json:"howToUseURL,omitempty"`
	// The license used by the model (e.g. Apache-2).                                                              
	License                                                                                      *string           `json:"license,omitempty"`
	// The lifecycle state of the model server API                                                                 
	Lifecycle                                                                                    string            `json:"lifecycle"`
	// The name of the model                                                                                       
	Name                                                                                         string            `json:"name"`
	// The Backstage user that will be responsible for the model                                                   
	Owner                                                                                        string            `json:"owner"`
	// Support information for the model / where to open issues                                                    
	Support                                                                                      *string           `json:"support,omitempty"`
	// Descriptive tags for the model                                                                              
	Tags                                                                                         []string          `json:"tags,omitempty"`
	// Information on how the model was trained                                                                    
	Training                                                                                     *string           `json:"training,omitempty"`
	// How to use and interact with the model                                                                      
	Usage                                                                                        *string           `json:"usage,omitempty"`
}

// The type of API that the model server exposes
type Type string

const (
	Asyncapi Type = "asyncapi"
	Graphql  Type = "graphql"
	Grpc     Type = "grpc"
	Openapi  Type = "openapi"
)
