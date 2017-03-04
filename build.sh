#!/usr/bin/env bash
PRO_ROOT=$(pwd)
APP_DIR=$PRO_ROOT/build/app
set -e -x
rm -rf $APP_DIR
mkdir -p $APP_DIR
go build -o $PRO_ROOT/build/registry-console $PRO_ROOT/cmd/plane/planectl.go
cp -rf $PRO_ROOT/js $APP_DIR/js
cp -rf $PRO_ROOT/pages $APP_DIR/pages
cp -rf $PRO_ROOT/fonts $APP_DIR/fonts
cp -rf $PRO_ROOT/css $APP_DIR/css
cp -rf $PRO_ROOT/Dockerfile $APP_DIR
cp -rf $PRO_ROOT/entrypoint.sh $APP_DIR
cp -rf $PRO_ROOT/kubernetes.yaml $APP_DIR
cd $PRO_ROOT/build
tar cf hub.tar app/
cd $PRO_ROOT
