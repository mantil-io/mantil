package security

const CredentialsTemplate = `{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Action": [
                "s3:PutObject"
            ],
            "Effect": "Allow",
            "Resource": "arn:aws:s3:::{{.Bucket}}/*"
        },
        {
            "Action": [
                "ecr:BatchCheckLayerAvailability",
                "ecr:CompleteLayerUpload",
                "ecr:InitiateLayerUpload",
                "ecr:PutImage",
                "ecr:UploadLayerPart"
            ],
            "Effect": "Allow",
            "Resource": "arn:aws:ecr:{{.Region}}:{{.AccountID}}:repository/mantil-project-{{.OrganizationName}}-{{.Name}}"
        },
        {
            "Action": "ecr:GetAuthorizationToken",
            "Effect": "Allow",
            "Resource": "*"
        },
        {
            "Action": [
                "logs:DescribeLogStreams",
                "logs:FilterLogEvents"
            ],
            "Effect": "Allow",
            "Resource": "arn:aws:logs:{{.Region}}:{{.AccountID}}:log-group:/aws/lambda/{{.Name}}*"
        }
    ]
}
`
