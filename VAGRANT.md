# Vagrant support for pouch

You can using Vagrant to quickly experience pouch or cross compile on non-linux.

## Requirements

* Vagrant 1.9.x or newer
* VirtuaBox

## Get started

```bash
vagrant up
vagrant ssh -c "sudo su -l"

# Start a nginx container with 80
pouch run -d --name nginx -p 80:80 nginx
curl http://localhost
```

## Build pouch with vagrant

```bash
export POUCH_BUILD=true # set POUCH_BUILD=true on Windows
vagrant up

# Install compiled pouch binarys for pouch service.
vagrant ssh -c "sudo su -l"
cd ~/go/src/github.com/alibaba/pouch
make DEST_DIR=/usr install
systemctl restart pouch
pouch version
```