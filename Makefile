SHELL := /bin/bash
PROJECT_NAME=security-posture

include .env

check-profile:
ifeq ($(origin AWS_PROFILE),undefined)
	$(error Please provide a profile by setting the AWS_PROFILE environment variable.)
endif

.DEFAULT_GOAL:=help
.PHONY: help
help:  ## Display this help
	$(info $(PROJECT_NAME) v$(VERSION))
	awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z0-9_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

test:
	$(info [+] Running unit tests)
	find . -name go.mod -execdir go test ./... -cover -coverprofile=coverage.out \;

reports/coverage.out:
	find . -name coverage.out -type f -exec sh -c 'cat {} | if [ "$1" = "{}" ]; then cat {}; else tail -n +2 {}; fi' sh {} \; > reports/coverage.temp
	echo "mode: set" | cat - reports/coverage.temp > reports/coverage.out
	rm reports/coverage.temp

coverage: reports/coverage.out
	$(info [+] Running code coverage)
	go tool cover -html=reports/coverage.out -o reports/coverage.html
	open reports/coverage.html

./events/collect-findings-payload.json:
	jq '.bucket = ${S3_BUCKET}' ./events/collect-findings.json > ./events/collect-findings-payload.json

.PHONY: collect-findings
collect-findings: ./events/collect-findings-payload.json
	sam local invoke \
		--profile ${AWS_PROFILE} \
		--event ./events/collect-findings-payload.json CollectFindingsFunction

.PHONY: build
build: ## Build the project
	sam build

.PHONY: deploy
deploy: check-profile ## Deploy the solution
	$(info [+] Deploy solution as $(PROJECT_NAME))
	sam deploy -t template.yaml \
		--stack-name $(PROJECT_NAME) \
		--capabilities=CAPABILITY_IAM

$(VERBOSE).SILENT:
