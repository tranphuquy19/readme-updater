#!/bin/bash
# author: Quy Tran <tranpphuquy19@gmail.com>
# date: Aug 8 2021
APP_NAME="${PWD##*/}" # APP NAME OF THIS PROJECT

PLATFORMS_FILE_ARGUMENT=$1
PLATFORMS_FILE="${PLATFORMS_FILE_ARGUMENT:-platforms.txt}"
WORKING_DIR="${PWD}"
FULL_PATH_PLATFORMS_FILE=$WORKING_DIR/$PLATFORMS_FILE
start=$SECONDS
# check file platforms.txt
if [ ! -f $FULL_PATH_PLATFORMS_FILE ]; then
    echo "Platforms file: $FULL_PATH_PLATFORMS_FILE does not exist"
    exit 1
fi

# remove build folder
echo "Clean build folder"
rm -rf ./build

# create build folder
echo "Create build folder"
mkdir -p build

# install dependencies
echo "Install dependencies"
go mod download

# read platforms file
while read line; do
    temp=(${line//\// })

    GOOS=${temp[0]}
    GOARCH=${temp[1]}

    output_name=$APP_NAME

    if [ $GOOS = "windows" ]; then
        output_name+='.exe'
    fi

    mkdir -p build/$GOOS/$GOARCH
    output_path="$WORKING_DIR/build/$GOOS/$GOARCH/$output_name"

    echo "===================================================="
    echo "Building for OS=$GOOS Architecture=$GOARCH" 
    
    env GOOS=$GOOS GOARCH=$GOARCH go build -ldflags="-s -w" -o $output_path

    if [ ! -f $output_path ]; then
        echo "Failed when build for OS=$GOOS Architecture=$GOARCH"
        exit 1
    else
        fileSize=$(find "$output_path" -printf "%s")
        fileSizeInMb=$((fileSize / 1048576)).$(( (fileSize * 1000 / 1048576) %1000 ))
        echo "Done with output file: $output_path ($fileSizeInMb MB)"
    fi

    if [ $? -ne 0 ]; then
        echo 'An error has occurred! Aborting the script execution...'
        exit 1
    fi
done < $PLATFORMS_FILE

# Clean task
#echo "Clean build"
#go mod tidy
#rm -rf ./build

echo "Done in $(( SECONDS - start )) seconds!"