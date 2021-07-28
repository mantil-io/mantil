package security

const FederationTokenPolicyTemplate = `
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
        		"s3:GetObject",
        		"s3:PutObject"
      		],
            "Resource": "arn:aws:s3:::{{.Bucket}}/*"
        },
        {
            "Effect": "Allow",
            "Action": [
                "ecr:BatchCheckLayerAvailability",
                "ecr:CompleteLayerUpload",
                "ecr:InitiateLayerUpload",
                "ecr:PutImage",
                "ecr:UploadLayerPart"
            ],
            "Resource": "arn:aws:ecr:{{.Region}}:{{.AccountID}}:repository/mantil-project-{{.OrganizationName}}-{{.Name}}"
        },
		{
	    	"Effect": "Allow",
            "Action": "ecr:GetAuthorizationToken",
            "Resource": "*"
		}
    ]
}
`
