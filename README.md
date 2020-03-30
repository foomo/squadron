<a href="https://travis-ci.org/foomo/configurd">
    <img src="https://travis-ci.org/foomo/configurd.svg?branch=master" alt="Travis CI: build status">
</a>
<a href="https://goreportcard.com/report/github.com/foomo/configurd">
    <img src="https://goreportcard.com/badge/github.com/foomo/configurd" alt="GoReportCard">
</a>
<a href="https://godoc.org/github.com/foomo/configurd">
    <img src="https://godoc.org/github.com/foomo/configurd?status.svg" alt="GoDoc">
</a>

# Configurd

Application for managing kubernetes microservice environment

## Quickstart

```text
# Create a new folder with an example application with configurd:
$ configurd init [NAME]

$ cd [NAME]/

# Run install for predefined group and namespace:
$ configurd install [GROUP] -n [NAMESPACE]
```

## Structure

```text
/configurd
    /charts (Helm Charts)
        /<chart name>
    /services
        service-a.yaml
        service-b.yaml
    /namespaces
        /local (reserved, local)
            group-a.yaml
            group-b.yaml
        /node-a (remote)
            groub-c.yaml
```
## Commands

```text
# See:
$ configurd help
```