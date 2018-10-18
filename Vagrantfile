# -*- mode: ruby -*-
# vi: set ft=ruby :

# All Vagrant configuration is done below. The "2" in Vagrant.configure
# configures the configuration version (we support older styles for
# backwards compatibility). Please don't change it unless you know what
# you're doing.

VAGRANTFILE_API_VERSION = "2"
Vagrant.require_version ">= 1.7.4"

Vagrant.configure("2") do |config|
  config.vm.define :pouch do |pouch|
    pouch.vm.hostname = "pouch"
    pouch.vm.box = "ubuntu/xenial64"
    pouch.vm.provider 'virtualbox' do |v, override|
      v.memory = 2048
    end

    pouch.vm.provision "shell", inline: <<-SHELL
      until apt-get update; do echo "Waiting apt-get for 3 seconds..."; sleep 3; done
      apt-get --no-install-recommends install lxcfs
      apt-get --no-install-recommends install curl apt-transport-https ca-certificates software-properties-common
      curl -fsSL http://mirrors.aliyun.com/opsx/pouch/linux/debian/opsx@service.alibaba.com.gpg.key | apt-key add -
      add-apt-repository "deb http://mirrors.aliyun.com/opsx/pouch/linux/debian/ pouch stable"
      apt-get update
      apt-get --no-install-recommends install pouch
      systemctl enable pouch
      systemctl start pouch
      echo "alias docker='pouch'" >> ~/.bashrc
    SHELL

    if ENV["POUCH_BUILD"] == "true"
      pouch.vm.provision "shell", inline: <<-SHELL
        # configring environments for pouch
        GO_VERSION=1.10.4
        GOROOT=/opt/go
        GOPATH=/root/go
        apt-get install -y --no-install-recommends build-essential
        wget --progress=bar:force:noscroll https://dl.google.com/go/go$GO_VERSION.linux-amd64.tar.gz -O /tmp/go$GO_VERSION.linux-amd64.tar.gz
        tar xf /tmp/go$GO_VERSION.linux-amd64.tar.gz -C /opt/
        echo "export GOROOT=$GOROOT" >> ~/.bashrc
        echo "export GOPATH=$GOPATH" >> ~/.bashrc
        echo "export PATH=$PATH:$GOROOT/bin:$GOPATH/bin" >> ~/.bashrc

        export PATH=$PATH:$GOROOT/bin:$GOPATH/bin
        mkdir -p $GOPATH/src/github.com/alibaba
        ln -s -f /vagrant $GOPATH/src/github.com/alibaba/pouch
        cd $GOPATH/src/github.com/alibaba/pouch
        make && make PREFIX=/usr install
        systemctl restart pouch
      SHELL
    end

    if ENV["POUCH_KUBE"] == "true"
      env = {
        "KUBERNETES_VERSION"  => "1.10",
        "CRI_VERSION"         => "v1alpha2",
        "RELEASE_UBUNTU"      => "v1.10.2",
        "MASTER_NODE"         => "true",
        "INSTALL_FLANNEL"     => "true",
        "INSTALL_SAMPLE"      => "true"
      }

      if ENV["http_proxy"] != ""
        proxy =  ENV["http_proxy"]

        env["http_proxy"]   = proxy
        env["https_proxy"]  = proxy
        env["no_proxy"]     = "localhost,127.0.0.1,10.96.0.0/16,10.0.0.0/16,10.244.0.0/16"
      end

      pouch.vm.provision "install_kubernetes", type: "shell", path: "hack/kubernetes/allinone_aliyun.sh", env: env
    end
  end
end
