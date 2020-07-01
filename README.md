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

Application for managing kubernetes microservice environments. 

Use it, if a helm chart is not enough in order to organize multiple services into a effective squadron.

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