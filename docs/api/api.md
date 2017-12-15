# Pouch Engine API


<a name="overview"></a>
## Overview
API is an HTTP API served by Pouch Engine.


### Version information
*Version* : 1.24


### URI scheme
*BasePath* : /v1.24  
*Schemes* : HTTP, HTTPS


### Consumes

* `application/json`
* `text/plain`


### Produces

* `application/json`




<a name="paths"></a>
## Paths

<a name="ping-get"></a>
### GET /_ping

#### Responses

|HTTP Code|Description|Schema|
|---|---|---|
|**200**|no error|string|
|**500**|server error|[Error](#error)|


#### Example HTTP response

##### Response 200
```
json :
"OK"
```


<a name="containers-create-post"></a>
### Create a container
```
POST /containers/create
```


#### Parameters

|Type|Name|Description|Schema|
|---|---|---|---|
|**Query**|**name**  <br>*optional*|Assign the specified name to the container. Must match `/?[a-zA-Z0-9_-]+`.|string|
|**Body**|**body**  <br>*required*|Container to create|[body](#containers-create-post-body)|

<a name="containers-create-post-body"></a>
**body**

|Name|Description|Schema|
|---|---|---|
|**ArgsEscaped**  <br>*optional*|Command is already escaped (Windows only)|boolean|
|**AttachStderr**  <br>*optional*|Whether to attach to `stderr`.  <br>**Default** : `true`|boolean|
|**AttachStdin**  <br>*optional*|Whether to attach to `stdin`.  <br>**Default** : `false`|boolean|
|**AttachStdout**  <br>*optional*|Whether to attach to `stdout`.  <br>**Default** : `true`|boolean|
|**Cmd**  <br>*optional*|Command to run specified as a string or an array of strings.|< string > array|
|**Domainname**  <br>*optional*|The domain name to use for the container.|string|
|**Entrypoint**  <br>*optional*|The entry point for the container as a string or an array of strings.<br>If the array consists of exactly one empty string (`[""]`) then the entry point is reset to system default (i.e., the entry point used by docker when there is no `ENTRYPOINT` instruction in the `Dockerfile`).|< string > array|
|**Env**  <br>*optional*|A list of environment variables to set inside the container in the form `["VAR=value", ...]`. A variable without `=` is removed from the environment, rather than to have an empty value.|< string > array|
|**ExposedPorts**  <br>*optional*|An object mapping ports to an empty object in the form:<br><br>`{"<port>/<tcp\|udp>": {}}`|< string, object > map|
|**HostConfig**  <br>*optional*||[HostConfig](#hostconfig)|
|**Hostname**  <br>*optional*|The hostname to use for the container, as a valid RFC 1123 hostname.|string|
|**Image**  <br>*optional*|The name of the image to use when creating the container|string|
|**Labels**  <br>*optional*|User-defined key/value metadata.|< string, string > map|
|**MacAddress**  <br>*optional*|MAC address of the container.|string|
|**NetworkDisabled**  <br>*optional*|Disable networking for the container.|boolean|
|**NetworkingConfig**  <br>*optional*|This container's networking configuration.|[NetworkingConfig](#containers-create-post-networkingconfig)|
|**OnBuild**  <br>*optional*|`ONBUILD` metadata that were defined in the image's `Dockerfile`.|< string > array|
|**OpenStdin**  <br>*optional*|Open `stdin`  <br>**Default** : `false`|boolean|
|**Shell**  <br>*optional*|Shell for when `RUN`, `CMD`, and `ENTRYPOINT` uses a shell.|< string > array|
|**StdinOnce**  <br>*optional*|Close `stdin` after one attached client disconnects  <br>**Default** : `false`|boolean|
|**StopSignal**  <br>*optional*|Signal to stop a container as a string or unsigned integer.  <br>**Default** : `"SIGTERM"`|string|
|**StopTimeout**  <br>*optional*|Timeout to stop a container in seconds.|integer|
|**Tty**  <br>*optional*|Attach standard streams to a TTY, including `stdin` if it is not closed.  <br>**Default** : `false`|boolean|
|**User**  <br>*optional*|The user that commands are run as inside the container.|string|
|**Volumes**  <br>*optional*|An object mapping mount point paths inside the container to empty objects.|[Volumes](#containers-create-post-volumes)|
|**WorkingDir**  <br>*optional*|The working directory for commands to run in.|string|

<a name="containers-create-post-networkingconfig"></a>
**NetworkingConfig**

|Name|Description|Schema|
|---|---|---|
|**EndpointsConfig**  <br>*optional*|A mapping of network name to endpoint configuration for that network.|< string, [EndpointSettings](#endpointsettings) > map|

<a name="containers-create-post-volumes"></a>
**Volumes**

|Name|Schema|
|---|---|
|**additionalProperties**  <br>*optional*|object|


#### Responses

|HTTP Code|Description|Schema|
|---|---|---|
|**201**|Container created successfully|[ContainerCreateResp](#containercreateresp)|
|**400**|bad parameter|[Error](#error)|
|**404**|no such container|[Error](#error)|
|**409**|conflict|[Error](#error)|
|**500**|server error|[Error](#error)|


#### Consumes

* `application/json`


#### Produces

* `application/json`


#### Tags

* Container


#### Example HTTP response

##### Response 201
```
json :
{
  "Id" : "e90e34656806",
  "Warnings" : [ ]
}
```


##### Response 404
```
json :
{
  "message" : "No such container: c2ada9df5af8"
}
```


<a name="containerlist"></a>
### List containers
```
GET /containers/json
```


#### Responses

|HTTP Code|Description|Schema|
|---|---|---|
|**200**|Summary containers that matches the query|[Container](#container)|
|**500**|Server error|[Error](#error)|


#### Produces

* `application/json`


<a name="containerinspect"></a>
### Inspect a container
```
GET /containers/{id}/json
```


#### Description
Return low-level information about a container.


#### Parameters

|Type|Name|Description|Schema|Default|
|---|---|---|---|---|
|**Path**|**id**  <br>*required*|ID or name of the container|string||
|**Query**|**size**  <br>*optional*|Return the size of container as fields `SizeRw` and `SizeRootFs`|boolean|`"false"`|


#### Responses

|HTTP Code|Description|Schema|
|---|---|---|
|**200**|no error|[ContainerJSON](#containerjson)|
|**404**|no such container|[Error](#error)|
|**500**|server error|[Error](#error)|


#### Produces

* `application/json`


#### Tags

* Container


#### Example HTTP response

##### Response 404
```
json :
{
  "message" : "No such container: c2ada9df5af8"
}
```


<a name="containerpause"></a>
### Pause a container
```
POST /containers/{id}/pause
```


#### Parameters

|Type|Name|Description|Schema|
|---|---|---|---|
|**Path**|**id**  <br>*required*|ID or name of the container|string|


#### Responses

|HTTP Code|Description|Schema|
|---|---|---|
|**204**|no error|No Content|
|**404**|no such container|[Error](#error)|
|**500**|server error|[Error](#error)|


#### Tags

* Container


<a name="containerrename"></a>
### Rename a container
```
POST /containers/{id}/rename
```


#### Parameters

|Type|Name|Description|Schema|
|---|---|---|---|
|**Path**|**id**  <br>*required*|ID or name of the container|string|
|**Query**|**name**  <br>*required*|New name for the container|string|


#### Responses

|HTTP Code|Description|Schema|
|---|---|---|
|**204**|no error|No Content|
|**404**|no such container|[Error](#error)|
|**409**|name already in use|[Error](#error)|
|**500**|server error|[Error](#error)|


#### Tags

* Container


<a name="containerremove"></a>
### Remove one container
```
DELETE /containers/{name}
```


#### Parameters

|Type|Name|Description|Schema|
|---|---|---|---|
|**Path**|**name**  <br>*required*|ID or name of the container|string|


#### Responses

|HTTP Code|Description|Schema|
|---|---|---|
|**204**|no error|No Content|
|**500**|server error|[Error](#error)|


#### Tags

* Container


<a name="images-create-post"></a>
### Create an image by pulling from a registry or importing from an existing source file
```
POST /images/create
```


#### Parameters

|Type|Name|Description|Schema|
|---|---|---|---|
|**Header**|**X-Registry-Auth**  <br>*optional*|A base64-encoded auth configuration. [See the authentication section for details.](#section/Authentication)|string|
|**Query**|**fromImage**  <br>*optional*|Name of the image to pull. The name may include a tag or digest. This parameter may only be used when pulling an image. The pull is cancelled if the HTTP connection is closed.|string|
|**Query**|**fromSrc**  <br>*optional*|Source to import. The value may be a URL from which the image can be retrieved or `-` to read the image from the request body. This parameter may only be used when importing an image.|string|
|**Query**|**repo**  <br>*optional*|Repository name given to an image when it is imported. The repo may include a tag. This parameter may only be used when importing an image.|string|
|**Query**|**tag**  <br>*optional*|Tag or digest. If empty when pulling an image, this causes all tags for the given image to be pulled.|string|
|**Body**|**inputImage**  <br>*optional*|Image content if the value `-` has been specified in fromSrc query parameter|string|


#### Responses

|HTTP Code|Description|Schema|
|---|---|---|
|**200**|no error|No Content|
|**404**|image not found|[Error](#error)|
|**500**|server error|[Error](#error)|


#### Consumes

* `text/plain`
* `application/octet-stream`


#### Produces

* `application/json`


<a name="imagelist"></a>
### List Images
```
GET /images/json
```


#### Description
Return a list of stored images.


#### Parameters

|Type|Name|Description|Schema|Default|
|---|---|---|---|---|
|**Query**|**all**  <br>*optional*|Show all images. Only images from a final layer (no children) are shown by default.|boolean|`"false"`|
|**Query**|**digests**  <br>*optional*|Show digest information as a `RepoDigests` field on each image.|boolean|`"false"`|
|**Query**|**filters**  <br>*optional*|A JSON encoded value of the filters (a `map[string][]string`) to process on the images list. Available filters:<br><br>- `before`=(`<image-name>[:<tag>]`,  `<image id>` or `<image@digest>`)<br>- `dangling=true`<br>- `label=key` or `label="key=value"` of an image label<br>- `reference`=(`<image-name>[:<tag>]`)<br>- `since`=(`<image-name>[:<tag>]`,  `<image id>` or `<image@digest>`)|string||


#### Responses

|HTTP Code|Description|Schema|
|---|---|---|
|**200**|Summary image data for the images matching the query|< [ImageInfo](#imageinfo) > array|
|**500**|server error|[Error](#error)|


#### Produces

* `application/json`


#### Example HTTP response

##### Response 200
```
json :
[ {
  "Id" : "sha256:e216a057b1cb1efc11f8a268f37ef62083e70b1b38323ba252e25ac88904a7e8",
  "ParentId" : "",
  "RepoTags" : [ "ubuntu:12.04", "ubuntu:precise" ],
  "RepoDigests" : [ "ubuntu@sha256:992069aee4016783df6345315302fa59681aae51a8eeb2f889dea59290f21787" ],
  "Created" : 1474925151,
  "Size" : 103579269,
  "VirtualSize" : 103579269,
  "SharedSize" : 0,
  "Labels" : { },
  "Containers" : 2
}, {
  "Id" : "sha256:3e314f95dcace0f5e4fd37b10862fe8398e3c60ed36600bc0ca5fda78b087175",
  "ParentId" : "",
  "RepoTags" : [ "ubuntu:12.10", "ubuntu:quantal" ],
  "RepoDigests" : [ "ubuntu@sha256:002fba3e3255af10be97ea26e476692a7ebed0bb074a9ab960b2e7a1526b15d7", "ubuntu@sha256:68ea0200f0b90df725d99d823905b04cf844f6039ef60c60bf3e019915017bd3" ],
  "Created" : 1403128455,
  "Size" : 172064416,
  "VirtualSize" : 172064416,
  "SharedSize" : 0,
  "Labels" : { },
  "Containers" : 5
} ]
```


<a name="images-search-get"></a>
### Search images
```
GET /images/search
```


#### Responses

|HTTP Code|Description|Schema|
|---|---|---|
|**200**|No error|< [SearchResultItem](#searchresultitem) > array|
|**500**|server error|[Error](#error)|


#### Produces

* `application/json`


<a name="images-name-delete"></a>
### Remove an image
```
DELETE /images/{name}
```


#### Description
Remove an image by reference.


#### Parameters

|Type|Name|Description|Schema|
|---|---|---|---|
|**Path**|**name**  <br>*required*|Image reference|string|


#### Responses

|HTTP Code|Description|Schema|
|---|---|---|
|**204**|No error|No Content|
|**500**|Server deletes an image error|[Error](#error)|


<a name="info-get"></a>
### Get System information
```
GET /info
```


#### Responses

|HTTP Code|Description|Schema|
|---|---|---|
|**200**|no error|[SystemInfo](#systeminfo)|
|**500**|server error|[Error](#error)|


<a name="version-get"></a>
### Get Pouchd version
```
GET /version
```


#### Responses

|HTTP Code|Description|Schema|
|---|---|---|
|**200**|no error|[SystemVersion](#systemversion)|
|**500**|server error|[Error](#error)|


<a name="volumelist"></a>
### List volumes
```
GET /volumes
```


#### Parameters

|Type|Name|Description|Schema|
|---|---|---|---|
|**Query**|**filters**  <br>*optional*|JSON encoded value of the filters (a `map[string][]string`) to<br>process on the volumes list. Available filters:<br><br>- `dangling=<boolean>` When set to `true` (or `1`), returns all<br>   volumes that are not in use by a container. When set to `false`<br>   (or `0`), only volumes that are in use by one or more<br>   containers are returned.<br>- `driver=<volume-driver-name>` Matches volumes based on their driver.<br>- `label=<key>` or `label=<key>:<value>` Matches volumes based on<br>   the presence of a `label` alone or a `label` and a value.<br>- `name=<volume-name>` Matches all or part of a volume name.|string (json)|


#### Responses

|HTTP Code|Description|Schema|
|---|---|---|
|**200**|Summary volume data that matches the query|[VolumeListResp](#volumelistresp)|
|**500**|Server error|[Error](#error)|


#### Produces

* `application/json`


#### Tags

* Volume


#### Example HTTP response

##### Response 200
```
json :
{
  "Volumes" : [ {
    "CreatedAt" : "2017-07-19T12:00:26Z",
    "Name" : "tardis",
    "Driver" : "local",
    "Mountpoint" : "/var/lib/docker/volumes/tardis",
    "Labels" : {
      "com.example.some-label" : "some-value",
      "com.example.some-other-label" : "some-other-value"
    },
    "Scope" : "local",
    "Options" : {
      "device" : "tmpfs",
      "o" : "size=100m,uid=1000",
      "type" : "tmpfs"
    }
  } ],
  "Warnings" : [ ]
}
```


<a name="volumecreate"></a>
### Create a volume
```
POST /volumes/create
```


#### Parameters

|Type|Name|Description|Schema|
|---|---|---|---|
|**Body**|**VolumeConfig**  <br>*required*|Volume configuration|[VolumeConfig](#volumeconfig)|


#### Responses

|HTTP Code|Description|Schema|
|---|---|---|
|**201**|The volume was created successfully|[VolumeInfo](#volumeinfo)|
|**500**|Server error|[Error](#error)|


#### Consumes

* `application/json`


#### Produces

* `application/json`


#### Tags

* Volume


#### Example HTTP request

##### Request body
```
json :
{
  "Name" : "tardis",
  "Labels" : {
    "com.example.some-label" : "some-value",
    "com.example.some-other-label" : "some-other-value"
  },
  "Driver" : "custom"
}
```




<a name="definitions"></a>
## Definitions

<a name="container"></a>
### Container
Container contains response of Engine API:
GET "/containers/json""


|Name|Description|Schema|
|---|---|---|
|**Command**  <br>*optional*||string|
|**Created**  <br>*optional*||string|
|**HostConfig**  <br>*optional*||[HostConfig](#hostconfig)|
|**ID**  <br>*optional*||string|
|**Image**  <br>*optional*||string|
|**ImageID**  <br>*optional*||string|
|**Labels**  <br>*optional*||< string, string > map|
|**Names**  <br>*optional*|**Example** : `[ "container_1", "container_2" ]`|< string > array|
|**SizeRootFs**  <br>*optional*||integer (int64)|
|**SizeRw**  <br>*optional*||integer (int64)|
|**State**  <br>*optional*||string|
|**Status**  <br>*optional*||string|


<a name="containerconfig"></a>
### ContainerConfig
Configuration for a container that is portable between hosts


|Name|Description|Schema|
|---|---|---|
|**ArgsEscaped**  <br>*optional*|Command is already escaped (Windows only)|boolean|
|**AttachStderr**  <br>*optional*|Whether to attach to `stderr`.  <br>**Default** : `true`|boolean|
|**AttachStdin**  <br>*optional*|Whether to attach to `stdin`.  <br>**Default** : `false`|boolean|
|**AttachStdout**  <br>*optional*|Whether to attach to `stdout`.  <br>**Default** : `true`|boolean|
|**Cmd**  <br>*optional*|Command to run specified as a string or an array of strings.|< string > array|
|**Domainname**  <br>*optional*|The domain name to use for the container.|string|
|**Entrypoint**  <br>*optional*|The entry point for the container as a string or an array of strings.<br>If the array consists of exactly one empty string (`[""]`) then the entry point is reset to system default (i.e., the entry point used by docker when there is no `ENTRYPOINT` instruction in the `Dockerfile`).|< string > array|
|**Env**  <br>*optional*|A list of environment variables to set inside the container in the form `["VAR=value", ...]`. A variable without `=` is removed from the environment, rather than to have an empty value.|< string > array|
|**ExposedPorts**  <br>*optional*|An object mapping ports to an empty object in the form:<br><br>`{"<port>/<tcp\|udp>": {}}`|< string, object > map|
|**Hostname**  <br>*optional*|The hostname to use for the container, as a valid RFC 1123 hostname.|string|
|**Image**  <br>*optional*|The name of the image to use when creating the container|string|
|**Labels**  <br>*optional*|User-defined key/value metadata.|< string, string > map|
|**MacAddress**  <br>*optional*|MAC address of the container.|string|
|**NetworkDisabled**  <br>*optional*|Disable networking for the container.|boolean|
|**OnBuild**  <br>*optional*|`ONBUILD` metadata that were defined in the image's `Dockerfile`.|< string > array|
|**OpenStdin**  <br>*optional*|Open `stdin`  <br>**Default** : `false`|boolean|
|**Shell**  <br>*optional*|Shell for when `RUN`, `CMD`, and `ENTRYPOINT` uses a shell.|< string > array|
|**StdinOnce**  <br>*optional*|Close `stdin` after one attached client disconnects  <br>**Default** : `false`|boolean|
|**StopSignal**  <br>*optional*|Signal to stop a container as a string or unsigned integer.  <br>**Default** : `"SIGTERM"`|string|
|**StopTimeout**  <br>*optional*|Timeout to stop a container in seconds.|integer|
|**Tty**  <br>*optional*|Attach standard streams to a TTY, including `stdin` if it is not closed.  <br>**Default** : `false`|boolean|
|**User**  <br>*optional*|The user that commands are run as inside the container.|string|
|**Volumes**  <br>*optional*|An object mapping mount point paths inside the container to empty objects.|[Volumes](#containerconfig-volumes)|
|**WorkingDir**  <br>*optional*|The working directory for commands to run in.|string|

<a name="containerconfig-volumes"></a>
**Volumes**

|Name|Schema|
|---|---|
|**additionalProperties**  <br>*optional*|object|


<a name="containercreateresp"></a>
### ContainerCreateResp
response returned by daemon when container create successfully


|Name|Description|Schema|
|---|---|---|
|**Id**  <br>*required*|The ID of the created container|string|
|**Name**  <br>*optional*|Then name of the created container|string|
|**Warnings**  <br>*required*|Warnings encountered when creating the container|< string > array|


<a name="containerjson"></a>
### ContainerJSON

|Name|Description|Schema|
|---|---|---|
|**AppArmorProfile**  <br>*optional*||string|
|**Args**  <br>*optional*|The arguments to the command being run|< string > array|
|**Config**  <br>*optional*||[ContainerConfig](#containerconfig)|
|**Created**  <br>*optional*|The time the container was created|string|
|**Driver**  <br>*optional*||string|
|**ExecIDs**  <br>*optional*||string|
|**HostConfig**  <br>*optional*||[HostConfig](#hostconfig)|
|**HostnamePath**  <br>*optional*||string|
|**HostsPath**  <br>*optional*||string|
|**Id**  <br>*optional*|The ID of the container|string|
|**Image**  <br>*optional*|The container's image|string|
|**LogPath**  <br>*optional*||string|
|**MountLabel**  <br>*optional*||string|
|**Name**  <br>*optional*||string|
|**Path**  <br>*optional*|The path to the command being run|string|
|**ProcessLabel**  <br>*optional*||string|
|**ResolvConfPath**  <br>*optional*||string|
|**RestartCount**  <br>*optional*||integer|
|**SizeRootFs**  <br>*optional*|The total size of all the files in this container.|integer (int64)|
|**SizeRw**  <br>*optional*|The size of files that have been created or changed by this container.|integer (int64)|
|**State**  <br>*optional*|The state of the container.|[ContainerState](#containerstate)|


<a name="containerstate"></a>
### ContainerState

|Name|Description|Schema|
|---|---|---|
|**Dead**  <br>*optional*|Whether this container is dead.|boolean|
|**Error**  <br>*optional*|The error message of this container|string|
|**ExitCode**  <br>*optional*|The last exit code of this container|integer|
|**FinishedAt**  <br>*optional*|The time when this container last exited.|string|
|**OOMKilled**  <br>*optional*|Whether this container has been killed because it ran out of memory.|boolean|
|**Paused**  <br>*optional*|Whether this container is paused.|boolean|
|**Pid**  <br>*optional*|The process ID of this container|integer|
|**Restarting**  <br>*optional*|Whether this container is restarting.|boolean|
|**Running**  <br>*optional*|Whether this container is running.<br><br>Note that a running container can be _paused_. The `Running` and `Paused`<br>booleans are not mutually exclusive:<br><br>When pausing a container (on Linux), the cgroups freezer is used to suspend<br>all processes in the container. Freezing the process requires the process to<br>be running. As a result, paused containers are both `Running` _and_ `Paused`.<br><br>Use the `Status` field instead to determine if a container's state is "running".|boolean|
|**StartedAt**  <br>*optional*|The time when this container was last started.|string|
|**Status**  <br>*optional*||[Status](#status)|


<a name="endpointsettings"></a>
### EndpointSettings
Configuration for a network endpoint.


|Name|Description|Schema|
|---|---|---|
|**Aliases**  <br>*optional*|**Example** : `[ "server_x", "server_y" ]`|< string > array|
|**DriverOpts**  <br>*optional*|DriverOpts is a mapping of driver options and values. These options<br>are passed directly to the driver and are driver specific.  <br>**Example** : `{<br>  "com.example.some-label" : "some-value",<br>  "com.example.some-other-label" : "some-other-value"<br>}`|< string, string > map|
|**EndpointID**  <br>*optional*|Unique ID for the service endpoint in a Sandbox.  <br>**Example** : `"b88f5b905aabf2893f3cbc4ee42d1ea7980bbc0a92e2c8922b1e1795298afb0b"`|string|
|**Gateway**  <br>*optional*|Gateway address for this network.  <br>**Example** : `"172.17.0.1"`|string|
|**GlobalIPv6Address**  <br>*optional*|Global IPv6 address.  <br>**Example** : `"2001:db8::5689"`|string|
|**GlobalIPv6PrefixLen**  <br>*optional*|Mask length of the global IPv6 address.  <br>**Example** : `64`|integer (int64)|
|**IPAddress**  <br>*optional*|IPv4 address.  <br>**Example** : `"172.17.0.4"`|string|
|**IPPrefixLen**  <br>*optional*|Mask length of the IPv4 address.  <br>**Example** : `16`|integer|
|**IPv6Gateway**  <br>*optional*|IPv6 gateway address.  <br>**Example** : `"2001:db8:2::100"`|string|
|**Links**  <br>*optional*|**Example** : `[ "container_1", "container_2" ]`|< string > array|
|**MacAddress**  <br>*optional*|MAC address for the endpoint on this network.  <br>**Example** : `"02:42:ac:11:00:04"`|string|
|**NetworkID**  <br>*optional*|Unique ID of the network.  <br>**Example** : `"08754567f1f40222263eab4102e1c733ae697e8e354aa9cd6e18d7402835292a"`|string|


<a name="error"></a>
### Error

|Name|Schema|
|---|---|
|**message**  <br>*optional*|string|


<a name="execcreateconfig"></a>
### ExecCreateConfig

|Name|Description|Schema|
|---|---|---|
|**AttachStderr**  <br>*optional*|Attach the standard error|boolean|
|**AttachStdin**  <br>*optional*|Attach the standard input, makes possible user interaction|boolean|
|**AttachStdout**  <br>*optional*|Attach the standard output|boolean|
|**Cmd**  <br>*optional*|Execution commands and args|< string > array|
|**Detach**  <br>*optional*|Execute in detach mode|boolean|
|**DetachKeys**  <br>*optional*|Escape keys for detach|string|
|**Privileged**  <br>*optional*|Is the container in privileged mode|boolean|
|**Tty**  <br>*optional*|Attach standard streams to a tty|boolean|
|**User**  <br>*optional*|User that will run the command|string|


<a name="execcreateresponse"></a>
### ExecCreateResponse

|Name|Schema|
|---|---|
|**ID**  <br>*optional*|string|


<a name="execstartconfig"></a>
### ExecStartConfig

|Name|Schema|
|---|---|
|**Detach**  <br>*optional*|boolean|
|**Tty**  <br>*optional*|boolean|


<a name="hostconfig"></a>
### HostConfig
Container configuration that depends on the host we are running on

*Polymorphism* : Composition


|Name|Description|Schema|
|---|---|---|
|**AutoRemove**  <br>*optional*|Automatically remove the container when the container's process exits. This has no effect if `RestartPolicy` is set.|boolean|
|**Binds**  <br>*optional*|A list of volume bindings for this container. Each volume binding is a string in one of these forms:<br><br>- `host-src:container-dest` to bind-mount a host path into the container. Both `host-src`, and `container-dest` must be an _absolute_ path.<br>- `host-src:container-dest:ro` to make the bind mount read-only inside the container. Both `host-src`, and `container-dest` must be an _absolute_ path.<br>- `volume-name:container-dest` to bind-mount a volume managed by a volume driver into the container. `container-dest` must be an _absolute_ path.<br>- `volume-name:container-dest:ro` to mount the volume read-only inside the container.  `container-dest` must be an _absolute_ path.|< string > array|
|**CapAdd**  <br>*optional*|A list of kernel capabilities to add to the container.|< string > array|
|**CapDrop**  <br>*optional*|A list of kernel capabilities to drop from the container.|< string > array|
|**Cgroup**  <br>*optional*|Cgroup to use for the container.|string|
|**ConsoleSize**  <br>*optional*|Initial console size, as an `[height, width]` array. (Windows only)|< integer > array|
|**ContainerIDFile**  <br>*optional*|Path to a file where the container ID is written|string|
|**Dns**  <br>*optional*|A list of DNS servers for the container to use.|< string > array|
|**DnsOptions**  <br>*optional*|A list of DNS options.|< string > array|
|**DnsSearch**  <br>*optional*|A list of DNS search domains.|< string > array|
|**ExtraHosts**  <br>*optional*|A list of hostnames/IP mappings to add to the container's `/etc/hosts` file. Specified in the form `["hostname:IP"]`.|< string > array|
|**GroupAdd**  <br>*optional*|A list of additional groups that the container process will run as.|< string > array|
|**IpcMode**  <br>*optional*|IPC sharing mode for the container. Possible values are:<br><br>- `"none"`: own private IPC namespace, with /dev/shm not mounted<br>- `"private"`: own private IPC namespace<br>- `"shareable"`: own private IPC namespace, with a possibility to share it with other containers<br>- `"container:<name\|id>"`: join another (shareable) container's IPC namespace<br>- `"host"`: use the host system's IPC namespace<br><br>If not specified, daemon default is used, which can either be `"private"`<br>or `"shareable"`, depending on daemon version and configuration.|string|
|**Isolation**  <br>*optional*|Isolation technology of the container. (Windows only)|enum (default, process, hyperv)|
|**Links**  <br>*optional*|A list of links for the container in the form `container_name:alias`.|< string > array|
|**LogConfig**  <br>*optional*|The logging configuration for this container|[LogConfig](#hostconfig-logconfig)|
|**NetworkMode**  <br>*optional*|Network mode to use for this container. Supported standard values are: `bridge`, `host`, `none`, and `container:<name\|id>`. Any other value is taken as a custom network's name to which this container should connect to.|string|
|**OomScoreAdj**  <br>*optional*|An integer value containing the score given to the container in order to tune OOM killer preferences.  <br>**Example** : `500`|integer|
|**PidMode**  <br>*optional*|Set the PID (Process) Namespace mode for the container. It can be either:<br><br>- `"container:<name\|id>"`: joins another container's PID namespace<br>- `"host"`: use the host's PID namespace inside the container|string|
|**PortBindings**  <br>*optional*|A map of exposed container ports and the host port they should map to.|< string, [PortBindings](#hostconfig-portbindings) > map|
|**Privileged**  <br>*optional*|Gives the container full access to the host.|boolean|
|**PublishAllPorts**  <br>*optional*|Allocates a random host port for all of a container's exposed ports.|boolean|
|**ReadonlyRootfs**  <br>*optional*|Mount the container's root filesystem as read only.|boolean|
|**Runtime**  <br>*optional*|Runtime to use with this container.|string|
|**SecurityOpt**  <br>*optional*|A list of string values to customize labels for MLS systems, such as SELinux.|< string > array|
|**ShmSize**  <br>*optional*|Size of `/dev/shm` in bytes. If omitted, the system uses 64MB.  <br>**Minimum value** : `0`|integer|
|**StorageOpt**  <br>*optional*|Storage driver options for this container, in the form `{"size": "120G"}`.|< string, string > map|
|**Sysctls**  <br>*optional*|A list of kernel parameters (sysctls) to set in the container. For example: `{"net.ipv4.ip_forward": "1"}`|< string, string > map|
|**Tmpfs**  <br>*optional*|A map of container directories which should be replaced by tmpfs mounts, and their corresponding mount options. For example: `{ "/run": "rw,noexec,nosuid,size=65536k" }`.|< string, string > map|
|**UTSMode**  <br>*optional*|UTS namespace to use for the container.|string|
|**UsernsMode**  <br>*optional*|Sets the usernamespace mode for the container when usernamespace remapping option is enabled.|string|
|**VolumeDriver**  <br>*optional*|Driver that this container uses to mount volumes.|string|
|**VolumesFrom**  <br>*optional*|A list of volumes to inherit from another container, specified in the form `<container name>[:<ro\|rw>]`.|< string > array|

<a name="hostconfig-logconfig"></a>
**LogConfig**

|Name|Schema|
|---|---|
|**Config**  <br>*optional*|< string, string > map|
|**Type**  <br>*optional*|enum (json-file, syslog, journald, gelf, fluentd, awslogs, splunk, etwlogs, none)|

<a name="hostconfig-portbindings"></a>
**PortBindings**

|Name|Description|Schema|
|---|---|---|
|**HostIp**  <br>*optional*|The host IP address|string|
|**HostPort**  <br>*optional*|The host port number, as a string|string|


<a name="imageinfo"></a>
### ImageInfo
An object containing all details of an image at API side


|Name|Description|Schema|
|---|---|---|
|**CreatedAt**  <br>*optional*|Time of image creation|string|
|**Digest**  <br>*optional*|digest of image.|string|
|**ID**  <br>*optional*|ID of an image.|string|
|**Name**  <br>*optional*|name of an image.|string|
|**Size**  <br>*optional*|size of image's taking disk space.|integer|
|**Tag**  <br>*optional*|tag of an image.|string|


<a name="searchresultitem"></a>
### SearchResultItem
search result item in search results.


|Name|Description|Schema|
|---|---|---|
|**description**  <br>*optional*|description just shows the description of this image|string|
|**is_automated**  <br>*optional*|is_automated means whether this image is automated.|boolean|
|**is_official**  <br>*optional*|is_official shows if this image is marked official.|boolean|
|**name**  <br>*optional*|name represents the name of this image|string|
|**star_count**  <br>*optional*|star_count refers to the star count of this image.|integer|


<a name="status"></a>
### Status
The status of the container. For example, "running" or "exited".

*Type* : enum (created, running, stopped, paused, restarting, removing, exited, dead)


<a name="systeminfo"></a>
### SystemInfo

|Name|Schema|
|---|---|
|**Architecture**  <br>*optional*|string|
|**Containers**  <br>*optional*|integer|
|**ContainersPaused**  <br>*optional*|integer|
|**ContainersRunning**  <br>*optional*|integer|
|**ContainersStopped**  <br>*optional*|integer|


<a name="systemversion"></a>
### SystemVersion

|Name|Description|Schema|
|---|---|---|
|**ApiVersion**  <br>*optional*|Api Version held by daemon  <br>**Example** : `""`|string|
|**Arch**  <br>*optional*|Arch type of underlying hardware  <br>**Example** : `"amd64"`|string|
|**BuildTime**  <br>*optional*|The time when this binary of daemon is built  <br>**Example** : `"2017-08-29T17:41:57.729792388+00:00"`|string|
|**GitCommit**  <br>*optional*|Commit ID held by the latest commit operation  <br>**Example** : `""`|string|
|**GoVersion**  <br>*optional*|version of Go runtime  <br>**Example** : `"1.8.3"`|string|
|**KernelVersion**  <br>*optional*|Operating system kernel version  <br>**Example** : `"3.13.0-106-generic"`|string|
|**Os**  <br>*optional*|Operating system type of underlying system  <br>**Example** : `"linux"`|string|
|**Version**  <br>*optional*|version of Pouch Daemon  <br>**Example** : `"0.1.2"`|string|


<a name="volumeconfig"></a>
### VolumeConfig
config used to create a volume


|Name|Description|Schema|
|---|---|---|
|**Driver**  <br>*optional*|Name of the volume driver to use.  <br>**Default** : `"local"`|string|
|**DriverOpts**  <br>*optional*|A mapping of driver options and values. These options are passed directly to the driver and are driver specific.|< string, string > map|
|**Labels**  <br>*optional*|User-defined key/value metadata.|< string, string > map|
|**Name**  <br>*optional*|The new volume's name. If not specified, Docker generates a name.|string|


<a name="volumecreaterequest"></a>
### VolumeCreateRequest
VolumeCreateRequest contains the response for the remote API: POST /volumes/create


|Name|Description|Schema|
|---|---|---|
|**Driver**  <br>*optional*|Driver is the Driver name used to create the volume.|string|
|**DriverOpts**  <br>*optional*|DriverOpts holds the driver specific options to use for when creating the volume.|< string, string > map|
|**Labels**  <br>*optional*|Labels is metadata specific to the volume.|< string, string > map|
|**Name**  <br>*optional*|Name is the name of the volume.|string|


<a name="volumeinfo"></a>
### VolumeInfo
Volume represents the configuration of a volume for the container.


|Name|Description|Schema|
|---|---|---|
|**CreatedAt**  <br>*optional*|Date/Time the volume was created.|string (dateTime)|
|**Driver**  <br>*optional*|Driver is the Driver name used to create the volume.|string|
|**Labels**  <br>*optional*|Labels is metadata specific to the volume.|< string, string > map|
|**Mountpoint**  <br>*optional*|Mountpoint is the location on disk of the volume.|string|
|**Name**  <br>*optional*|Name is the name of the volume.|string|
|**Scope**  <br>*optional*|Scope describes the level at which the volume exists<br>(e.g. `global` for cluster-wide or `local` for machine level)|string|
|**Status**  <br>*optional*|Status provides low-level status information about the volume.|< string, object > map|


<a name="volumelistresp"></a>
### VolumeListResp

|Name|Description|Schema|
|---|---|---|
|**Volumes**  <br>*required*|List of volumes|< [VolumeInfo](#volumeinfo) > array|
|**Warnings**  <br>*required*|Warnings that occurred when fetching the list of volumes|< string > array|





