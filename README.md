<a href="https://travis-ci.org/foomo/squadron">
    <img src="https://travis-ci.org/foomo/squadron.svg?branch=master" alt="Travis CI: build status">
</a>
<a href="https://goreportcard.com/report/github.com/foomo/squadron">
    <img src="https://goreportcard.com/badge/github.com/foomo/squadron" alt="GoReportCard">
</a>
<a href="https://godoc.org/github.com/foomo/squadron">
    <img src="https://godoc.org/github.com/foomo/squadron?status.svg" alt="GoDoc">
</a>

# Squadron

Application for managing kubernetes microservice environment

## Quickstart

```text
# Create a new folder with an example application with squadron:
$ squadron init [NAME]

$ cd [NAME]/

# Run install for predefined group and namespace:
$ squadron install [GROUP] -n [NAMESPACE]
```

## Structure

```text
/squadron
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
$ squadron help
```