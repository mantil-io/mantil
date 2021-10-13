package security

const credentialsTemplate = `{
    "Version": "2012-10-17",
    "Statement": [
        {{- if ne .LogGroup "" }}
        {
            "Action": [
                "logs:DescribeLogStreams",
                "logs:FilterLogEvents"
            ],
            "Effect": "Allow",
            "Resource": "arn:aws:logs:{{.Region}}:{{.AccountID}}:log-group:/aws/lambda/{{.LogGroup}}*"
        },
        {{ end }}
        {{- if ne .Stage "" }}
        {
            "Action": [
                "s3:PutObject"
            ],
            "Effect": "Allow",
            "Resource": "arn:aws:s3:::mantil-public-{{.Project}}-{{.Stage}}-*/*"
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
