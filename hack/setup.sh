#!/bin/bash
set -eou pipefail

OS=""
ARCH=""
DOWNLOAD_URL=""
DOWNLOAD_DIR=""
TEMP_DIRS=()

HELM=""
CHART_NAME="postgres-stash"
CHART_LOCATION="chart"

APPSCODE_ENV=${APPSCODE_ENV:-prod}
DOCKER_REGISTRY=${DOCKER_REGISTRY:-appscode}
IMAGE_TAG=10.2

BACKUP_ARGS=""
RESTORE_ARGS=""

ENABLE_PROMETHEUS_METRICS=true
METRICS_LABELS=""

UNINSTALL=0

function cleanup() {
  # remove temporary directories
  for dir in "${TEMP_DIRS[@]}"; do
    rm -rf "${dir}"
  done
}

# detect operating system
function detectOS() {
  OS=$(echo $(uname) | tr '[:upper:]' '[:lower:]')

  case "$OS" in
    # Minimalist GNU for Windows
    cygwin* | mingw* | msys*) OS='windows' ;;
  esac
}

# detect machine architecture
function detectArch() {
  ARCH=$(uname -m)
  case $ARCH in
    armv5*) ARCH="armv5" ;;
    armv6*) ARCH="armv6" ;;
    armv7*) ARCH="arm" ;;
    aarch64) ARCH="arm64" ;;
    x86) ARCH="386" ;;
    x86_64) ARCH="amd64" ;;
    i686) ARCH="386" ;;
    i386) ARCH="386" ;;
  esac
}

detectOS
detectArch

# download file pointed by DOWNLOAD_URL variable
# store download file to the directory pointed by DOWNLOAD_DIR variable
# you have to sent the output file name as argument. i.e. downloadFile myfile.tar.gz
function downloadFile() {
  if curl --output /dev/null --silent --head --fail "${DOWNLOAD_URL}"; then
    curl -fsSL ${DOWNLOAD_URL} -o ${DOWNLOAD_DIR}/$1
  else
    echo "File does not exist"
    exit 1
  fi
}

trap cleanup EXIT

show_help() {
  echo "install.sh - install postgres-stash catalog for stash"
  echo " "
  echo "install.sh [options]"
  echo " "
  echo "options:"
  echo "-h, --help                             show brief help"
  echo "    --docker-registry                  docker registry used to pull postgres-stash images (default: appscode)"
  echo "    --image-tag                        tag to use to pull postgres-stash docker image"
  echo "    --backup-args                      optional arguments to pass to pgdump command during backup"
  echo "    --restore-args                     optional arguments to pass to psql command during restore"
  echo "    --metrics-enabled                  specify whether to send prometheus metrics during backup or restore (default: true)"
  echo "    --metrics-labels                   labels to apply to prometheus metrics for backup or restore process (format: k1=v1,k2=v2)"
  echo "    --uninstall                        uninstall postgres-stash catalog"
}

while test $# -gt 0; do
  case "$1" in
    -h | --help)
      show_help
      exit 0
      ;;
    --docker-registry*)
      DOCKER_REGISTRY=$(echo $1 | sed -e 's/^[^=]*=//g')
      shift
      ;;
    --image-tag*)
      IMAGE_TAG=$(echo $1 | sed -e 's/^[^=]*=//g')
      shift
      ;;
    --backup-args*)
      BACKUP_ARGS=$(echo $1 | sed -e 's/^[^=]*=//g')
      shift
      ;;
    --restore-args*)
      RESTORE_ARGS=$(echo $1 | sed -e 's/^[^=]*=//g')
      shift
      ;;
    --metrics-enabled*)
      val=$(echo $1 | sed -e 's/^[^=]*=//g')
      if [[ "$val" == "false" ]]; then
        ENABLE_PROMETHEUS_METRICS=false
      fi
      shift
      ;;
    --metrics-labels*)
      METRICS_LABELS=$(echo $1 | sed -e 's/^[^=]*=//g')
      shift
      ;;
    --uninstall)
      UNINSTALL=1
      shift
      ;;
    *)
      echo "unknown flag: $1"
      echo " "
      show_help
      exit 1
      ;;
  esac
done

# Download helm if already not installed
if [ -x "$(command -v helm)" ]; then
  HELM=helm
else
  echo "Helm is not installed!. Downloading Helm."
  ARTIFACT="https://get.helm.sh"
  HELM_VERSION="v2.14.1"
  HELM_BIN=helm
  HELM_DIST=${HELM_BIN}-${HELM_VERSION}-${OS}-${ARCH}.tar.gz

  case "$OS" in
    cygwin* | mingw* | msys*)
      HELM_BIN=${HELM_BIN}.exe
      ;;
  esac

  DOWNLOAD_URL=${ARTIFACT}/${HELM_DIST}
  DOWNLOAD_DIR="$(mktemp -dt helm-XXXXXX)"
  TEMP_DIRS+=($DOWNLOAD_DIR)

  downloadFile ${HELM_DIST}

  tar xf ${DOWNLOAD_DIR}/${HELM_DIST} -C ${DOWNLOAD_DIR}
  HELM=${DOWNLOAD_DIR}/${OS}-${ARCH}/${HELM_BIN}
  chmod +x $HELM
fi

if [[ "$APPSCODE_ENV" == "dev" ]]; then
  CHART_LOCATION="chart"
else
  # download chart from remove repository and extract into a temporary directory
  CHART_LOCATION="$(mktemp -dt appscode-XXXXXX)"
  TEMP_DIRS+=(${CHART_LOCATION})
  TEMP_INSTALLER_REPO="${CHART_NAME}-installer"
  $HELM repo add "${TEMP_INSTALLER_REPO}" "https://charts.appscode.com/stable"
  $HELM fetch --untar --untardir ${CHART_LOCATION} "${TEMP_INSTALLER_REPO}/${CHART_NAME}"
  $HELM repo remove "${TEMP_INSTALLER_REPO}"
fi

if [ "$UNINSTALL" -eq 1 ]; then
  $HELM template ${CHART_LOCATION}/${CHART_NAME} \
  | kubectl delete -f -
  
  echo " "
  echo "Successfully uninstalled postgres-stash catalog"
else
# render the helm template and apply the resulting YAML
$HELM template ${CHART_LOCATION}/${CHART_NAME} \
  --set global.registry=${DOCKER_REGISTRY} \
  --set global.tag=${IMAGE_TAG} \
  --set global.backup.pgArgs=${BACKUP_ARGS} \
  --set global.restore.pgArgs=${RESTORE_ARGS} \
  --set global.metrics.enabled=${ENABLE_PROMETHEUS_METRICS} \
  --set global.metrics.labels=${METRICS_LABELS} \
| kubectl apply -f -

echo " "
echo "Successfully installed postgres-stash catalog"
fi
