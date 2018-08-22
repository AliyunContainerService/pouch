# Introducing

Daemon config file is a entry for user to set flags for pouchd. PouchContainer
support two ways for users to pass flags to daemon, one is run pouchd
directly with flags specified, like: `pouchd -c /var/run/containerd.sock`, the
other one is using daemon config file, and of course, you can use them simultaneously.

## Support daemon flags

We list some flags pouchd supports, for the detail flags explanations, you
can find in [pouchd flags](https://github.com/alibaba/pouch/blob/master/docs/commandline/pouchd.md).

| Flag                  | Description                             |
|-----------------------|-----------------------------------------|
| `-c`, `--containerd`  | where does containerd listened on. |
| `-l`, `--listen`      | which address to listen on.            |

## Configuring pouchd config file

We recommend users to set daemon flag through daemon config file, the default
path to config file is `/etc/pouch/config.json`, you can change it by set
value of `--config-file`.

### Note

* The same flag(exclude slice or array type) can not be set through command
  line and config file simultaneously.
* We allow users set slice or array type of flag simultaneously from command
  and config file line，and merge them.

### Steps to configure config file

1. Install PouchContainer, you can find detail steps in [PouchContainer install](https://github.com/alibaba/pouch/blob/master/INSTALLATION.md).
2. Edit daemon config file, like:

```
{
    "image-proxy": "http://127.0.0.1:65001",
    "debug": false
}
```

3. Start pouchd.
