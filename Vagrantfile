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
    pouch.vm.box_url="https://cloud-images.ubuntu.com/xenial/current/xenial-server-cloudimg-amd64-vagrant.box"
    pouch.vm.box = "ubuntu_16.04"
    pouch.vm.provision "shell", inline: <<-SHELL
      until apt-get update &> /dev/null; do echo "Waiting apt-get for 3 seconds..."; sleep 3; done
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
        GOPATH=/root/go
        apt-get install -y --no-install-recommends build-essential
        wget --progress=bar:force:noscroll https://dl.google.com/go/go1.10.3.linux-amd64.tar.gz -O /tmp/go1.10.3.linux-amd64.tar.gz
        tar xf /tmp/go1.10.3.linux-amd64.tar.gz -C /opt/
        echo "export GOROOT=/opt/go" >> ~/.bashrc
        echo "export GOPATH=$GOPATH" >> ~/.bashrc
        cd /usr/bin && find /opt/go/bin -type f | xargs -n1 ln -s

        mkdir -p $GOPATH/src/github.com/alibaba
        ln -s /vagrant $GOPATH/src/github.com/alibaba/pouch
        cd $GOPATH/src/github.com/alibaba/pouch && make
      SHELL
    end
  end
end
