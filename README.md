![Travis CI: build status](https://travis-ci.org/foomo/squadron.svg?branch=master) ![GoReportCard](https://goreportcard.com/badge/github.com/foomo/squadron) ![godoc](https://godoc.org/github.com/foomo/squadron?status.svg) ![goreleaser](https://github.com/foomo/squadron/workflows/goreleaser/badge.svg)

# Squadron

Application for managing kubernetes microservice environments. 

Use it, if a helm chart is not enough in order to organize multiple services into an effective squadron.

Another way to think of it would be `helm-compose`, because it makes k8s and helm way more approachable, not matter if it is development or production (where it just becomes another helm chart)

## Quickstart

```text
# Create a new folder with an example application with squadron:
$ squadron init [NAME]

$ cd [NAME]/

# Run install for predefined squadron and namespace:
$ squadron install [SQUADRON] -n [NAMESPACE]
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
            squadron-a.yaml
            squadron-b.yaml
        /node-a (remote)
            squadron-c.yaml
```
## Commands

```text
# See:
$ squadron help
```

## See also

Sometimes as a sailor or a pirate you might need to get a grapple : go get [github.com/foomo/gograpple/...](https//:github.com/foomo/gograpple)
