#!/usr/bin/env bash
#
# Configures and creates kubernetes ingress controller, it will use default
# values from ./helpers.bash
#######################################

echo "Installing ingress into kubectl cluster..."

dir=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

source "${dir}/helpers.bash"

set -e

if [ -z "${dns_probes_ips}" ]; then
    dns_probes_ips=( "127.0.0.1" "127.0.0.1" )
fi

ingress="${dir}/../deployments/guestbook/ingress"

kubectl create -f "${ingress}" || true

kubectl --namespace=kube-system get svc
kubectl --namespace=kube-system get pods

echo "Installing ingress into kubectl cluster... DONE!"
