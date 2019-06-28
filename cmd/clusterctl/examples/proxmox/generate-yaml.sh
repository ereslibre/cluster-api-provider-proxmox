#!/bin/sh

# Copyright 2019 Rafael Fernández López <ereslibre@ereslibre.es>

# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at

#     http://www.apache.org/licenses/LICENSE-2.0

# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -e

OVERWRITE=0
OUTPUT_DIR=${OUTPUT_DIR:-out}
CONTROLLER_VERSION=${CONTROLLER_VERSION:-0.2.0}

MACHINES_TEMPLATE_FILE=machines.yaml.template
MACHINES_GENERATED_FILE=${OUTPUT_DIR}/machines.yaml
CLUSTER_TEMPLATE_FILE=cluster.yaml.template
CLUSTER_GENERATED_FILE=${OUTPUT_DIR}/cluster.yaml

CLUSTER_NAME=${CLUSTER_NAME:-test-1}

NAMESPACE=${NAMESPACE:-kube-system}

SCRIPT=$(basename $0)
while test $# -gt 0; do
        case "$1" in
          -h|--help)
            echo "$SCRIPT - generates input yaml files for Cluster API on openstack"
            echo " "
            echo "$SCRIPT [options]"
            echo " "
            echo "options:"
            echo "-h, --help                show brief help"
            echo "-f, --force-overwrite     if file to be generated already exists, force script to overwrite it"
            exit 0
            ;;
          -f)
            OVERWRITE=1
            shift
            ;;
          --force-overwrite)
            OVERWRITE=1
            shift
            ;;
          *)
            break
            ;;
        esac
done

if [ $OVERWRITE -ne 1 ] && [ -f $MACHINES_GENERATED_FILE ]; then
  echo "File $MACHINES_GENERATED_FILE already exists. Delete it manually before running this script."
  exit 1
fi

if [ $OVERWRITE -ne 1 ] && [ -f $CLUSTER_GENERATED_FILE ]; then
  echo "File $CLUSTER_GENERATED_FILE already exists. Delete it manually before running this script."
  exit 1
fi

mkdir -p ${OUTPUT_DIR}

cat $MACHINES_TEMPLATE_FILE \
    | sed -e "s/\$CLUSTER_NAME/$CLUSTER_NAME/" \
  > $MACHINES_GENERATED_FILE
echo "Done generating $MACHINES_GENERATED_FILE"

cat $CLUSTER_TEMPLATE_FILE \
    | sed -e "s/\$CLUSTER_NAME/$CLUSTER_NAME/" \
  > $CLUSTER_GENERATED_FILE
echo "Done generating $CLUSTER_GENERATED_FILE"
