[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](https://www.gnu.org/licenses/gpl-3.0)
[![Build](https://github.com/DigitalChinaOpenSource/MDI-kernel/actions/workflows/build.yml/badge.svg)](https://github.com/DigitalChinaOpenSource/MDI-kernel/actions/workflows/build.yml)
[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=MDI-kernel&metric=alert_status)](https://sonarcloud.io/dashboard?id=MDI-kernel)
[![Code Smells](https://sonarcloud.io/api/project_badges/measure?project=MDI-kernel&metric=code_smells)](https://sonarcloud.io/dashboard?id=MDI-kernel)
[![Lines of Code](https://sonarcloud.io/api/project_badges/measure?project=MDI-kernel&metric=ncloc)](https://sonarcloud.io/dashboard?id=MDI-kernel)
[![Maintainability Rating](https://sonarcloud.io/api/project_badges/measure?project=MDI-kernel&metric=sqale_rating)](https://sonarcloud.io/dashboard?id=MDI-kernel)
[![Coverage](https://sonarcloud.io/api/project_badges/measure?project=MDI-kernel&metric=coverage)](https://sonarcloud.io/dashboard?id=MDI-kernel)



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