#!/usr/bin/env bash

set -euo pipefail

CNI_VERSION=v0.7

# keep the first one only
GOPATH="${GOPATH%%:*}"

# add bin folder into PATH.
export PATH="${GOPATH}/bin:${PATH}"


# cni::install_cni installs cni plugins.
cni::install_cni() {
  echo "install cni..."

  local workdir pkg

  # for multiple GOPATHs, keep the first one only
  pkg="github.com/containernetworking/plugins"
  workdir="${GOPATH}/src/${pkg}"

  # downloads github.com/containernetworking/plugins
  if [ ! -d "${workdir}" ]; then
    mkdir -p "${workdir}"
    cd "${workdir}"
    git clone https://${pkg}.git .
  fi
  cd "${workdir}"
  git fetch --all
  git checkout "${CNI_VERSION}"

  # build and copy into /opt/cni/bin
  "${workdir}"/build.sh
  mkdir -p /etc/cni/net.d /opt/cni/bin
  cp "${workdir}"/bin/* /opt/cni/bin

  # setup the config
  sh -c 'cat >/etc/cni/net.d/10-mynet.conflist <<-EOF
{
    "cniVersion": "0.3.1",
    "name": "mynet",
    "plugins": [
        {
            "type": "bridge",
            "bridge": "cni0",
            "isGateway": true,
            "ipMasq": true,
            "ipam": {
                "type": "host-local",
                "subnet": "10.30.0.0/16",
                "routes": [
                    { "dst": "0.0.0.0/0"   }
                ]
            }
        },
        {
            "type": "portmap",
            "capabilities": {"portMappings": true},
            "snat": true
        }
    ]
}
EOF'

  sh -c 'cat >/etc/cni/net.d/99-loopback.conf <<-EOF
{
    "cniVersion": "0.3.1",
    "type": "loopback"
}
EOF'

  echo
}

main() {
  cni::install_cni
}

main "$@"
