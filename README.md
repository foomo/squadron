<a href="https://travis-ci.org/foomo/configurd">
    <img src="https://travis-ci.org/foomo/configurd.svg?branch=master" alt="Travis CI: build status">
</a>
<a href='https://coveralls.io/github/foomo/configurd'>
    <img src='https://coveralls.io/repos/github/foomo/configurd/badge.svg' alt='Coverage Status' />
</a>
<a href="https://goreportcard.com/report/github.com/foomo/configurd">
    <img src="https://goreportcard.com/badge/github.com/foomo/configurd" alt="GoReportCard">
</a>
<a href="https://godoc.org/github.com/foomo/configurd">
    <img src="https://godoc.org/github.com/foomo/configurd?status.svg" alt="GoDoc">
</a>

# Configurd 

Application for managing kubernetes microservice environment


## Structure

/configurd
    /templates (Helm Templates)
        /services (Charts)
        /applications        
    /services
        service-a.yaml
        service-b.yaml
    /namespaces
        /local (reserved, local)
            service-a.yaml
            servicegroup-a.yaml
        /node-a (remote)
            global.yaml
            servicegroup-a.yaml
            