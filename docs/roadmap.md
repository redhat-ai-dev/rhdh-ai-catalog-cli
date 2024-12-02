# Roadmap

## Fit and finish

These updates are more along the lines of general usability.  We'll use [this epic](https://issues.redhat.com/browse/RHDHPAI-40)
to officially track these items, but until we have a public Jira, we'll give some sense at least of what we considering
with the table below.

| idea                             | description                                                    | tracker                                             | status        |
|----------------------------------|----------------------------------------------------------------|-----------------------------------------------------|---------------|
| config file                      | capture connection and global parameters for reuse             | [Jira](https://issues.redhat.com/browse/RHDHPAI-43) | unimplemented |
| entity field configmap           | with new-model, allow for field overrides from configmap       | [Jira](https://issues.redhat.com/browse/RHDHPAI-44) | unimplemented |
| backstage cert/token cm/secret   | store/retrieve cert and token for backstage                    | [Jira](https://issues.redhat.com/browse/RHDHPAI-45) | unimplemented |
| third party cert/token cm/secret | store/retrieve cert and token for third party                  | [Jira](https://issues.redhat.com/browse/RHDHPAI-46) | unimplemented |
| backstage cert flag              | file/env var for backstage cert                                | [Jira](https://issues.redhat.com/browse/RHDHPAI-47) | unimplemented |
| third party cert flag            | file/env var for third party cert                              | [Jira](https://issues.redhat.com/browse/RHDHPAI-48) | unimplemented |
| entity field local file          | with new-mode, allow for field overrides from file             | [Jira](https://issues.redhat.com/browse/RHDHPAI-49) | unimplemented |
| fetch URLs from routes/ingress   | when backstage,third party running on K8s, find URL            | [Jira](https://issues.redhat.com/browse/RHDHPAI-54) | unimplemented |
| fetch URLs from routes/ingress   | when kubeflow third party running on K8s, find URL             | [Jira](https://issues.redhat.com/browse/RHDHPAI-55) | unimplemented |
| flags for output                 | allow for output summary vs. json vs. yaml etc.                | [Jira](https://issues.redhat.com/browse/RHDHPAI-56) | unimplemented |
| flags for field overrides        | new-model provide field values via command line flags          | [Jira](https://issues.redhat.com/browse/RHDHPAI-50) | unimplemented |
| release process                  | initially github action/goreleaser; eventually konflux         | [Jira](https://issues.redhat.com/browse/RHDHPAI-57) | unimplemented |
| e2e tests                        | running against "live" data somehow                            | [Jira](https://issues.redhat.com/browse/RHDHPAI-59) | unimplemented |
| filter api queries for "ai"      | with no unique spec.type for API either state no filter or fix | [Jira](https://issues.redhat.com/browse/RHDHPAI-58) | unimplemented |

## Augment Initial Backstage Query Support

Quite possibly use of the query function will drive desire for more features in this space.

Some ideas have already emerged, but until we get some consensus on which ideas have traction, we'll wait on
building a table here.  Look at [this epic](https://issues.redhat.com/browse/RHDHPAI-62) for the latest set of ideas.

## Upstream Backstage

### Direct injection of YAML when Creating Entities in the Catalog

While the input format of the body supplied to this REST API has a type field, best as we can tell, the only types supported
are a HTTP accessible URL or a local file.

Will users of the CLI be happy with having to take the extra step of pushing the YAML from `bac new-model ...` to say a
file hosted in Git repository and then provide that URL to `bac import-model ...` ?

Or are will they be enamored with the idea (albeit it punts on Gitops) with a flow like

```shell
bac new-model kserve | bac import-model -f -
```

or 

```shell
bac new-model kubeflow > catalog-info.yaml
bac import-model catalog-info.yaml 
```

If we get enough demand for such a change with upstream Backstage, we'll drive 
work with it with [this Jira](https://issues.redhat.com/browse/RHDHPAI-61).

#### One advantage of being in Golang with respect to Gitops with GitHub ...

The uber functional `gh` command, [GitHub's official command line tool](https://github.com/cli/cli), happens to be written
in Golang.

The `gh` command has a A LOT of usability features outside of the standard `git` command.

In response, GitLab has the `glab` command, [GitLab's official command line tool](https://gitlab.com/gitlab-org/cli).  It is 
also written in golang, and has a very similar syntax, with the same features in many cases.

The `bac new-model ...` invocation could take in the necessary Git credentials and connection parameters and via use of `gh` or `glab` code as 
vendored dependencies:
- create the repository if needed
- create not just a commit, but a pull request, with the `catalog-info.yaml` file(s)
- rebase if necessary
- merge that pull request
- and then post to the Backstage Catalog REST API to import ... i.e. what we do with `bac import-model ...`


## New 'Model Metadata' sources

| Source      | Summary/REST/CRDs                | Questions/Comments                                            | Priority | Tracker                                              | Status  |
|-------------|----------------------------------|---------------------------------------------------------------|----------|------------------------------------------------------|---------|
| Kubeflow    | Endpoint URL.  Has both REST/CRD | RHOAI Jira marked done.  Which version?  End to end examples? | high     | [Jira](https://issues.redhat.com/browse/RHDHPAI-64)  | waiting |
| 3Scale      | All data ready.  Yes REST/CRDs   | Perhaps the next highest item. Devex vs. RHOAI priorities     | high     | [Jira](https://issues.redhat.com/browse/RHDHPAI-65)  | new     |
| HuggingFace | All data ready.  REST only       | Most popular source for public models. Best for tech docs     |          | [Jira](https://issues.redhat.com/browse/RHDHPAI-667) | new     |
| MLFlow      | All data ready.  REST only       | Mature. KServe support. ai-on-openshift.io refs. Competitor?  |          |                                                      | new     |
| Ollama      | All data ready.  REST only       | RHDH AI/Devex use vs. RHOAI sanctioned, indemnification       |          | [Jira](https://issues.redhat.com/browse/RHDHPAI-66)  | new     |
| OCI         | Endpoint URL ? REST, 'oc image'  | Often cited at strategy level. Requires coupling with ?       | high     |                                                      | new     |
| Open WebUI  | All data ready.  REST only       | Competition? But supports Kubernetes.                         |          |                                                      | new     |
|             |                                  |                                                               |          |                                                      |         |
|             |                                  |                                                               |          |                                                      |         |

## TechDocs

TechDocs might be what most holds back bypassing storage in a Git repo when importing model.  A key, positive aspect of 
TechDocs is co-locating markdown doc with code/config convention.  And it is not part of the Catalog's "Kubernetes-like" API.  

A typical flow in most cases then will be:

- Run `bac new-model ...` to get the `Component`, `Resource`, and `API` definitions for the AI Model
- Store in a Git repo
- Build the TechDocs manually and store in the same Git repo in the correct spot
- Run `bac import-model <backstage url>` 

Many of the AI Model Registries still don't emphasize key developer scenarios, including the need for documentation of the
AI Model.

Now, most AI servers themselves have the `/docs` URI that allows for Swagger or FastAPI styled documentation.  Certainly
a start, but TechDocs typically provide this plus additional information.

But HuggingFace with its ModelCards is really the only game in town wrt the Model Registries and something that approaches
Tech Docs.  And even there, it is one markdown file, where as TechDoc typically is broken up into multiple markdown files.

And the internal 3Scale based "Models as a Service" has some additional examples in various languages and frameworks.

With that inventory, it is still TBD on what value add the CLI can provide with respect to Backstage TechDocs.