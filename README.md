# AWS Security Posture - Solution

[AWS Security Posture](http://github.com/conijnio/aws-security-posture) collects security hub findings on a configurable interval, and extracts meaningful metrics. These metrics are stored in CloudWatch Metrics and can be visualized using CloudWatch DashBoards.


## Supported Generators

By default, the following generators are used to generate the compliance scores: 

- arn:aws:securityhub:::ruleset/cis-aws-foundations-benchmark
- aws-foundational-security-best-practices

## Getting started

Building and deploying the solution:

```shell
# Builds the solution
make build
# Deploys the solution 
make deploy
```
