#!/usr/bin/env bash
set -ex

cd floe
rm floe
wget https://s3-eu-west-1.amazonaws.com/floe-deploys/floe
chmod +x floe
sudo supervisorctl stop floe
sleep 2
sudo supervisorctl start floe