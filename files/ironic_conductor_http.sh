#!/bin/bash
set -ex

mkdir -p /var/lib/pod_data/httpboot
cp -v /tmp/pod-shared/nginx.conf /etc/nginx/nginx.conf
exec nginx -g 'daemon off;'
