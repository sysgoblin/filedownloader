#!/bin/bash
# should be run from project root

# set the app name
APP_NAME="godownload"

# set the target platforms and architectures
PLATFORMS=("linux" "darwin" "windows")
ARCHITECTURES=("amd64" "386")

# loop through each platform/architecture combination and compile the app
for PLATFORM in "${PLATFORMS[@]}"
do
  for ARCH in "${ARCHITECTURES[@]}"
  do
    # set the output file name
    OUTPUT_FILE="build/${APP_NAME}_${PLATFORM}_${ARCH}"

    # compile the app for the current platform/architecture
    env GOOS="$PLATFORM" GOARCH="$ARCH" go build -o "$OUTPUT_FILE"

    # add the appropriate file extension for Windows
    if [ "$PLATFORM" = "windows" ]; then
      mv "$OUTPUT_FILE" "$OUTPUT_FILE.exe"
    fi
  done
done