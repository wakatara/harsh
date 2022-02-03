#!/bin/bash

export OWNER=wakatara
export REPO=harsh
export SUCCESS_CMD="$REPO --version"
export BINLOCATION="/usr/local/bin"

version=$(curl -sI https://github.com/$OWNER/$REPO/releases/latest | grep -i "location:" | awk -F"/" '{ printf "%s", $NF }' | tr -d '\r')

if [ ! $version ]; then
    echo "Failed while attempting to install $REPO. Please manually install:"
    echo ""
    echo "1. Open your web browser and go to https://github.com/$OWNER/$REPO/releases"
    echo "2. Download the latest release for your platform. Call it '$REPO'."
    echo "3. chmod +x ./$REPO"
    echo "4. mv ./$REPO $BINLOCATION"
    exit 1
fi

hasCli() {

    hasCurl=$(which curl)
    if [ "$?" = "1" ]; then
        echo "You need curl to use this script."
        exit 1
    fi
}

getPackage() {
    uname=$(uname)
    userid=$(id -u)

    suffix=""
    case $uname in
    "Darwin")
    # suffix=".tar.gz"
        arch=$(uname -m)
        case $arch in
        "x86_64")
        suffix="Darwin_x86_64.tar.gz"
        ;;
        esac
        case $arch in
        "aarch64")
        suffix="Darwin_arm64.tar.gz"
        ;;
        esac
    ;;
    "Linux")
        arch=$(uname -m)
        echo $arch
        case $arch in
        "x86_64")
        suffix="Linux_x86_64.tar.gz"
        ;;
        esac
        case $arch in
        "i386")
        suffix="Linux_i386.tar.gz"
        ;;
        esac
        case $arch in
        "aarch64")
        suffix="Linux_arm64.tar.gz"
        ;;
        esac
        case $arch in
        "armv6l" | "armv7l")
        suffix="Linux_armv6.tar.gz"
        ;;
        esac
    ;;
    esac

    targetFile="/tmp/${REPO}_${suffix}"
    if [ "$userid" != "0" ]; then
        targetFile="$(pwd)/${REPO}_${suffix}"
    fi

    if [ -e "$targetFile" ]; then
        rm "$targetFile"
    fi

    url="https://github.com/$OWNER/$REPO/releases/download/$version/${REPO}_${suffix}"
    echo "Downloading package $url as $targetFile"

    curl -sSLf $url --output "$targetFile"

    if [ $? -ne 0 ]; then
        echo "Download Failed!"
        exit 1
    else
        extractFolder=$(pwd)
        echo "Download Complete, extracting $targetFile to $extractFolder ..."
        tar -xzf "$targetFile" -C "$extractFolder"
    fi

    if [ $? -ne 0 ]; then
        echo "\nFailed to expand archve: $targetFile"
        exit 1
    else
        # Remove the LICENSE and README
        echo "OK"
        rm "$(pwd)/LICENSE"
        rm "$(pwd)/README.md"

        # Get the parent dir of the 'bin' folder holding the binary
        # targetFile=$(echo "$targetFile" | sed "s+/${REPO}${suffix}++g")
        # suffix=$(echo $suffix | sed 's/.tgz//g')

        # installFile="${targetFile}/${REPO}/harsh"
        installFile="$(pwd)/harsh"

        chmod +x "$installFile"

        # Calculate SHA
        # https://github.com/wakatara/harsh/releases/download/v0.8.12/checksums.txt
        shaurl="https://github.com/$OWNER/$REPO/releases/download/$version/checksums.txt"
        shacheck="$(curl -sSLf $shaurl | grep ${REPO}_${suffix})"
        SHA256="$(echo $shacheck | awk '{print $1}')"
        echo "SHA256 fetched from release: $SHA256"
        # NOTE to other maintainers
        # There needs to be two spaces between the SHA and the file in the echo statement
        # for shasum to compare the checksums
        echo "$SHA256  $targetFile" | shasum -a 256 -c -s
        
        # Don't need the tar.gz any more so delete it
        rm "$targetFile"

        if [ $? -ne 0 ]; then
            echo "SHA mismatch! This means there must be a problem with the download"
            exit 1
        else
            if [ ! -w "$BINLOCATION" ]; then
                echo
                echo "============================================================"
                echo "  The script was run as a user who is unable to write"
                echo "  to $BINLOCATION. To complete the installation the"
                echo "  following commands may need to be run manually."
                echo "============================================================"
                echo
                echo "  sudo mv $installFile $BINLOCATION/$REPO"
                echo
                ./${REPO} --version
            else

                echo
                echo "SHA 256 integrity check on release ${version} succesful"
                echo "Running with sufficient permissions to attempt to move $REPO to $BINLOCATION"

                mv "$installFile" $BINLOCATION/$REPO

                if [ "$?" = "0" ]; then
                    echo "New version of $REPO installed to $BINLOCATION"
                fi

                if [ -e "$installFile" ]; then
                    rm "$installFile"
                fi
            echo "Checking successful install: running 'harsh --version'"
            ${SUCCESS_CMD}          
            fi
        fi
    fi
}

hasCli
getPackage

