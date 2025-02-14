# Commands

### download
Downloads a file from S3
```yaml
steps:
  - run-download:
      service: "s3_service"
      task: "download"
      input:
        file_path: "/..."
        bucket_name: "..."
        s3_key: "..."
        aws_access_key: "..."
        aws_secret_key: "..."
        aws_region: "us-east-1"
        aws_session_token: "..."
```

Output
```json
{
}
```


### upload
Downloads a file from S3
```yaml
steps:
  - run-upload:
      service: "s3_service"
      task: "upload"
      input:
        file_path: "/..."
        bucket_name: "..."
        s3_key: "..."
        aws_access_key: "..."
        aws_secret_key: "..."
        aws_region: "us-east-1"
        aws_session_token: "..."
```

Output
```json
{
}
```