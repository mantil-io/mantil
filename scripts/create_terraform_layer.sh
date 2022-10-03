#!/usr/bin/env bash -e

version="1.3.1"

tmp=$(mktemp -d)

cd $tmp

wget https://releases.hashicorp.com/terraform/"$version"/terraform_"$version"_linux_arm64.zip
unzip terraform_"$version"_linux_arm64.zip
rm terraform_"$version"_linux_arm64.zip

mkdir bin
mv terraform bin

cd $tmp
zip -r layer.zip bin/
unzip -l $tmp/layer.zip

layer_name=terraform-"${version//./_}"

regions=(ap-south-1 ap-southeast-1 ap-southeast-2 ap-northeast-1 eu-central-1 eu-west-1 eu-west-2 us-east-1 us-east-2 us-west-2)

for region in ${regions[@]}; do

    export AWS_REGION=$region

    echo publishing layer in $region ...

    aws lambda publish-layer-version \
        --layer-name "$layer_name" \
        --zip-file  "fileb://$tmp/layer.zip" \
        --compatible-architectures "arm64" \
        --compatible-runtimes "provided.al2" \
        --description "terraform $version" \
        --no-cli-pager > out_publish_layer_version

    arn=$(jq -r ".LayerArn" $tmp/out_publish_layer_version)
    version=$(jq -r ".Version" $tmp/out_publish_layer_version)


    aws lambda add-layer-version-permission \
        --layer-name $layer_name \
        --statement-id xaccount \
        --action lambda:GetLayerVersion  \
        --principal "*" \
        --version-number $version > $tmp/out_add_layer_version_permission

    echo arn: $arn, version: $version
done
