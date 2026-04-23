#!/bin/bash
DOMAIN_NAME="grounds.citybits.co.uk"
PROJECT_NAME="GroundsCityBits"
CURRENT_DATE=$(date "+%F_%T")
LOCAL_DIRECTORY="/Users/me/home/Dev/go-projects/grounds-maintenance"
REMOTE_DIRECTORY="/srv/apps/citybits/grounds"
USER="deploy@citybits.co.uk"
HAS_UPLOADED=0
SERVICE_NAME="grounds.citybits.service"

while true; do
    read -p "Upload latest build of ${PROJECT_NAME} to ${DOMAIN_NAME}?: " yn
    case $yn in
    [Yy]*)
        cd $LOCAL_DIRECTORY && make production-build-app
        ssh -n $USER "sudo systemctl stop $SERVICE_NAME && exit"
        echo "stopped $SERVICE_NAME ..."
        echo "copying latest files ..."
        rsync -rv $LOCAL_DIRECTORY/grounds.db $USER:$REMOTE_DIRECTORY/
        rsync -rv $LOCAL_DIRECTORY/bin/grounds.amd64 $USER:$REMOTE_DIRECTORY/
        ssh -n $USER " echo pwd && sudo systemctl start $SERVICE_NAME &&  exit"
        echo "starting $SERVICE_NAME ..."
        HAS_UPLOADED=1
        break
        ;;
    [Nn]*) exit ;;
    *) echo "Please answer y/n." ;;
    esac
done

