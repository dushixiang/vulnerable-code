#!/bin/bash

# 通用Docker镜像构建脚本
# 用法: ./build-image-common.sh <image_name> [push]
# 示例: ./build-image-common.sh cyberpoc/exec:v0.0.1 push

if [ $# -lt 1 ]; then
    echo "Usage: $0 <image_name> [push]"
    echo "Example: $0 cyberpoc/exec:v0.0.1 push"
    exit 1
fi

image_name="$1"
push_flag="$2"

echo "Building Docker image: ${image_name}"
docker build -t ${image_name} .

if [ $? -ne 0 ]; then
    echo "Build failed!"
    exit 1
fi

echo "Build completed successfully!"

if [ "$push_flag" = "push" ]; then
    echo "Pushing image to registry..."
    docker push ${image_name}
    if [ $? -eq 0 ]; then
        echo "Push completed successfully!"
    else
        echo "Push failed!"
        exit 1
    fi
fi
