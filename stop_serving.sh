# /bin/bash

if docker ps | grep ":80"
then
    echo "OK";
    docker stop $(docker ps | grep ":80" | awk '{print $1}')
else
    echo "NOT OK";
fi