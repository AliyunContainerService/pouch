package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

var searchDescription = "\nSearch the images from specific registry."

// SearchCommand implements search images.
type SearchCommand struct {
	baseCommand
	registry string
}

// Init initialize search command.
func (s *SearchCommand) Init(c *Cli) {
	s.cli = c

	s.cmd = &cobra.Command{
		Use:   "search [OPTIONS] TERM",
		Short: "Search the images from specific registry",
		Long:  searchDescription,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return s.runSearch(args)
		},
		Example: searchExample(),
	}
	s.addFlags()
}

// addFlags adds flags for specific command.
func (s *SearchCommand) addFlags() {
	flagSet := s.cmd.Flags()

	flagSet.StringVarP(&s.registry, "registry", "r", "", "set registry name")
}

func (s *SearchCommand) runSearch(args []string) error {
	ctx := context.Background()
	apiClient := s.cli.Client()

	term := args[0]

	// TODO: add flags --filter、--format、--limit、--no-trunc
	searchResults, err := apiClient.ImageSearch(ctx, term, s.registry, fetchRegistryAuth(s.registry))

	if err != nil {
		return err
	}

	display := s.cli.NewTableDisplay()
	display.AddRow([]string{"NAME", "DESCRIPTION", "STARS", "OFFICIAL", "AUTOMATED"})

	for _, result := range searchResults {
		display.AddRow([]string{result.Name, result.Description, fmt.Sprint(result.StarCount), boolToOKOrNot(result.IsOfficial), boolToOKOrNot(result.IsAutomated)})
	}

	display.Flush()
	return nil
}

func boolToOKOrNot(isTrue bool) string {
	if isTrue {
		return "[OK]"
	}
	return ""
}

func searchExample() string {
	return `$ pouch search nginx
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
`
}
