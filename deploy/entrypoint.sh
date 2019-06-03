#!/bin/bash
set -e

CURRENT_UID=${uid:-9999}

echo "Current UID : $CURRENT_UID"
useradd --shell /bin/bash -u $CURRENT_UID -o -c "" -m docker
export HOME=/home/docker

# Execute process
exec /usr/sbin/gosu docker /schemahero "$@"
