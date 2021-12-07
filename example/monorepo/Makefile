.DEFAULT_GOAL:=help

export PROJECT_ROOT=$(PWD)

storefinder.%:files=./squadrons/squadron.yaml,./squadrons/storefinder/squadron.yaml,./squadrons/storefinder/frontend/squadron.yaml,./squadrons/storefinder/backend/squadron.yaml

.PHONY: storefinder.frontend.config
## Show config
storefinder.frontend.config:
	@squadron -f ${files} config

.PHONY: storefinder.frontend.template
## Show config
storefinder.frontend.template:
	@squadron -f ${files} template

.PHONY: storefinder.frontend.generate
## Generate helm chart
storefinder.frontend.generate:
	@squadron -f ${files} generate

storefinder.%:files=./squadrons/squadron.yaml,./squadrons/storefinder/squadron.yaml,./squadrons/storefinder/frontend/squadron.yaml,./squadrons/storefinder/backend/squadron.yaml

.PHONY: storefinder.backend.config
## Show config
storefinder.backend.config:
	@squadron -f ${files} config

.PHONY: storefinder.backend.template
## Show template
storefinder.backend.template:
	@squadron -f ${files} template

.PHONY: storefinder.backend.generate
## Generate helm chart
storefinder.backend.generate:
	@squadron -f ${files} generate

## === Utils ===

## Show help text
help:
	@awk '{ \
			if ($$0 ~ /^.PHONY: [a-zA-Z\-\_0-9]+$$/) { \
				helpCommand = substr($$0, index($$0, ":") + 2); \
				if (helpMessage) { \
					printf "\033[36m%-23s\033[0m %s\n", \
						helpCommand, helpMessage; \
					helpMessage = ""; \
				} \
			} else if ($$0 ~ /^[a-zA-Z\-\_0-9.]+:/) { \
				helpCommand = substr($$0, 0, index($$0, ":")); \
				if (helpMessage) { \
					printf "\033[36m%-23s\033[0m %s\n", \
						helpCommand, helpMessage"\n"; \
					helpMessage = ""; \
				} \
			} else if ($$0 ~ /^##/) { \
				if (helpMessage) { \
					helpMessage = helpMessage"\n                        "substr($$0, 3); \
				} else { \
					helpMessage = substr($$0, 3); \
				} \
			} else { \
				if (helpMessage) { \
					print "\n                        "helpMessage"\n" \
				} \
				helpMessage = ""; \
			} \
		}' \
		$(MAKEFILE_LIST)
