package security

const CredentialsTemplate = `{
    "Version": "2012-10-17",
    "Statement": [
        {{ if ne .Public.Bucket "" }}
        {
            "Action": [
                "s3:PutObject"
            ],
            "Effect": "Allow",
            "Resource": "arn:aws:s3:::{{.Public.Bucket}}/*"
        },
        {{ end }}
        {{ if ne .LogGroup "" }}
        {
            "Action": [
                "logs:DescribeLogStreams",
                "logs:FilterLogEvents"
            ],
            "Effect": "Allow",
            "Resource": "arn:aws:logs:{{.Region}}:{{.AccountID}}:log-group:/aws/lambda/{{.LogGroup}}*"
        },
        {{ end }}
        {
            "Action": [
                "s3:PutObject"
            ],
            "Effect": "Allow",
            "Resource": "arn:aws:s3:::{{.Bucket}}/*"
        }
    ]
}`
