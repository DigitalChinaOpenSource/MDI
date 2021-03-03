
# Introduction

This is part of Dc low code projects for data layer.

# Reference links
[Components & licences](ComponentsApplied.MD)

## What's contained in this project

```
mdi
├── cmd/
│   └── kernel/                                     the core part of api
│   │   └── router/
│   │   │   └── handler/ 
│   │   │   └── router.go                               
│   │   └── main.go   
│   │   └── .dockerfile                                  
│   └── integrationTests/                               
│   └── Agent/                                      the User-Generated API service
│   │   └── router/
│   │   │   └── handler/ 
│   │   │   │   └── pg/
│   │   │   │   │   └── rest/
│   │   │   │   │   └── graphQL/
│   │   │   │   └── ti/
│   │   │   │       └── rest/
│   │   │   │       └── graphQL/
│   │   │   └── router.go
│   │   └── main.go
│   │   └── .dockerfile 
├── internal/
│   └── kernel/                                     common components for modeling and persistence
│   │   └── metadata/                                        
|   │   │   └──modeling
│   │   │   │   └── modeling.xsd
│   │   │   │   └── modeling.go
|   │   │   └──migration
│   │   │       └── migration.xsd
│   │   │       └── migration.go
│   │   └── postgres/
│   │   │       └──handler.go
│   │   └── tiDB/                                                
│   │           └──handler.go
│   └── unitTests/                                                
├── └── pkg/
│       └── middleware/

```

## Dependencies

Install the following

- [gin](https://github.com/gin-gonic/gin)

- [gorm](https://gorm.io/zh_CN/docs/index.html)

## Run Service

```shell
cd cmd\kernel

go run main.go
```

your service endpoint will run like such url:  http://localhost:8080/ping

## Test Coverage
```
Methods and tests need to be written in the same file
Example:
    pkg/
    └── xxx.go
    └── xxx_test.go
```