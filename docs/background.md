# Background

## Explosion of 'AI Model Repositories'

A list that starts off incomplete, and will most likely always be a subset, of places where "AI Model Metadata" is 
accessible in some form or fashion:
- HuggingFace
- Ollama
- Kubeflow Model Registry
- KServe CRDs in K8s clusters
- MLFlow Model Registry
- OCI image registries like quay.io, registry.redhat.io, or docker.io
- API GWs like 3Scale or Kong

Each have some form of REST API.  Most if not all have a CLI that interacts with said REST API.

Building some form of normalization around taking data from those sources and constructing Backstage catalog
artifacts or "Entities" (where, by the way, the Backstage Catalog also has a REST API).

## Developer exploration vs. prescription from the enterprise

So "who" should be making the decision of which AI Models land in a Backstage instance to facilitate the use of AI
in applications developed with Backstage?

Whether the answer includes
- any developer on the given Backstage instance (where the developer team may have set it up)
- select developers on the instance
- or only the DevOps or MLOps or Platform engineers who set up the Backstage instances, where those folks are different people than the developers using Backstage 

can potentially result in different preference for "how" the Catalog is updated.

## Personas and their scenarios

Boiling that down into an agile story description:

As a platform/MLOPs engineer, I want to administrate the Backstage Catalog from the command line so that I can better automate administration of Backstage for AI related application development.

As a developer/tester engineer, I want to administrate the Backstage Catalog from the command line so that I can better automate verification pipelines for testing of Backstage’s AI related features.

## Syntax

Both [this UXD CLI guidelines reference](https://www.uxd-hub.com/entries/design/cli-guidelines) and the relative success (and initial contributors' background) in recent years with various CLI in the cloud computing space:

- `kubectl`
- `docker`
- `oc`
- `podman`
- `tkn`
- `aws`
- `rosa`
- `shp`

Generally speaking, you'll see either some form of:

- "cmd verb subject args" pattern ... `kubectl get pods ...` or `rosa create cluster` or `oc delete routes`
- "cmd verb-subject args" pattern ... `oc new-app ...` or `rosa list-clusters` or `oc new-build` or `oc cancel-build` or `oc import-image`
- and sometimes even "cmd subject verb args" .... `tkn pr list ...` or `shp build create ...`

When considering Backstage's current Catalog REST API, particularly some lack of symetry between which verbs apply to which subjects:

- You can only import to the ‘location’ REST with a URL pointing to a YAML document  containing the definition of multiple subjects (Components, Resource, API, pointers to TechDocs)
- But you can get/delete on all the subject types via REST api that don’t include the subject name in REST URI

The (initial) decision:  the prior art, a mixture of some "cmd verb-subject .." with  "cmd verb subject" where it makes sense.

So we have:

- `bac new-model kserve` for generating Backstage Catalog Entities in YAML format based on KServ CRD instances on a running Kubernetes cluster.
- `bac new-model kubeflow` for generating Backstage Catalog Entities in YAML format based information pulled from the Kubeflow Model Registry
- (with more sources to be added to `bac new-model`, see the [roadmap](roadmap.md))
- then after storing the YAML from `bac new-model` in a HTTP accessible file, you call `bac import-model <URL of that file>` to create a new Backstage `Location` with the entities defined in the YAML file referenced by the URL in a Backstage instance's catalog.  The output of that command will include the ID for the `Location`
- later on, if need be, you can run `bac delete-model <ID from bac import-model>` to remove the `Location` and associated entities.
- Lastly, there is a `bac get [locations|entities|components|resources|apis]` command for querying the Backstage Catalog for AI related entities (which we designate with special values for the `.spec.type` field)

The current plan is not to update this list when say everytime we add a new model registry option to `bac new-model`.  However, if we pivot on the
syntax philosophy in a significant way, we'll update this section.

## Implementation language(s)

The initial drop of this CLI is Golang.  Simply put, initial team expertise, plus some potential ideas around Kubernetes based 
services to help coordinate AI development tooling and AI deployments for development, staging, and production, lead to 
this decision.  Similarly, the OCI ecosystem, most notably the `docker` and `podman`, CLI are also in Golang.

That said, the Backstage ecosystem is TypeScript based.  And a TypeScript CLI could serve as a building block to 
a Backstage plugin.  And the AI space already has CLI written in several languages, most notably Python.  Lastly, API
Gateway CLI further expands the spectrum.  Ruby, vanilla JavaScript as well as TypeScript, based CLI exist in that space.

So, this project will entertain, or at least not dismiss, the notion of versions in more than one language to facilitate
plays in those different ecosystems.