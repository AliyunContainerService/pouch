## pouch search

Search the images from specific registry

### Synopsis


Search the images from specific registry.

```
pouch search [OPTIONS] TERM
```

### Examples

```
$ pouch search nginx
NAME                                                   DESCRIPTION                                     STARS               OFFICIAL            AUTOMATED
nginx                                                  Official build of Nginx.                        11403               [OK]
jwilder/nginx-proxy                                    Automated Nginx reverse proxy for docker con…   1600                                    [OK]
richarvey/nginx-php-fpm                                Container running Nginx + PHP-FPM capable of…   712                                     [OK]
jrcs/letsencrypt-nginx-proxy-companion                 LetsEncrypt container to use with nginx as p…   509                                     [OK]
webdevops/php-nginx                                    Nginx with PHP-FPM                              127                                     [OK]
zabbix/zabbix-web-nginx-mysql                          Zabbix frontend based on Nginx web-server wi…   101                                     [OK]
bitnami/nginx                                          Bitnami nginx Docker Image                      66                                      [OK]
linuxserver/nginx                                      An Nginx container, brought to you by LinuxS…   61
1and1internet/ubuntu-16-nginx-php-phpmyadmin-mysql-5   ubuntu-16-nginx-php-phpmyadmin-mysql-5          50                                      [OK]
zabbix/zabbix-web-nginx-pgsql                          Zabbix frontend based on Nginx with PostgreS…   33                                      [OK]
tobi312/rpi-nginx                                      NGINX on Raspberry Pi / ARM                     26                                      [OK]
nginx/nginx-ingress                                    NGINX Ingress Controller for Kubernetes         20
schmunk42/nginx-redirect                               A very simple container to redirect HTTP tra…   15                                      [OK]
nginxdemos/hello                                       NGINX webserver that serves a simple page co…   14                                      [OK]
blacklabelops/nginx                                    Dockerized Nginx Reverse Proxy Server.          12                                      [OK]
wodby/drupal-nginx                                     Nginx for Drupal container image                12                                      [OK]
centos/nginx-18-centos7                                Platform for running nginx 1.8 or building n…   10
centos/nginx-112-centos7                               Platform for running nginx 1.12 or building …   9
nginxinc/nginx-unprivileged                            Unprivileged NGINX Dockerfiles                  4
1science/nginx                                         Nginx Docker images that include Consul Temp…   4                                       [OK]
nginx/nginx-prometheus-exporter                        NGINX Prometheus Exporter                       4
mailu/nginx                                            Mailu nginx frontend                            3                                       [OK]
toccoag/openshift-nginx                                Nginx reverse proxy for Nice running on same…   1                                       [OK]
ansibleplaybookbundle/nginx-apb                        An APB to deploy NGINX                          0                                       [OK]
wodby/nginx                                            Generic nginx                                   0                                       [OK]

```

### Options

```
  -h, --help              help for search
  -r, --registry string   set registry name
```

### Options inherited from parent commands

```
  -D, --debug              Switch client log level to DEBUG mode
  -H, --host string        Specify connecting address of Pouch CLI (default "unix:///var/run/pouchd.sock")
      --tlscacert string   Specify CA file of TLS
      --tlscert string     Specify cert file of TLS
      --tlskey string      Specify key file of TLS
      --tlsverify          Use TLS and verify remote
```

### SEE ALSO

* [pouch](pouch.md)	 - An efficient container engine

