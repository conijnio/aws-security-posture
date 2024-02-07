SHELL := /bin/bash
PROJECT_NAME=security-posture

include .env

check-profile:
ifeq ($(origin AWS_PROFILE),undefined)
	$(error Please provide a profile by setting the AWS_PROFILE environment variable.)
endif

check-region:
ifeq ($(origin AWS_REGION),undefined)
	$(error Please provide a profile by setting the AWS_REGION environment variable.)
endif


.DEFAULT_GOAL:=help
.PHONY: help
help:  ## Display this help
	$(info $(PROJECT_NAME) v$(VERSION))
	awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z0-9_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

.PHONY: test
test:
	$(info [+] Running unit tests)
	find . -name go.mod -mindepth 2 -execdir go test ./... -coverprofile=coverage.out -covermode count \;

.PHONY: lint
lint:
	$(info [+] Running go fmt)
	find . -name go.mod -mindepth 2 -execdir go fmt ./... \;
	find . -name go.mod -mindepth 2 -execdir go vet ./... \;

.PHONY: coverage
coverage: clean-reports reports/coverage.out
	$(info [+] Running code coverage)
	go tool cover -html=reports/coverage.out -o reports/coverage.html
	open reports/coverage.html

.PHONY: clean-reports
clean-reports:
	rm -rf $(PWD)/reports

reports:
	$(info [+] Clean previous reports)
	mkdir reports

reports/report.xml: reports
	find . -mindepth 2 -name go.mod -execdir go run gotest.tools/gotestsum@latest --junitfile report.xml --format testname  \;
	go run github.com/nezorflame/junit-merger@latest -o ./reports/report.xml ./lambdas/**/report.xml

reports/coverage.out: reports test
	find . -name coverage.out -type f -exec sh -c 'cat {} | if [ "$1" = "{}" ]; then cat {}; else tail -n +2 {}; fi' sh {} \; > reports/coverage.temp
	echo "mode: count" | cat - reports/coverage.temp > reports/coverage.out
	rm reports/coverage.temp

reports/coverage.xml: reports/coverage.out
	$(info Collecting Code Coverage)
	go run github.com/boumenot/gocover-cobertura < reports/coverage.out > reports/coverage.xml

./events/collect-findings-payload.json:
	jq '.bucket = ${S3_BUCKET}' ./events/collect-findings.json > ./events/collect-findings-payload.json

.PHONY: collect-findings
collect-findings: ./events/collect-findings-payload.json
	sam local invoke \
		--profile ${AWS_PROFILE} \
		--event ./events/collect-findings-payload.json CollectFindingsFunction

.PHONY: build
build: ## Build the project
	sam build --parallel

.PHONY: validate
validate: ## Validate the SAM template
	sam validate --lint

.PHONY: deploy
deploy: check-profile check-region ## Deploy the solution
	$(info [+] Deploy solution: $(PROJECT_NAME))
	sam deploy -t .aws-sam/build/template.yaml \
		--stack-name $(PROJECT_NAME) \
		--capabilities=CAPABILITY_IAM CAPABILITY_NAMED_IAM \
		--region=$(AWS_REGION) \
		--s3-bucket cf-templates-1tw6xuuyelyb4-$(AWS_REGION) \
		--no-fail-on-empty-changeset \
		--parameter-overrides \
			CollectInterval="'6 hours'" \
			ConformancePack="OrgConformsPack-lz-framework-6cogv2l0"

.PHONY: delete
delete: check-region ## Delete the solution
	$(info [+] Deleting solution: $(PROJECT_NAME))
	aws cloudformation delete-stack \
		--stack-name $(PROJECT_NAME) \
		--region=$(AWS_REGION) \

.PHONY: tidy
tidy: ## Run `go get -u` and `go mod tidy` for all modules
	$(info [+] Running `go get -u` and `go mod tidy`)
	find . -name go.mod -execdir go get -u \;
	find . -name go.mod -execdir go mod tidy \;

$(VERBOSE).SILENT:
