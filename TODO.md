# TODO


- [x] `down` fails if a release doesn't exists

- [ ] specify output dir (param and env var)


## General

- [ ] update README.md

- [ ] update examples folder

- [x] test `completion` command

- [ ] provide full example YAML


## Configuration (squadron.yaml)

- [x] provide warning where usage of `${ENV_VAR}` is not supported! Or implement it. => not implementing it yet 

- [x] test `op` template func

- [x] test `base64` template func

- [ ] check release name if merged

- [ ] check `default` template func

## Commands

## config

- [x] specifying the config file should override the default
   
    ```bash
    $ squadron config -f ${PWD}/squadrons/storefinder/squadron.yaml
    > Error: open ${PWD}/squadron.yaml: no such file or directory
    ```
   
- [x] send `squadron config` to stdout i.e. for `squadron config | bat`

- [x] provide `tag:` policies (git hash | timestamp | env ) @see scaffold => using template var

## build

- [x] not using `--files` flag

### up

- [ ] pass on `--verbose` flag to helm (`--debug`)

- [x] add `--no-generate` flag to skip re-generation of chart => use your own helm

### down

- [ ] pass on `--verbose` flag helm (`--debug`)


## Ideas

- [x] provide a way to add standard resources e.g. `secrets` as templates (to prevent sth like `shared`)


## Questions

- [x] Namespaces are still referenced in the code? => can be removed

- [x] what is the example data needed for? => remove

- [x] should we refactor it to see charts separately? Or if name provided? ‼️
- [x] concept: is it a good idea to up | generate only specific units? (chart will change depending on specified units) ‼️ 

- [x] is `units` the correct name? => check naval lingo ... ship, unit, battleship => unit is fine

- [x] do we want to re-support the `--data` parameter? => try it with overrides and ENV first

- [x] verbose flag `-v` doesn't really output more? => add logging to commands

## Backlog

- [ ] add `status` command to to check what is currently installed/orphaned on clusters. 
      This might require to somehow store meta data through helm.

- [x] add separate `push` command

- [x] add `global` variables within `squadron.yaml` as chart's `global`.

- [x] support `helm diff` plugin

- [ ] define reuse-able flags like `--namespace`

- [x] remove comments before passing to go template

- [x] improve 1Password integration: `export OP_SESSION_account_alias="XXX" | op signin account_alias --raw`

- [ ] pretty print template errors e.g. like zeus does

- [x] add initial tests

- [ ] add `lint` command to check valid yaml and pass it to helm

- [x] add `template` command to check valid yaml and pass it to helm

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