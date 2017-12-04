# start

Command `start` is used to invoke a container with the container name or ID.

## Description

The `pouch start` command starts a container in `created` or `exited` state and runs the user-specified program. `container name` or `container ID` must be specified. If the container doesn't exist, this command will return error.

## Help Information

Use the following command to get the usage of `pouch start`:

```
# pouch start -h
# pouch start --help
```

## Flag Guidance

### -a, --attach        

Attach the console to cotainer's STDOUT and STDERR. 

### -i, --interactive

Attach the console to container's STDIN. This flag is often used along with `-a` flag as following:

```
#pouch start ${containerID} -a -i
/ # ls /
bin   dev   etc   home  proc  root  run   sys   tmp   usr   var
/ # exit
```

### -h, --help

User can get help information via `-h` flag.

### Global Flags

All CLI side commands have the same global flags. For more details about global flags, please refer to [Global Flags](pouch.md#global-flags).
