#!/usr/bin/env bash -e

aws sts get-federation-token --name mantil-proj1 --duration-seconds 900 --policy "$(jq -c . policy.json)" --no-cli-pager > output
cat output

export AWS_ACCESS_KEY_ID="$(jq -r .Credentials.AccessKeyId output)"
export AWS_SECRET_ACCESS_KEY="$(jq -r .Credentials.SecretAccessKey output)"
export AWS_SESSION_TOKEN="$(jq -r .Credentials.SessionToken output)"

name=hello
build_name=hello:v016b704-dirty.zip
    aws lambda update-function-code --no-cli-pager \
        --function-name "proj1-try-mantil-team-$name" \
        --s3-bucket try.mantil.team-lambda-functions \
        --s3-key "functions/$build_name" \
        --publish
