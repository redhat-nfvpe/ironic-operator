#!/bin/bash
set -ex

sed "s|OSH_PXE_IP|${PXE_IP}|g" /etc/nginx/nginx.conf > /tmp/pod-shared/nginx.conf

