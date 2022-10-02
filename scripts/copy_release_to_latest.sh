#!/usr/bin/env bash

git_tag=$(git describe --always)
tag="${1:-$git_tag}"

version="${tag:1}" # tag without leading v
bucket=s3://releases.mantil.io

function do-copy() {
    os=$1
    arch=$2
    from="$bucket"/"$tag"/mantil_"$version"_"$os"_"$arch".tar.gz
    to="$bucket"/latest/mantil_"$os"_"$arch".tar.gz
    #echo $from $to
    aws s3 cp "$from" "$to"
}

for arch in x86_64 arm64; do
    do-copy Darwin $arch
done

for arch in x86_64 i386; do
    do-copy Windows $arch
done


for arch in x86_64 i386 armv6 arm64; do
    do-copy Linux $arch
done

