#!/bin/bash
set -e

CURRENT_UID=${uid:-9999}

useradd --shell /bin/bash -u $CURRENT_UID -o -c "" -m docker
export HOME=/home/docker

# Execute process
exec /usr/sbin/gosu docker /schemahero "$@"
