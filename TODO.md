# TODO


## Configuration (squadron.yaml)

- [ ] check release name if merged

- [ ] check `default` template func

## Commands

### up

- [ ] pass on `--verbose` flag to helm (`--debug`)

### down

- [ ] pass on `--verbose` flag helm (`--debug`)


## Backlog

- [ ] add `status` command to to check what is currently installed/orphaned on clusters.
      This might require to somehow store meta data through helm.

- [ ] define reuse-able flags like `--namespace`

- [ ] pretty print template errors e.g. like zeus does

- [ ] add `lint` command to check valid yaml and pass it to helm

- [ ] provide some release information commands via helm e.g. `list`, `status`, ...

- [ ] check `helm get ...` commands

- [ ] check `helm history ...` commands

- [ ] check `helm rollback ...` commands

- [ ] check `helm template ...` commands

- [ ] add YAML schema

- [ ] colorized output

- [ ] show progress as output

- [ ] check `helm` exists and version is supported

- [ ] add further test till at least coverage > 80%

- [ ] add validation check to prevent `global` squadron name (collides with vars)!

- [ ] support parallel `build` and `push` with context canceling
