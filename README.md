
# Simple Saver Service

Learn where to save $$$ in S3

Simple Saver Service will scan your S3 buckets, provide summary data and recommendations on how you could save $$$.

#### Amazon Simple Storage Service Requests Rates apply.
## Run Locally

Clone the project

```bash
  go get https://github.com/helloevanhere/simple_saver_service
```

Go to the project directory

```bash
  cd simple_saver_service
```

Install dependencies

```bash
  go install
  go mod tidy
```

Start the server

```bash
  make run
```

Stop the server

```bash
  make clean
```

## API Reference

### Storage Report
Retrieves a snapshot summary of your S3 usage

```
    POST /storage_report
    Host: localhost
    Content-Type: application/json
```

| Parameter | Type     | Description                |
| :-------- | :------- | :------------------------- |
| `buckets` | `[]string` | **Required**. List of S3 bucket names in your AWS account  |

Use "*" to retrieve storage report for all buckets.

#### Example: One Bucket

```bash
curl -X POST -H "Content-Type: application/json" -d '{"buckets":["my-bucket"]}' http://localhost:8080/storage_report
```
#### Example: All buckets
```bash
curl -X POST -H "Content-Type: application/json" -d '{"buckets":["*"]}' http://localhost:8080/storage_report
```


### Storage Recommendations
Retrieves analysis of bucket data and recommends simple saving solutions 

```
    POST /storage_recommendation
    Host: localhost
    Content-Type: application/json
```

| Parameter | Type     | Description                |
| :-------- | :------- | :------------------------- |
| `buckets` | `[]string` | **Required**. List of S3 bucket names in your AWS account  |


Use "*" to retrieve storage Recommendations for all buckets.
#### Example: One Bucket
```bash
curl -X POST -H "Content-Type: application/json" -d '{"buckets":["*"]}' http://localhost:8080/storage_report
```
#### Example: All Buckets
```bash
curl -X POST -H "Content-Type: application/json" -d '{"buckets":["*"]}' http://localhost:8080/storage_recommendation
```
## Environment Variables

To run this project, you will need to add the following environment variables to your `~/.zshrc` or `~/.bash-profile`

`AWS_ACCESS_KEY_ID`

`AWS_SECRET_ACCESS_KEY`

`AWS_REGION`



## AWS Credentials

The AWS user associated with your `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY` should have the following IAM permissions at a minimum:

```
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "VisualEditor0",
            "Effect": "Allow",
            "Action": [
                "s3:ListAllMyBuckets",
                "s3:GetLifecycleConfiguration",
                "s3:GetBucketVersioning",
                "s3:ListBucketMultipartUploads",
                "s3:ListBucket"
            ],
            "Resource": [
                "arn:aws:s3:::*"
            ]
        }
    ]
}
```
