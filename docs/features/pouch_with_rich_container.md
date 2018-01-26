# Rich Container

Rich container is a very useful container mode when containerizing applications. This mode helps technical staff to complete packaging fat applications almost with no effort. It provides efficient ways to equip more basic software or system services except for target application in a single container . Then applications in containers could be running as smoothly as usual in VM or physical machine. This is a more generlized application-centric mode. This mode brings no invasiveness at all to both developers and operators. Especially for operators, they could have abilities to maintain applications in container with all essential tools or service processes they may need as usual.

Rich container mode is not the default mode Pouch provides. It is an additional mode pouch brings to extend users' container experience. Users can still manage ordinary containers by switching rich container flag off.

## Scenario

Container technology and orchestration platforms have turned quite popular right now. They both offer much better environment for applications. Despite this, we have to say containerization is the first step for enterprises to embrace container-related technologies, such as container, orchestration, service mesh and so on. It is quite a real problem to move traditional application into containers. Although some simple applications are always showing friendly to container, more traditional and complicated enterprise applications may not so lucky. These traditional applications are usually coupled with underlying infrastructure, such as architecture of machine, old kernels, even certain software out of maintainence as well. Definitely, strong coupling is not everyone's dish. It is the initiator on the road of digital tranformation in enterprises. So, all the industry is seeking one possible way to work it out. The way docker provides is one, but not the best. In the past 7 years, Alibaba has also experienced the same issue. Fortunately, rich container mode is a much better way to handle this.

Developers have their own programming style. Their work is to create useful applications, not to design absolute decoupled ones, so they usually take advantages of tools or system services to make it. When containerizing these applications, it is quite weak if only setting one application one process in container. Rich container mode finds out ways to make users configure the inner startup sequence of processes in container including application and system services around.

Operators have a sacred duty to guard normal running of the applications. For the sake of business running in applications, technology must show enough respect for operator's tradition. Environment change is not a good message when debugging and solving issue online. Rich container mode can ensure that environment in rich container in totally the same as that in traditional VM or physical machine. If operator needs some system tools, they are located there still. If some pre and post hooks should take effect, just set them when starting rich containers. If some issues happen inside, system services started by rich container can fix them just like self-healing.   

## Get started

Users can start rich container mode in Pouch quite easily. Provided that we need to running an ordinary image in rich container mode via Pouch, there are only two flags we may add: `--rich` and `--initscript`. Here are more description about both flags:

* `--rich`: identifies whether to switch on rich container mode or not. This flag has a type of `boolean`, and the default value is `false`.
* `--initscript`: identifies initial script executed in container. The script will be executed before entrypoint or command. Sometimes, it is called prestart hook. Lots of work can be done in this prestart hook, such as environment checking, environment preparation, network routes preparation, all kinds of agent settings, security setting and so on. This initscript may fail and user gets an related error message, if pouch daemon cannot find this initscript in container's filesystem which is provided by the rootfs constructed from image and potential mount volumes actually outside the container. If initscript works fine, the control of container process would be taken over by process pid 1, mainly `/sbin/init` or `dumbinit`.

In fact, pouch team plans to add another flag `--initcmd` to make users input prestart hook. Actually it is a simplified one of `--initscript`. Meanwhile it brings more convenience than `--initscript`. `--initcmd` can set any command as user's wish, and things do not need to be located in image in advance. We can say command is decoupled with image. But for `--initscript`, script file must be located in image first. It is some kind of coupling.

Here is a really simple example for rich container mode:

``` shell
pouch run --rich --initscript /home/root/startup.sh richapp:v1
```

If user specifies `--rich` flag and no `--initscript` flag is provided, rich container mode will still be enbaled, but no initscript will be executed. If `-rich` flag misses in command line, while `--initscript` is there, Pouch CLI or pouch daemon will return an error to show that `--initscipt` can only be used along with `--rich` flag.

If a container is running with `--rich` flag, then every start or restart of this container will trigger the corresponding initscipt if there is any. 

## Underlying Implementation

Before learning underlying implementation we shall take a brief review of `systemd`, `entrypoint` and `cmd`. In addition, prestart hook is executed by runC.

### systemd, entrypoint and cmd

To be added

### initscript and runC

`initscript` is to be added.

`runc` is a CLI tool for spawning and running containers according to the OCI specification.