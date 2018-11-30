#!/bin/bash
set -ex

mkdir -p /var/lib/pod_data/httpboot
exec nginx -g 'daemon off;'
