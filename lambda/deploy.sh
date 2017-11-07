#!/bin/bash
set -e

# http://jeremievallee.com/2017/03/26/aws-lambda-terraform/

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
BUILD_DIR="$SCRIPT_DIR/build"

if [ ! -e $BUILD_DIR ] ; then
    echo "Creating build dir"
    mkdir $BUILD_DIR
fi

VIRTUAL_ENV_DIR="$BUILD_DIR/venv"
if [ ! -e $VIRTUAL_ENV_DIR ] ; then
    echo "Creating virtualenv"
    virtualenv -p /usr/bin/python3.6 $VIRTUAL_ENV_DIR

    . $VIRTUAL_ENV_DIR/bin/activate
    pip install -r requirements.txt
    deactivate
fi


cd $VIRTUAL_ENV_DIR/lib/python3.6/site-packages
zip -r9 $BUILD_DIR/lambda.zip *
cd $SCRIPT_DIR

zip -r9 $BUILD_DIR/lambda.zip lambda.py
