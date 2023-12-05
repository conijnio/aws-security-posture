# Collecting Findings

We are collecting all findings that match a certain filter. Meaning that the result can become really big. We have ran into service limitations previously like:

- MaxResults of the `GetFindings` call of Security Hub API, Minimum value of 1. Maximum value of 100.
- The rate limits of the [Security Hub API](https://docs.aws.amazon.com/securityhub/1.0/APIReference/Welcome.html), RateLimit of 3 requests per second. BurstLimit of 6 requests per second.
- AWS Lambda Function timeout, 900 seconds (15 minutes)
- AWS Step Functions payload size, 256KB

For this reason we architected the collection of the findings as followed:

- Invoke a single page fetch to not need to sleep in the Lambda function.
- Reduce the result to the bare minimum.
- Store the results in a S3 bucket.
- We will loop until there is no longer a `NextToken`.
- Implement retry logic on the task invocation.
- Pass the list of objects to the next step.

## Required information

The following information is marked as mandatory to do a proper scoring of the security posture:

- **Id**, the actual unique id of the finding.
- **Compliance.Status**, what the status is of the specific finding.
- **ProductArn**, what product raised the finding.
- **GeneratorId**, what generator raised the finding.
- **AwsAccountId**, what account raised the finding.

By stripping out all the other data we are reducing the needed storage on S3, and improve the overall speed as we don't need to shift big files around.
