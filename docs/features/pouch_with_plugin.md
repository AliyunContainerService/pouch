# Pouch with plugin
In order to run custom code at some point, we support a plugin frame work which introduced from golang 1.8. At this time in this plugin frame work we enable user to add custom code at four point:
* pre-start daemon point
* pre-stop daemon point
* pre-create container point
* pre-start container point

Above four point orgnized by two Plugin, which are DaemonPlugin and ContainerPlugin:
```
// DaemonPlugin defines in which place does pouch daemon support plugin
type DaemonPlugin interface {
	// PreStartHook is invoked by pouch daemon before real start, in this hook user could start dfget proxy or other
	// standalone process plugins
	PreStartHook() error

	// PreStopHook is invoked by pouch daemon before daemon process exit, not a promise if daemon is killed, in this
	// hook user could stop the process or plugin started by PreStartHook
	PreStopHook() error
}

// ContainerPlugin defines in which place a plugin will be triggered in container lifecycle
type ContainerPlugin interface {
	// PreCreate defines plugin point where recevives an container create request, in this plugin point user
	// could change the container create body passed-in by http request body
	PreCreate(io.ReadCloser) (io.ReadCloser, error)

	// PreStart returns an array of priority and args which will pass to runc, the every priority
	// used to sort the pre start array that pass to runc, network plugin hook always has priority value 0.
	PreStart(interface{}) ([]int, [][]string, error)
}

```
These two Plugin symbol will be fetch by name `DaemonPlugin` and `ContainerPlugin` like this:
```
daemonPlugin, _ := p.Lookup("DaemonPlugin")
containerPlugin, _ := p.Lookup("ContainerPlugin")
```

## example
define two plugin only print some log at correspond point
```
package main

import (
	"fmt"
	"io"
)

var ContainerPlugin ContPlugin

type ContPlugin int

var DaemonPlugin DPlugin

type DPlugin int

func (d DPlugin) PreStartHook() error {
	fmt.Println("pre-start hook in daemon is called")
	return nil
}

func (d DPlugin) PreStopHook() error {
	fmt.Println("pre-stop hook in daemon is called")
	return nil
}

func (c ContPlugin) PreCreate(in io.ReadCloser) (io.ReadCloser, error) {
	fmt.Println("pre create method called")
	return in, nil
}

func (c ContPlugin) PreStart(interface{}) ([]int, [][]string, error) {
	fmt.Println("pre start method called")
	return []int{-4}, [][]string{{"/usr/bin/touch", "touch", "/tmp/pre_start_hook"}}, nil
}

func main() {
	fmt.Println(ContainerPlugin, DaemonPlugin)
}
```
then build it with command line like:
```
go build -buildmode=plugin -ldflags "-pluginpath=plugins_$(date +%s)" -o hook_plugin.so
```
to use the shared object file generated, start pouchd which falg `--plugin=path_to_hook_plugin.so`, then when you start, stop daemon and create container in the log there will be some logs like:
```
pre-start hook in daemon is called
pre create method called
pre-stop hook in daemon is called
```
when you start a container, the config.json file (whose place is $home_dir/containerd/state/io.containerd.runtime.v1.linux/default/$container_id/config.json) will contains the pre-start hook you specified in your code, eg:
```
    "hooks": {
        "prestart": [
            {
                "args": [
                    "libnetwork-setkey",
                    "f67df14e96fa4b94a6e386d0795bdd2703ca7b01713d48c9567203a37b05ae3d",
                    "8e3d8db7f72a66edee99d4db6ab911f8d618af057485731e9acf24b3668e25b6"
                ],
                "path": "/usr/local/bin/pouchd"
            },
            {
                "args": [
                    "touch",
                    "/tmp/pre_start_hook"
                ],
                "path": "/usr/bin/touch"
            }
        ]
    }
```

and if you use the exact code above, after starting a container the file at /tmp/pre_start_hook will be touched.

## usage

at pre-start daemon point you can start assit processes like network plugins and dfget proxy which need by pouchd and life cycle is the same as pouchd.
at pre-stop daemon point you can stop the assist processes gracefully, but this is point is not a promise, because pouchd may be killed by SIGKILL.
at pre-create container point you can change the input stream by some rule, in some company they have some stale orchestration system who use env to pass-in some limit which is an attribute in pouch, then you can use this point to convert value in env to attribute in ContainerConfig or HostConfig.
at pre-start container point you can set more pre-start hooks to oci spec, where you can do some special thing before container entrypoint start, priority decide the order of executing the hook. libnetwork hook has priority 0, so if the hook is expected to run before network in container setup you should set priority to a value big then 0, and vice versa.