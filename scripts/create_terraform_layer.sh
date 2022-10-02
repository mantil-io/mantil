#!/usr/bin/env bash -ex

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

layer_name=terraform-"${version//./-}"

aws lambda publish-layer-version \
 --layer-name "$layer_name" \
 --zip-file  "fileb://$tmp/layer.zip" \
 --compatible-architectures "arm64" \
 --compatible-runtimes "provided.al2" \
 --description "terraform $version" \
 --no-cli-pager
