#!/bin/bash
NAME=$1
PUSH=$2
DEPLOY=$3

for file in ./*
do
    if test -d $file
    then
      name=${file##*./}
      if [ $NAME == "ALL" ] || [ $name == $NAME ];then
        cd $file
        rm -rf $name
        set -ex
        CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o $name
        docker build -t kubegames/$name:latest .
        rm -rf $name
        set +ex
        
        if [ $PUSH == "true" ];then
          set -ex
          docker push kubegames/$name:latest
          set +ex
        fi

        if [ $DEPLOY == "true" ];then
          kubectl delete -f ./deploy.yaml
          set -ex
          kubectl apply -f ./deploy.yaml
          set +ex
        fi

        cd ../
      fi 
    fi
done

docker images|grep none|awk '{print $3 }'|xargs docker rmi
