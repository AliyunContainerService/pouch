## pouch top

Display the running processes of a container

### Synopsis

top comand is to display the running processes of a container.Your can add options just like using Linux ps command.

```
pouch top CONTAINER [ps OPTIONS]
```

### Examples

```
$ pouch top 44f675
	UID     PID      PPID     C    STIME    TTY    TIME        CMD
	root    28725    28714    0    3月14     ?      00:00:00    sh
	
```

### Options

```
  -h, --help   help for top
```

### Options inherited from parent commands

```
  -H, --host string        Specify connecting address of Pouch CLI (default "unix:///var/run/pouchd.sock")
      --tlscacert string   Specify CA file of TLS
      --tlscert string     Specify cert file of TLS
      --tlskey string      Specify key file of TLS
      --tlsverify          Use TLS and verify remote
```

### SEE ALSO

* [pouch](pouch.md)	 - An efficient container engine

