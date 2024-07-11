# template-go-consumer

## Create Go Project

```sh
# /app>
go mod init github.com/<username>/<project-name>
```

## Folder Structure

```
├── README.md
├── app
│   ├── cmd
│   │   ├── main.go                         // entry point
│   │   └── services
│   │       ├── services.go                 // run all services
│   │       ├── config.service.go           // init config
│   │       └── goroutine.service.go        // init goroutines
│   ├── config
│   │   └── config.go                       // all config models
│   ├── Dockerfile                          // dockerfile
│   ├── env.example.yaml                    // environment variables
│   ├── go.mod
│   ├── go.sum
│   ├── pkg
│   │   ├── handlers                        // all handlers
│   │   ├── models                          // all dtos
│   │   ├── repository                      // all repositories
│   │   │   └── entities                    // all db entities
│       └── service                         // all services
├── .gitlab-ci.yml                          // devops ci/cd
├── k8s-manifests
│   ├── .env                                // define base environment variables
│   ├── deployment.yaml                     // kubernetes deployment
│   ├── config.yaml                         // kubernetes service configMap
│   ├── hpa.yaml                            // kubernetes service horizontal pod autoscaler
│   └── service.yaml                        // kubernetes service on cluster
```