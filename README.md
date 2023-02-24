
# Simple Saver Service

Learn where to save $$$ in S3

Simple Saver Service will scan your S3 buckets, provide summary data, analyses, and recommendations on how you could save $$$.

#### Amazon Simple Storage Service Requests Rates apply.

Visualizations created with API Response <br>
<img width="453" alt="Screenshot 2023-02-23 at 4 03 00 PM" src="https://user-images.githubusercontent.com/76878477/221059250-f29b76f5-697d-48fa-83b9-f1f352b9b926.png"><img width="456" alt="Screenshot 2023-02-23 at 4 05 20 PM" src="https://user-images.githubusercontent.com/76878477/221059412-4c35096b-e79b-4ce8-8323-d204212d2210.png"><img width="458" alt="Screenshot 2023-02-23 at 4 05 30 PM" src="https://user-images.githubusercontent.com/76878477/221059439-1ed1aae8-6ba0-47b1-bbba-14ea7db28909.png"><img width="459" alt="Screenshot 2023-02-23 at 4 05 38 PM" src="https://user-images.githubusercontent.com/76878477/221059460-6c640985-f7c2-46dd-bebe-683f823798d4.png">
## Run Locally

Clone the project

```bash
  go get https://github.com/helloevanhere/simple_saver_service
```

Go to the project directory

```bash
  cd ~/{YOUR-GO-PATH}/simple_saver_service
```

Start the server

```bash
  make run
```

Run tests

```bash
  make test
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
curl -X POST -H "Content-Type: application/json" -d '{"buckets":["my-bucket"]}' http://localhost:8080/storage_report
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

## Documentation
[Documentation Link](https://drive.google.com/drive/folders/18Yd0WPEuEYAYE0uWcVnp7xUh0xZXDRSb?usp=sharing)
