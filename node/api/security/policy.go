package security

const credentialsTemplate = `{
    "Version": "2012-10-17",
    "Statement": [
        {{$first := true}}
        {{- range .Buckets}}
        {{- if ne . "" }}
        {{if $first}}{{$first = false}}{{else}},{{end}}
        {
            "Action": [
                "s3:PutObject"
            ],
            "Effect": "Allow",
            "Resource": "arn:aws:s3:::{{.}}/*"
        }
        {{ end }}
        {{ end }}
        {{- if ne .LogGroupsPrefix "" }}
        ,{
            "Action": [
                "logs:DescribeLogStreams",
                "logs:FilterLogEvents"
            ],
            "Effect": "Allow",
            "Resource": "arn:aws:logs:{{.Region}}:{{.AccountID}}:log-group:{{.LogGroupsPrefix}}*"
        }
        {{ end }}
    ]
}`
