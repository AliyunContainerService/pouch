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
|**500**|An unexpected server error occured.|[Error](#error)|


#### Example HTTP response

##### Response 200
```
json :
"OK"
```


<a name="auth-post"></a>
### Check auth configuration
```
POST /auth
```


#### Description
Validate credentials for a registry and, if available, get an identity token for accessing the registry without password.


#### Parameters

|Type|Name|Description|Schema|
|---|---|---|---|
|**Body**|**authConfig**  <br>*optional*|Authentication to check|[AuthConfig](#authconfig)|


#### Responses

|HTTP Code|Description|Schema|
|---|---|---|
|**200**|An identity token was generated successfully.|[AuthResponse](#authresponse)|
|**401**|Login unauthorized|[1ErrorResponse](#1errorresponse)|
|**500**|Server error|[0ErrorResponse](#0errorresponse)|


#### Consumes

* `application/json`


#### Produces

* `application/json`


<a name="containers-create-post"></a>
### Create a container
```
POST /containers/create
```


#### Parameters

|Type|Name|Description|Schema|
|---|---|---|---|
|**Query**|**name**  <br>*optional*|Assign the specified name to the container. Must match `/?[a-zA-Z0-9_-]+`.|string|
|**Body**|**body**  <br>*required*|Container to create|[ContainerCreateConfig](#containercreateconfig)|


#### Responses

|HTTP Code|Description|Schema|
|---|---|---|
|**201**|Container created successfully|[ContainerCreateResp](#containercreateresp)|
|**400**|bad parameter|[Error](#error)|
|**404**|no such image|[Error](#error)|
|**409**|conflict|[Error](#error)|
|**500**|An unexpected server error occured.|[Error](#error)|


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
  "message" : "image: xxx:latest: not found"
}
```


<a name="containerlist"></a>
### List containers
```
GET /containers/json
```


#### Parameters

|Type|Name|Description|Schema|Default|
|---|---|---|---|---|
|**Query**|**all**  <br>*optional*|Return all containers. By default, only running containers are shown|boolean|`"false"`|


#### Responses

|HTTP Code|Description|Schema|
|---|---|---|
|**200**|Summary containers that matches the query|< [Container](#container) > array|
|**500**|An unexpected server error occured.|[Error](#error)|


#### Produces

* `application/json`


<a name="containerremove"></a>
### Remove one container
```
DELETE /containers/{id}
```


#### Parameters

|Type|Name|Description|Schema|
|---|---|---|---|
|**Path**|**id**  <br>*required*|ID or name of the container|string|
|**Query**|**force**  <br>*optional*|If the container is running, force query is used to kill it and remove it forcefully.|boolean|


#### Responses

|HTTP Code|Description|Schema|
|---|---|---|
|**204**|no error|No Content|
|**404**|An unexpected 404 error occured.|[Error](#error)|
|**500**|An unexpected server error occured.|[Error](#error)|


#### Tags

* Container


<a name="containerattach"></a>
### Attach to a container
```
POST /containers/{id}/attach
```


#### Description
Attach to a container to read its output or send it input. You can attach to the same container multiple times and you can reattach to containers that have been detached.

Either the `stream` or `logs` parameter must be `true` for this endpoint to do anything.

### Hijacking

This endpoint hijacks the HTTP connection to transport `stdin`, `stdout`, and `stderr` on the same socket.

This is the response from the daemon for an attach request:

```
HTTP/1.1 200 OK
Content-Type: application/vnd.raw-stream

[STREAM]
```

After the headers and two new lines, the TCP connection can now be used for raw, bidirectional communication between the client and server.

To hint potential proxies about connection hijacking, the Pouch client can also optionally send connection upgrade headers.

For example, the client sends this request to upgrade the connection:

```
POST /containers/16253994b7c4/attach?stream=1&stdout=1 HTTP/1.1
Upgrade: tcp
Connection: Upgrade
```

The Pouch daemon will respond with a `101 UPGRADED` response, and will similarly follow with the raw stream:

```
HTTP/1.1 101 UPGRADED
Content-Type: application/vnd.raw-stream
Connection: Upgrade
Upgrade: tcp

[STREAM]
```

### Stream format

When the TTY setting is disabled in [`POST /containers/create`](#operation/ContainerCreate), the stream over the hijacked connected is multiplexed to separate out `stdout` and `stderr`. The stream consists of a series of frames, each containing a header and a payload.

The header contains the information which the stream writes (`stdout` or `stderr`). It also contains the size of the associated frame encoded in the last four bytes (`uint32`).

It is encoded on the first eight bytes like this:

```go
header := [8]byte{STREAM_TYPE, 0, 0, 0, SIZE1, SIZE2, SIZE3, SIZE4}
```

`STREAM_TYPE` can be:

- 0: `stdin` (is written on `stdout`)
- 1: `stdout`
- 2: `stderr`

`SIZE1, SIZE2, SIZE3, SIZE4` are the four bytes of the `uint32` size encoded as big endian.

Following the header is the payload, which is the specified number of bytes of `STREAM_TYPE`.

The simplest way to implement this protocol is the following:

1. Read 8 bytes.
2. Choose `stdout` or `stderr` depending on the first byte.
3. Extract the frame size from the last four bytes.
4. Read the extracted size and output it on the correct output.
5. Goto 1.

### Stream format when using a TTY

When the TTY setting is enabled in [`POST /containers/create`](#operation/ContainerCreate), the stream is not multiplexed. The data exchanged over the hijacked connection is simply the raw data from the process PTY and client's `stdin`.


#### Parameters

|Type|Name|Description|Schema|Default|
|---|---|---|---|---|
|**Path**|**id**  <br>*required*|ID or name of the container|string||
|**Query**|**detachKeys**  <br>*optional*|Override the key sequence for detaching a container.Format is a single character `[a-Z]` or `ctrl-<value>` where `<value>` is one of: `a-z`, `@`, `^`, `[`, `,` or `_`.|string||
|**Query**|**logs**  <br>*optional*|Replay previous logs from the container.<br><br>This is useful for attaching to a container that has started and you want to output everything since the container started.<br><br>If `stream` is also enabled, once all the previous output has been returned, it will seamlessly transition into streaming current output.|boolean|`"false"`|
|**Query**|**stderr**  <br>*optional*|Attach to `stderr`|boolean|`"false"`|
|**Query**|**stdin**  <br>*optional*|Attach to `stdin`|boolean|`"false"`|
|**Query**|**stdout**  <br>*optional*|Attach to `stdout`|boolean|`"false"`|
|**Query**|**stream**  <br>*optional*|Stream attached streams from the time the request was made onwards|boolean|`"false"`|


#### Responses

|HTTP Code|Description|Schema|
|---|---|---|
|**101**|no error, hints proxy about hijacking|No Content|
|**200**|no error, no upgrade header found|No Content|
|**400**|bad parameter|[Error](#error)|
|**404**|no such container|[Error](#error)|
|**500**|server error|[Error](#error)|


#### Produces

* `application/vnd.raw-stream`


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


<a name="containerexec"></a>
### Create an exec instance
```
POST /containers/{id}/exec
```


#### Description
Run a command inside a running container.


#### Parameters

|Type|Name|Description|Schema|
|---|---|---|---|
|**Path**|**id**  <br>*required*|ID or name of the container|string|
|**Body**|**body**  <br>*required*||[ExecCreateConfig](#execcreateconfig)|


#### Responses

|HTTP Code|Description|Schema|
|---|---|---|
|**201**|no error|[ExecCreateResp](#execcreateresp)|
|**404**|An unexpected 404 error occured.|[Error](#error)|
|**409**|container is paused|[Error](#error)|
|**500**|An unexpected server error occured.|[Error](#error)|


#### Consumes

* `application/json`


#### Produces

* `application/json`


#### Tags

* Exec


<a name="containerinspect"></a>
### Inspect a container
```
GET /containers/{id}/json
```


#### Description
Return low-level information about a container.


#### Parameters

|Type|Name|Description|Schema|
|---|---|---|---|
|**Path**|**id**  <br>*required*|ID or name of the container|string|
|**Query**|**size**  <br>*optional*|Return the size of container as fields `SizeRw` and `SizeRootFs`|boolean|


#### Responses

|HTTP Code|Description|Schema|
|---|---|---|
|**200**|no error|[ContainerJSON](#containerjson)|
|**404**|An unexpected 404 error occured.|[Error](#error)|
|**500**|An unexpected server error occured.|[Error](#error)|


#### Produces

* `application/json`


#### Tags

* Container


<a name="containerlogs"></a>
### Get container logs
```
GET /containers/{id}/logs
```


#### Description
Get `stdout` and `stderr` logs from a container.

Note: This endpoint works only for containers with the `json-file` or `journald` logging driver.


#### Parameters

|Type|Name|Description|Schema|Default|
|---|---|---|---|---|
|**Path**|**id**  <br>*required*|ID or name of the container|string||
|**Query**|**follow**  <br>*optional*|Return the logs as a stream.|boolean|`"false"`|
|**Query**|**since**  <br>*optional*|Only return logs since this time, as a UNIX timestamp|integer|`0`|
|**Query**|**stderr**  <br>*optional*|Return logs from `stderr`|boolean|`"false"`|
|**Query**|**stdout**  <br>*optional*|Return logs from `stdout`|boolean|`"false"`|
|**Query**|**tail**  <br>*optional*|Only return this number of log lines from the end of the logs. Specify as an integer or `all` to output all log lines.|string|`"all"`|
|**Query**|**timestamps**  <br>*optional*|Add timestamps to every log line|boolean|`"false"`|
|**Query**|**until**  <br>*optional*|Only return logs before this time, as a UNIX timestamp|integer|`0`|


#### Responses

|HTTP Code|Description|Schema|
|---|---|---|
|**101**|logs returned as a stream|string (binary)|
|**200**|logs returned as a string in response body|string|
|**404**|An unexpected 404 error occured.|[Error](#error)|
|**500**|An unexpected server error occured.|[Error](#error)|


#### Tags

* Container


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
|**404**|An unexpected 404 error occured.|[Error](#error)|
|**500**|An unexpected server error occured.|[Error](#error)|


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
|**404**|An unexpected 404 error occured.|[Error](#error)|
|**409**|name already in use|[Error](#error)|
|**500**|An unexpected server error occured.|[Error](#error)|


#### Tags

* Container


<a name="containerresize"></a>
### changes the size of the tty for a container
```
POST /containers/{id}/resize
```


#### Parameters

|Type|Name|Description|Schema|
|---|---|---|---|
|**Path**|**id**  <br>*required*|ID or name of the container|string|
|**Query**|**height**  <br>*optional*|height of the tty|string|
|**Query**|**width**  <br>*optional*|width of the tty|string|


#### Responses

|HTTP Code|Description|Schema|
|---|---|---|
|**200**|no error|No Content|
|**400**|bad parameter|[Error](#error)|


#### Tags

* Container


<a name="containerrestart"></a>
### Restart a container
```
POST /containers/{id}/restart
```


#### Parameters

|Type|Name|Description|Schema|
|---|---|---|---|
|**Path**|**id**  <br>*required*|ID or name of the container|string|
|**Query**|**name**  <br>*required*|New name for the container|string|
|**Query**|**t**  <br>*optional*|Number of seconds to wait before killing the container|integer|


#### Responses

|HTTP Code|Description|Schema|
|---|---|---|
|**204**|no error|No Content|
|**404**|An unexpected 404 error occured.|[Error](#error)|
|**500**|An unexpected server error occured.|[Error](#error)|


#### Tags

* Container


<a name="containerstart"></a>
### Start a container
```
POST /containers/{id}/start
```


#### Parameters

|Type|Name|Description|Schema|
|---|---|---|---|
|**Path**|**id**  <br>*required*|ID or name of the container|string|
|**Query**|**detachKeys**  <br>*optional*|Override the key sequence for detaching a container. Format is a single character `[a-Z]` or `ctrl-<value>` where `<value>` is one of: `a-z`, `@`, `^`, `[`, `,` or `_`.|string|


#### Responses

|HTTP Code|Description|Schema|
|---|---|---|
|**204**|no error|No Content|
|**404**|An unexpected 404 error occured.|[Error](#error)|
|**500**|An unexpected server error occured.|[Error](#error)|


#### Tags

* Container


<a name="containerstop"></a>
### Stop a container
```
POST /containers/{id}/stop
```


#### Parameters

|Type|Name|Description|Schema|
|---|---|---|---|
|**Path**|**id**  <br>*required*|ID or name of the container|string|
|**Query**|**t**  <br>*optional*|Number of seconds to wait before killing the container|integer|


#### Responses

|HTTP Code|Description|Schema|
|---|---|---|
|**204**|no error|No Content|
|**404**|An unexpected 404 error occured.|[Error](#error)|
|**500**|An unexpected server error occured.|[Error](#error)|


#### Tags

* Container


<a name="containertop"></a>
### Display the running processes of a container
```
POST /containers/{id}/top
```


#### Parameters

|Type|Name|Description|Schema|
|---|---|---|---|
|**Path**|**id**  <br>*required*|ID or name of the container|string|
|**Query**|**ps_args**  <br>*optional*|top arguments|string|


#### Responses

|HTTP Code|Description|Schema|
|---|---|---|
|**200**|no error|[ContainerProcessList](#containerprocesslist)|
|**404**|An unexpected 404 error occured.|[Error](#error)|
|**500**|An unexpected server error occured.|[Error](#error)|


#### Tags

* Container


<a name="containerunpause"></a>
### Unpause a container
```
POST /containers/{id}/unpause
```


#### Parameters

|Type|Name|Description|Schema|
|---|---|---|---|
|**Path**|**id**  <br>*required*|ID or name of the container|string|


#### Responses

|HTTP Code|Description|Schema|
|---|---|---|
|**204**|no error|No Content|
|**404**|An unexpected 404 error occured.|[Error](#error)|
|**500**|An unexpected server error occured.|[Error](#error)|


#### Tags

* Container


<a name="containerupdate"></a>
### Update the configurations of a container
```
POST /containers/{id}/update
```


#### Parameters

|Type|Name|Description|Schema|
|---|---|---|---|
|**Path**|**id**  <br>*required*|ID or name of the container|string|
|**Body**|**updateConfig**  <br>*optional*||[UpdateConfig](#updateconfig)|


#### Responses

|HTTP Code|Description|Schema|
|---|---|---|
|**200**|no error|No Content|
|**400**|bad parameter|[Error](#error)|
|**404**|An unexpected 404 error occured.|[Error](#error)|
|**500**|An unexpected server error occured.|[Error](#error)|


#### Tags

* Container


<a name="containerupgrade"></a>
### Upgrade a container with new image and args
```
POST /containers/{id}/upgrade
```


#### Parameters

|Type|Name|Description|Schema|
|---|---|---|---|
|**Path**|**id**  <br>*required*|ID or name of the container|string|
|**Body**|**upgradeConfig**  <br>*optional*||[ContainerUpgradeConfig](#containerupgradeconfig)|


#### Responses

|HTTP Code|Description|Schema|
|---|---|---|
|**200**|no error|No Content|
|**400**|bad parameter|[Error](#error)|
|**404**|An unexpected 404 error occured.|[Error](#error)|
|**500**|An unexpected server error occured.|[Error](#error)|


#### Tags

* Container


<a name="daemon-update-post"></a>
### Update daemon's labels and image proxy
```
POST /daemon/update
```


#### Parameters

|Type|Name|Description|Schema|
|---|---|---|---|
|**Body**|**DaemonUpdateConfig**  <br>*optional*|Config used to update daemon, only labels and image proxy are allowed.|[DaemonUpdateConfig](#daemonupdateconfig)|


#### Responses

|HTTP Code|Description|Schema|
|---|---|---|
|**200**|no error|No Content|
|**400**|bad parameter|[Error](#error)|
|**500**|An unexpected server error occured.|[Error](#error)|


#### Consumes

* `application/json`


#### Produces

* `application/json`


<a name="execinspect"></a>
### Inspect an exec instance
```
GET /exec/{id}/json
```


#### Description
Return low-level information about an exec instance.


#### Parameters

|Type|Name|Description|Schema|
|---|---|---|---|
|**Path**|**id**  <br>*required*|Exec instance ID|string|


#### Responses

|HTTP Code|Description|Schema|
|---|---|---|
|**200**|No error|[ContainerExecInspect](#containerexecinspect)|
|**404**|No such exec instance|[4ErrorResponse](#4errorresponse)|
|**500**|Server error|[0ErrorResponse](#0errorresponse)|


#### Produces

* `application/json`


#### Tags

* Exec


<a name="execstart"></a>
### Start an exec instance
```
POST /exec/{id}/start
```


#### Description
Starts a previously set up exec instance. If detach is true, this endpoint returns immediately after starting the command. Otherwise, it sets up an interactive session with the command.


#### Parameters

|Type|Name|Description|Schema|
|---|---|---|---|
|**Path**|**id**  <br>*required*|Exec instance ID|string|
|**Body**|**execStartConfig**  <br>*optional*||[ExecStartConfig](#execstartconfig)|


#### Responses

|HTTP Code|Description|Schema|
|---|---|---|
|**200**|No error|No Content|
|**404**|No such exec instance|[Error](#error)|
|**409**|Container is stopped or paused|[Error](#error)|


#### Consumes

* `application/json`


#### Produces

* `application/vnd.raw-stream`


#### Tags

* Exec


#### Example HTTP request

##### Request body
```
json :
{
  "Detach" : false,
  "Tty" : false
}
```


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
|**500**|An unexpected server error occured.|[Error](#error)|


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

|Type|Name|Description|Schema|
|---|---|---|---|
|**Query**|**all**  <br>*optional*|Show all images. Only images from a final layer (no children) are shown by default.|boolean|
|**Query**|**digests**  <br>*optional*|Show digest information as a `RepoDigests` field on each image.|boolean|
|**Query**|**filters**  <br>*optional*|A JSON encoded value of the filters (a `map[string][]string`) to process on the images list. Available filters:<br><br>- `before`=(`<image-name>[:<tag>]`,  `<image id>` or `<image@digest>`)<br>- `dangling=true`<br>- `label=key` or `label="key=value"` of an image label<br>- `reference`=(`<image-name>[:<tag>]`)<br>- `since`=(`<image-name>[:<tag>]`,  `<image id>` or `<image@digest>`)|string|


#### Responses

|HTTP Code|Description|Schema|
|---|---|---|
|**200**|Summary image data for the images matching the query|< [ImageInfo](#imageinfo) > array|
|**500**|An unexpected server error occured.|[Error](#error)|


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
|**500**|An unexpected server error occured.|[Error](#error)|


#### Produces

* `application/json`


<a name="images-imageid-delete"></a>
### Remove an image
```
DELETE /images/{imageid}
```


#### Description
Remove an image by reference.


#### Parameters

|Type|Name|Description|Schema|Default|
|---|---|---|---|---|
|**Path**|**imageid**  <br>*required*|Image name or id|string||
|**Query**|**force**  <br>*optional*|Remove the image even if it is being used|boolean|`"false"`|


#### Responses

|HTTP Code|Description|Schema|
|---|---|---|
|**204**|No error|No Content|
|**404**|no such image|[Error](#error)|
|**500**|An unexpected server error occured.|[Error](#error)|


#### Example HTTP response

##### Response 404
```
json :
{
  "message" : "No such image: c2ada9df5af8"
}
```


<a name="imageinspect"></a>
### Inspect a image
```
GET /images/{imageid}/json
```


#### Description
Return the information about image


#### Parameters

|Type|Name|Description|Schema|
|---|---|---|---|
|**Path**|**imageid**  <br>*required*|Image name or id|string|


#### Responses

|HTTP Code|Description|Schema|
|---|---|---|
|**200**|no error|[ImageInfo](#imageinfo)|
|**404**|no such image|[Error](#error)|
|**500**|An unexpected server error occured.|[Error](#error)|


#### Produces

* `application/json`


#### Example HTTP response

##### Response 200
```
json :
{
  "CreatedAt" : "2017-12-19 15:32:09",
  "Digest" : "sha256:e216a057b1cb1efc11f8a268f37ef62083e70b1b38323ba252e25ac88904a7e8",
  "ID" : "e216a057b1cb",
  "Name" : "ubuntu:12.04",
  "Size" : 103579269,
  "Tag" : "12.04"
}
```


##### Response 404
```
json :
{
  "message" : "No such image: e216a057b1cb"
}
```


<a name="info-get"></a>
### Get System information
```
GET /info
```


#### Responses

|HTTP Code|Description|Schema|
|---|---|---|
|**200**|no error|[SystemInfo](#systeminfo)|
|**500**|An unexpected server error occured.|[Error](#error)|


<a name="networklist"></a>
### List networks
```
GET /networks
```


#### Responses

|HTTP Code|Description|Schema|
|---|---|---|
|**200**|Summary networks that matches the query|[NetworkListResp](#networklistresp)|
|**500**|An unexpected server error occured.|[Error](#error)|


#### Produces

* `application/json`


#### Tags

* Network


<a name="networkcreate"></a>
### Create a network
```
POST /networks/create
```


#### Parameters

|Type|Name|Description|Schema|
|---|---|---|---|
|**Body**|**NetworkCreateConfig**  <br>*required*|Network configuration|[NetworkCreateConfig](#networkcreateconfig)|


#### Responses

|HTTP Code|Description|Schema|
|---|---|---|
|**201**|The network was created successfully|[NetworkCreateResp](#networkcreateresp)|
|**400**|bad parameter|[Error](#error)|
|**409**|name already in use|[Error](#error)|
|**500**|An unexpected server error occured.|[Error](#error)|


#### Consumes

* `application/json`


#### Produces

* `application/json`


#### Tags

* Network


<a name="networkinspect"></a>
### Inspect a network
```
GET /networks/{id}
```


#### Parameters

|Type|Name|Description|Schema|
|---|---|---|---|
|**Path**|**id**  <br>*required*|ID or name of the container|string|


#### Responses

|HTTP Code|Description|Schema|
|---|---|---|
|**200**|No error|[NetworkInspectResp](#networkinspectresp)|
|**404**|An unexpected 404 error occured.|[Error](#error)|
|**500**|An unexpected server error occured.|[Error](#error)|


#### Produces

* `application/json`


#### Tags

* Network


<a name="networkdelete"></a>
### Remove a network
```
DELETE /networks/{id}
```


#### Parameters

|Type|Name|Description|Schema|
|---|---|---|---|
|**Path**|**id**  <br>*required*|ID or name of the container|string|


#### Responses

|HTTP Code|Description|Schema|
|---|---|---|
|**204**|No error|No Content|
|**403**|operation not supported for pre-defined networks|[Error](#error)|
|**404**|An unexpected 404 error occured.|[Error](#error)|
|**500**|An unexpected server error occured.|[Error](#error)|


#### Tags

* Network


<a name="version-get"></a>
### Get Pouchd version
```
GET /version
```


#### Responses

|HTTP Code|Description|Schema|
|---|---|---|
|**200**|no error|[SystemVersion](#systemversion)|
|**500**|An unexpected server error occured.|[Error](#error)|


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
|**500**|An unexpected server error occured.|[Error](#error)|


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
    "Mountpoint" : "/var/lib/pouch/volumes/tardis",
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
|**Body**|**body**  <br>*required*|Volume configuration|[VolumeCreateConfig](#volumecreateconfig)|


#### Responses

|HTTP Code|Description|Schema|
|---|---|---|
|**201**|The volume was created successfully|[VolumeInfo](#volumeinfo)|
|**500**|An unexpected server error occured.|[Error](#error)|


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


<a name="volumeinspect"></a>
### Inspect a volume
```
GET /volumes/{id}
```


#### Parameters

|Type|Name|Description|Schema|
|---|---|---|---|
|**Path**|**id**  <br>*required*|ID or name of the container|string|


#### Responses

|HTTP Code|Description|Schema|
|---|---|---|
|**200**|No error|[VolumeInfo](#volumeinfo)|
|**404**|An unexpected 404 error occured.|[Error](#error)|
|**500**|An unexpected server error occured.|[Error](#error)|


#### Produces

* `application/json`


#### Tags

* Volume


<a name="volumedelete"></a>
### Delete a volume
```
DELETE /volumes/{id}
```


#### Parameters

|Type|Name|Description|Schema|
|---|---|---|---|
|**Path**|**id**  <br>*required*|ID or name of the container|string|


#### Responses

|HTTP Code|Description|Schema|
|---|---|---|
|**204**|No error|No Content|
|**404**|An unexpected 404 error occured.|[Error](#error)|
|**500**|An unexpected server error occured.|[Error](#error)|


#### Tags

* Volume




<a name="definitions"></a>
## Definitions

<a name="authconfig"></a>
### AuthConfig

|Name|Description|Schema|
|---|---|---|
|**Auth**  <br>*optional*||string|
|**IdentityToken**  <br>*optional*|IdentityToken is used to authenticate the user and get an access token for the registry|string|
|**Password**  <br>*optional*||string|
|**RegistryToken**  <br>*optional*|RegistryToken is a bearer token to be sent to a registry|string|
|**ServerAddress**  <br>*optional*||string|
|**Username**  <br>*optional*||string|


<a name="authresponse"></a>
### AuthResponse
The response returned by login to a registry


|Name|Description|Schema|
|---|---|---|
|**IdentityToken**  <br>*optional*|An opaque token used to authenticate a user after a successful login|string|
|**Status**  <br>*required*|The status of the authentication|string|


<a name="commit"></a>
### Commit
Commit holds the Git-commit (SHA1) that a binary was built from, as
reported in the version-string of external tools, such as `containerd`,
or `runC`.


|Name|Description|Schema|
|---|---|---|
|**Expected**  <br>*optional*|Commit ID of external tool expected by pouchd as set at build time.  <br>**Example** : `"2d41c047c83e09a6d61d464906feb2a2f3c52aa4"`|string|
|**ID**  <br>*optional*|Actual commit ID of external tool.  <br>**Example** : `"cfb82a876ecc11b5ca0977d1733adbe58599088a"`|string|


<a name="container"></a>
### Container
an array of Container contains response of Engine API:
GET "/containers/json"


|Name|Description|Schema|
|---|---|---|
|**Command**  <br>*optional*||string|
|**Created**  <br>*optional*|Created time of container in daemon.|integer (int64)|
|**HostConfig**  <br>*optional*|In Moby's API, HostConfig field in Container struct has following type <br>struct { NetworkMode string `json:",omitempty"` }<br>In Pouch, we need to pick runtime field in HostConfig from daemon side to judge runtime type,<br>So Pouch changes this type to be the complete HostConfig.<br>Incompatibility exists, ATTENTION.|[HostConfig](#hostconfig)|
|**Id**  <br>*optional*|Container ID|string|
|**Image**  <br>*optional*||string|
|**ImageID**  <br>*optional*||string|
|**Labels**  <br>*optional*||< string, string > map|
|**Mounts**  <br>*optional*|Set of mount point in a container.|< [MountPoint](#mountpoint) > array|
|**Names**  <br>*optional*|**Example** : `[ "container_1", "container_2" ]`|< string > array|
|**NetworkSettings**  <br>*optional*||object|
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
|**AttachStdin**  <br>*optional*|Whether to attach to `stdin`.|boolean|
|**AttachStdout**  <br>*optional*|Whether to attach to `stdout`.  <br>**Default** : `true`|boolean|
|**Cmd**  <br>*optional*|Command to run specified an array of strings.|< string > array|
|**DiskQuota**  <br>*optional*|Set disk quota for container|< string, string > map|
|**Domainname**  <br>*optional*|The domain name to use for the container.|string|
|**Entrypoint**  <br>*optional*|The entry point for the container as a string or an array of strings.<br>If the array consists of exactly one empty string (`[""]`) then the entry point is reset to system default.|< string > array|
|**Env**  <br>*optional*|A list of environment variables to set inside the container in the form `["VAR=value", ...]`. A variable without `=` is removed from the environment, rather than to have an empty value.|< string > array|
|**ExposedPorts**  <br>*optional*|An object mapping ports to an empty object in the form:`{<port>/<tcp\|udp>: {}}`|< string, object > map|
|**Hostname**  <br>*optional*|The hostname to use for the container, as a valid RFC 1123 hostname.  <br>**Minimum length** : `1`|string (hostname)|
|**Image**  <br>*required*|The name of the image to use when creating the container|string|
|**InitScript**  <br>*optional*|Initial script executed in container. The script will be executed before entrypoint or command|string|
|**Labels**  <br>*optional*|User-defined key/value metadata.|< string, string > map|
|**MacAddress**  <br>*optional*|MAC address of the container.|string|
|**NetworkDisabled**  <br>*optional*|Disable networking for the container.|boolean|
|**OnBuild**  <br>*optional*|`ONBUILD` metadata that were defined.|< string > array|
|**OpenStdin**  <br>*optional*|Open `stdin`|boolean|
|**Rich**  <br>*optional*|Whether to start container in rich container mode. (default false)|boolean|
|**RichMode**  <br>*optional*|Choose one rich container mode.(default dumb-init)|enum (dumb-init, sbin-init, systemd)|
|**Shell**  <br>*optional*|Shell for when `RUN`, `CMD`, and `ENTRYPOINT` uses a shell.|< string > array|
|**SpecAnnotation**  <br>*optional*|annotations send to runtime spec.|< string, string > map|
|**StdinOnce**  <br>*optional*|Close `stdin` after one attached client disconnects|boolean|
|**StopSignal**  <br>*optional*|Signal to stop a container as a string or unsigned integer.  <br>**Default** : `"SIGTERM"`|string|
|**StopTimeout**  <br>*optional*|Timeout to stop a container in seconds.|integer|
|**Tty**  <br>*optional*|Attach standard streams to a TTY, including `stdin` if it is not closed.|boolean|
|**User**  <br>*optional*|The user that commands are run as inside the container.|string|
|**Volumes**  <br>*optional*|An object mapping mount point paths inside the container to empty objects.|< string, object > map|
|**WorkingDir**  <br>*optional*|The working directory for commands to run in.|string|


<a name="containercreateconfig"></a>
### ContainerCreateConfig
ContainerCreateConfig is used for API "POST /containers/create".
It wraps all kinds of config used in container creation.
It can be used to encode client params in client and unmarshal request body in daemon side.

*Polymorphism* : Composition


|Name|Description|Schema|
|---|---|---|
|**ArgsEscaped**  <br>*optional*|Command is already escaped (Windows only)|boolean|
|**AttachStderr**  <br>*optional*|Whether to attach to `stderr`.  <br>**Default** : `true`|boolean|
|**AttachStdin**  <br>*optional*|Whether to attach to `stdin`.|boolean|
|**AttachStdout**  <br>*optional*|Whether to attach to `stdout`.  <br>**Default** : `true`|boolean|
|**Cmd**  <br>*optional*|Command to run specified an array of strings.|< string > array|
|**DiskQuota**  <br>*optional*|Set disk quota for container|< string, string > map|
|**Domainname**  <br>*optional*|The domain name to use for the container.|string|
|**Entrypoint**  <br>*optional*|The entry point for the container as a string or an array of strings.<br>If the array consists of exactly one empty string (`[""]`) then the entry point is reset to system default.|< string > array|
|**Env**  <br>*optional*|A list of environment variables to set inside the container in the form `["VAR=value", ...]`. A variable without `=` is removed from the environment, rather than to have an empty value.|< string > array|
|**ExposedPorts**  <br>*optional*|An object mapping ports to an empty object in the form:`{<port>/<tcp\|udp>: {}}`|< string, object > map|
|**HostConfig**  <br>*optional*||[HostConfig](#hostconfig)|
|**Hostname**  <br>*optional*|The hostname to use for the container, as a valid RFC 1123 hostname.  <br>**Minimum length** : `1`|string (hostname)|
|**Image**  <br>*required*|The name of the image to use when creating the container|string|
|**InitScript**  <br>*optional*|Initial script executed in container. The script will be executed before entrypoint or command|string|
|**Labels**  <br>*optional*|User-defined key/value metadata.|< string, string > map|
|**MacAddress**  <br>*optional*|MAC address of the container.|string|
|**NetworkDisabled**  <br>*optional*|Disable networking for the container.|boolean|
|**NetworkingConfig**  <br>*optional*||[NetworkingConfig](#networkingconfig)|
|**OnBuild**  <br>*optional*|`ONBUILD` metadata that were defined.|< string > array|
|**OpenStdin**  <br>*optional*|Open `stdin`|boolean|
|**Rich**  <br>*optional*|Whether to start container in rich container mode. (default false)|boolean|
|**RichMode**  <br>*optional*|Choose one rich container mode.(default dumb-init)|enum (dumb-init, sbin-init, systemd)|
|**Shell**  <br>*optional*|Shell for when `RUN`, `CMD`, and `ENTRYPOINT` uses a shell.|< string > array|
|**SpecAnnotation**  <br>*optional*|annotations send to runtime spec.|< string, string > map|
|**StdinOnce**  <br>*optional*|Close `stdin` after one attached client disconnects|boolean|
|**StopSignal**  <br>*optional*|Signal to stop a container as a string or unsigned integer.  <br>**Default** : `"SIGTERM"`|string|
|**StopTimeout**  <br>*optional*|Timeout to stop a container in seconds.|integer|
|**Tty**  <br>*optional*|Attach standard streams to a TTY, including `stdin` if it is not closed.|boolean|
|**User**  <br>*optional*|The user that commands are run as inside the container.|string|
|**Volumes**  <br>*optional*|An object mapping mount point paths inside the container to empty objects.|< string, object > map|
|**WorkingDir**  <br>*optional*|The working directory for commands to run in.|string|


<a name="containercreateresp"></a>
### ContainerCreateResp
response returned by daemon when container create successfully


|Name|Description|Schema|
|---|---|---|
|**Id**  <br>*required*|The ID of the created container|string|
|**Name**  <br>*optional*|The name of the created container|string|
|**Warnings**  <br>*required*|Warnings encountered when creating the container|< string > array|


<a name="containerexecinspect"></a>
### ContainerExecInspect
holds information about a running process started.


|Name|Description|Schema|
|---|---|---|
|**CanRemove**  <br>*optional*||boolean|
|**ContainerID**  <br>*optional*|The ID of this container|string|
|**DetachKeys**  <br>*optional*||string|
|**ExitCode**  <br>*optional*|The last exit code of this container|integer|
|**ID**  <br>*optional*|The ID of this exec|string|
|**OpenStderr**  <br>*optional*||boolean|
|**OpenStdin**  <br>*optional*||boolean|
|**OpenStdout**  <br>*optional*||boolean|
|**ProcessConfig**  <br>*optional*||[ProcessConfig](#processconfig)|
|**Running**  <br>*optional*||boolean|


<a name="containerjson"></a>
### ContainerJSON
ContainerJSON contains response of Engine API:
GET "/containers/{id}/json"


|Name|Description|Schema|
|---|---|---|
|**AppArmorProfile**  <br>*optional*||string|
|**Args**  <br>*optional*|The arguments to the command being run|< string > array|
|**Config**  <br>*optional*||[ContainerConfig](#containerconfig)|
|**Created**  <br>*optional*|The time the container was created|string|
|**Driver**  <br>*optional*||string|
|**ExecIDs**  <br>*optional*||string|
|**GraphDriver**  <br>*optional*||[GraphDriverData](#graphdriverdata)|
|**HostConfig**  <br>*optional*||[HostConfig](#hostconfig)|
|**HostnamePath**  <br>*optional*||string|
|**HostsPath**  <br>*optional*||string|
|**Id**  <br>*optional*|The ID of the container|string|
|**Image**  <br>*optional*|The container's image|string|
|**LogPath**  <br>*optional*||string|
|**MountLabel**  <br>*optional*||string|
|**Mounts**  <br>*optional*|Set of mount point in a container.|< [MountPoint](#mountpoint) > array|
|**Name**  <br>*optional*||string|
|**NetworkSettings**  <br>*optional*|NetworkSettings exposes the network settings in the API.|[NetworkSettings](#networksettings)|
|**Path**  <br>*optional*|The path to the command being run|string|
|**ProcessLabel**  <br>*optional*||string|
|**ResolvConfPath**  <br>*optional*||string|
|**RestartCount**  <br>*optional*||integer|
|**SizeRootFs**  <br>*optional*|The total size of all the files in this container.|integer (int64)|
|**SizeRw**  <br>*optional*|The size of files that have been created or changed by this container.|integer (int64)|
|**State**  <br>*optional*|The state of the container.|[ContainerState](#containerstate)|


<a name="containerlogsoptions"></a>
### ContainerLogsOptions
The parameters to filter the log.


|Name|Description|Schema|
|---|---|---|
|**Details**  <br>*optional*|Show extra details provided to logs|boolean|
|**Follow**  <br>*optional*|Return logs as a stream|boolean|
|**ShowStderr**  <br>*optional*|Return logs from `stderr`|boolean|
|**ShowStdout**  <br>*optional*|Return logs from `stdout`|boolean|
|**Since**  <br>*optional*|Only return logs after this time, as a UNIX timestamp|string|
|**Tail**  <br>*optional*|Only reture this number of log lines from the end of the logs. Specify as an integer or `all` to output all log lines.|string|
|**Timestamps**  <br>*optional*|Add timestamps to every log line|boolean|
|**Until**  <br>*optional*|Only reture logs before this time, as a UNIX timestamp|string|


<a name="containerprocesslist"></a>
### ContainerProcessList
OK Response to ContainerTop operation


|Name|Description|Schema|
|---|---|---|
|**Processes**  <br>*optional*|Each process running in the container, where each is process is an array of values corresponding to the titles|< < string > array > array|
|**Titles**  <br>*optional*|The ps column titles|< string > array|


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


<a name="containerupgradeconfig"></a>
### ContainerUpgradeConfig
ContainerUpgradeConfig is used for API "POST /containers/upgrade".
It wraps all kinds of config used in container upgrade.
It can be used to encode client params in client and unmarshal request body in daemon side.

*Polymorphism* : Composition


|Name|Description|Schema|
|---|---|---|
|**ArgsEscaped**  <br>*optional*|Command is already escaped (Windows only)|boolean|
|**AttachStderr**  <br>*optional*|Whether to attach to `stderr`.  <br>**Default** : `true`|boolean|
|**AttachStdin**  <br>*optional*|Whether to attach to `stdin`.|boolean|
|**AttachStdout**  <br>*optional*|Whether to attach to `stdout`.  <br>**Default** : `true`|boolean|
|**Cmd**  <br>*optional*|Command to run specified an array of strings.|< string > array|
|**DiskQuota**  <br>*optional*|Set disk quota for container|< string, string > map|
|**Domainname**  <br>*optional*|The domain name to use for the container.|string|
|**Entrypoint**  <br>*optional*|The entry point for the container as a string or an array of strings.<br>If the array consists of exactly one empty string (`[""]`) then the entry point is reset to system default.|< string > array|
|**Env**  <br>*optional*|A list of environment variables to set inside the container in the form `["VAR=value", ...]`. A variable without `=` is removed from the environment, rather than to have an empty value.|< string > array|
|**ExposedPorts**  <br>*optional*|An object mapping ports to an empty object in the form:`{<port>/<tcp\|udp>: {}}`|< string, object > map|
|**HostConfig**  <br>*optional*||[HostConfig](#hostconfig)|
|**Hostname**  <br>*optional*|The hostname to use for the container, as a valid RFC 1123 hostname.  <br>**Minimum length** : `1`|string (hostname)|
|**Image**  <br>*required*|The name of the image to use when creating the container|string|
|**InitScript**  <br>*optional*|Initial script executed in container. The script will be executed before entrypoint or command|string|
|**Labels**  <br>*optional*|User-defined key/value metadata.|< string, string > map|
|**MacAddress**  <br>*optional*|MAC address of the container.|string|
|**NetworkDisabled**  <br>*optional*|Disable networking for the container.|boolean|
|**OnBuild**  <br>*optional*|`ONBUILD` metadata that were defined.|< string > array|
|**OpenStdin**  <br>*optional*|Open `stdin`|boolean|
|**Rich**  <br>*optional*|Whether to start container in rich container mode. (default false)|boolean|
|**RichMode**  <br>*optional*|Choose one rich container mode.(default dumb-init)|enum (dumb-init, sbin-init, systemd)|
|**Shell**  <br>*optional*|Shell for when `RUN`, `CMD`, and `ENTRYPOINT` uses a shell.|< string > array|
|**SpecAnnotation**  <br>*optional*|annotations send to runtime spec.|< string, string > map|
|**StdinOnce**  <br>*optional*|Close `stdin` after one attached client disconnects|boolean|
|**StopSignal**  <br>*optional*|Signal to stop a container as a string or unsigned integer.  <br>**Default** : `"SIGTERM"`|string|
|**StopTimeout**  <br>*optional*|Timeout to stop a container in seconds.|integer|
|**Tty**  <br>*optional*|Attach standard streams to a TTY, including `stdin` if it is not closed.|boolean|
|**User**  <br>*optional*|The user that commands are run as inside the container.|string|
|**Volumes**  <br>*optional*|An object mapping mount point paths inside the container to empty objects.|< string, object > map|
|**WorkingDir**  <br>*optional*|The working directory for commands to run in.|string|


<a name="daemonupdateconfig"></a>
### DaemonUpdateConfig

|Name|Description|Schema|
|---|---|---|
|**ImageProxy**  <br>*optional*|Image proxy used to pull image.|string|
|**Labels**  <br>*optional*|Labels indentified the attributes of daemon  <br>**Example** : `[ "storage=ssd", "zone=hangzhou" ]`|< string > array|


<a name="devicemapping"></a>
### DeviceMapping
A device mapping between the host and container


|Name|Description|Schema|
|---|---|---|
|**CgroupPermissions**  <br>*optional*|cgroup permissions of the device|string|
|**PathInContainer**  <br>*optional*|path in container of the device mapping|string|
|**PathOnHost**  <br>*optional*|path on host of the device mapping|string|


<a name="endpointipamconfig"></a>
### EndpointIPAMConfig
IPAM configurations for the endpoint


|Name|Description|Schema|
|---|---|---|
|**IPv4Address**  <br>*optional*|ipv4 address|string|
|**IPv6Address**  <br>*optional*|ipv6 address|string|
|**LinkLocalIPs**  <br>*optional*|link to the list of local ip|< string > array|


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
|**IPAMConfig**  <br>*optional*||[EndpointIPAMConfig](#endpointipamconfig)|
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
is a small subset of the Config struct that holds the configuration.


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


<a name="execcreateresp"></a>
### ExecCreateResp
contains response of Remote API POST "/containers/{name:.*}/exec".


|Name|Description|Schema|
|---|---|---|
|**Id**  <br>*optional*|ID is the exec ID|string|


<a name="execstartconfig"></a>
### ExecStartConfig
ExecStartConfig is a temp struct used by execStart.


|Name|Description|Schema|
|---|---|---|
|**Detach**  <br>*optional*|ExecStart will first check if it's detached|boolean|
|**Tty**  <br>*optional*|Check if there's a tty|boolean|


<a name="graphdriverdata"></a>
### GraphDriverData
Information about a container's graph driver.


|Name|Schema|
|---|---|
|**Data**  <br>*required*|< string, string > map|
|**Name**  <br>*required*|string|


<a name="hostconfig"></a>
### HostConfig
Container configuration that depends on the host we are running on

*Polymorphism* : Composition


|Name|Description|Schema|
|---|---|---|
|**AutoRemove**  <br>*optional*|Automatically remove the container when the container's process exits. This has no effect if `RestartPolicy` is set.|boolean|
|**Binds**  <br>*optional*|A list of volume bindings for this container. Each volume binding is a string in one of these forms:<br><br>- `host-src:container-dest` to bind-mount a host path into the container. Both `host-src`, and `container-dest` must be an _absolute_ path.<br>- `host-src:container-dest:ro` to make the bind mount read-only inside the container. Both `host-src`, and `container-dest` must be an _absolute_ path.<br>- `volume-name:container-dest` to bind-mount a volume managed by a volume driver into the container. `container-dest` must be an _absolute_ path.<br>- `volume-name:container-dest:ro` to mount the volume read-only inside the container.  `container-dest` must be an _absolute_ path.|< string > array|
|**BlkioDeviceReadBps**  <br>*optional*|Limit read rate (bytes per second) from a device, in the form `[{"Path": "device_path", "Rate": rate}]`.|< [ThrottleDevice](#throttledevice) > array|
|**BlkioDeviceReadIOps**  <br>*optional*|Limit read rate (IO per second) from a device, in the form `[{"Path": "device_path", "Rate": rate}]`.|< [ThrottleDevice](#throttledevice) > array|
|**BlkioDeviceWriteBps**  <br>*optional*|Limit write rate (bytes per second) to a device, in the form `[{"Path": "device_path", "Rate": rate}]`.|< [ThrottleDevice](#throttledevice) > array|
|**BlkioDeviceWriteIOps**  <br>*optional*|Limit write rate (IO per second) to a device, in the form `[{"Path": "device_path", "Rate": rate}]`.|< [ThrottleDevice](#throttledevice) > array|
|**BlkioWeight**  <br>*optional*|Block IO weight (relative weight).  <br>**Minimum value** : `0`  <br>**Maximum value** : `1000`|integer (uint16)|
|**BlkioWeightDevice**  <br>*optional*|Block IO weight (relative device weight) in the form `[{"Path": "device_path", "Weight": weight}]`.|< [WeightDevice](#weightdevice) > array|
|**CapAdd**  <br>*optional*|A list of kernel capabilities to add to the container.|< string > array|
|**CapDrop**  <br>*optional*|A list of kernel capabilities to drop from the container.|< string > array|
|**Cgroup**  <br>*optional*|Cgroup to use for the container.|string|
|**CgroupParent**  <br>*optional*|Path to `cgroups` under which the container's `cgroup` is created. If the path is not absolute, the path is considered to be relative to the `cgroups` path of the init process. Cgroups are created if they do not already exist.|string|
|**ConsoleSize**  <br>*optional*|Initial console size, as an `[height, width]` array. (Windows only)|< integer > array|
|**ContainerIDFile**  <br>*optional*|Path to a file where the container ID is written|string|
|**CpuCount**  <br>*optional*|The number of usable CPUs (Windows only).<br>On Windows Server containers, the processor resource controls are mutually exclusive. The order of precedence is `CPUCount` first, then `CPUShares`, and `CPUPercent` last.|integer (int64)|
|**CpuPercent**  <br>*optional*|The usable percentage of the available CPUs (Windows only).<br>On Windows Server containers, the processor resource controls are mutually exclusive. The order of precedence is `CPUCount` first, then `CPUShares`, and `CPUPercent` last.|integer (int64)|
|**CpuPeriod**  <br>*optional*|CPU CFS (Completely Fair Scheduler) period.<br>The length of a CPU period in microseconds.  <br>**Minimum value** : `1000`  <br>**Maximum value** : `1000000`|integer (int64)|
|**CpuQuota**  <br>*optional*|CPU CFS (Completely Fair Scheduler) quota.<br>Microseconds of CPU time that the container can get in a CPU period."  <br>**Minimum value** : `1000`|integer (int64)|
|**CpuRealtimePeriod**  <br>*optional*|The length of a CPU real-time period in microseconds. Set to 0 to allocate no time allocated to real-time tasks.|integer (int64)|
|**CpuRealtimeRuntime**  <br>*optional*|The length of a CPU real-time runtime in microseconds. Set to 0 to allocate no time allocated to real-time tasks.|integer (int64)|
|**CpuShares**  <br>*optional*|An integer value representing this container's relative CPU weight versus other containers.|integer|
|**CpusetCpus**  <br>*optional*|CPUs in which to allow execution (e.g., `0-3`, `0,1`)  <br>**Example** : `"0-3"`|string|
|**CpusetMems**  <br>*optional*|Memory nodes (MEMs) in which to allow execution (0-3, 0,1). Only effective on NUMA systems.|string|
|**DeviceCgroupRules**  <br>*optional*|a list of cgroup rules to apply to the container|< string > array|
|**Devices**  <br>*optional*|A list of devices to add to the container.|< [DeviceMapping](#devicemapping) > array|
|**DiskQuota**  <br>*optional*|Disk limit (in bytes).|integer (int64)|
|**Dns**  <br>*optional*|A list of DNS servers for the container to use.|< string > array|
|**DnsOptions**  <br>*optional*|A list of DNS options.|< string > array|
|**DnsSearch**  <br>*optional*|A list of DNS search domains.|< string > array|
|**EnableLxcfs**  <br>*optional*|Whether to enable lxcfs.|boolean|
|**ExtraHosts**  <br>*optional*|A list of hostnames/IP mappings to add to the container's `/etc/hosts` file. Specified in the form `["hostname:IP"]`.|< string > array|
|**GroupAdd**  <br>*optional*|A list of additional groups that the container process will run as.|< string > array|
|**IOMaximumBandwidth**  <br>*optional*|Maximum IO in bytes per second for the container system drive (Windows only)|integer (uint64)|
|**IOMaximumIOps**  <br>*optional*|Maximum IOps for the container system drive (Windows only)|integer (uint64)|
|**InitScript**  <br>*optional*|Initial script executed in container. The script will be executed before entrypoint or command|string|
|**IntelRdtL3Cbm**  <br>*optional*|IntelRdtL3Cbm specifies settings for Intel RDT/CAT group that the container is placed into to limit the resources (e.g., L3 cache) the container has available.|string|
|**IpcMode**  <br>*optional*|IPC sharing mode for the container. Possible values are:<br>- `"none"`: own private IPC namespace, with /dev/shm not mounted<br>- `"private"`: own private IPC namespace<br>- `"shareable"`: own private IPC namespace, with a possibility to share it with other containers<br>- `"container:<name\|id>"`: join another (shareable) container's IPC namespace<br>- `"host"`: use the host system's IPC namespace<br>If not specified, daemon default is used, which can either be `"private"`<br>or `"shareable"`, depending on daemon version and configuration.|string|
|**Isolation**  <br>*optional*|Isolation technology of the container. (Windows only)|enum (default, process, hyperv)|
|**KernelMemory**  <br>*optional*|Kernel memory limit in bytes.|integer (int64)|
|**Links**  <br>*optional*|A list of links for the container in the form `container_name:alias`.|< string > array|
|**LogConfig**  <br>*optional*|The logging configuration for this container|[LogConfig](#hostconfig-logconfig)|
|**Memory**  <br>*optional*|Memory limit in bytes.|integer|
|**MemoryExtra**  <br>*optional*|MemoryExtra is an integer value representing this container's memory high water mark percentage.<br>The range is in [0, 100].  <br>**Minimum value** : `0`  <br>**Maximum value** : `100`|integer (int64)|
|**MemoryForceEmptyCtl**  <br>*optional*|MemoryForceEmptyCtl represents whether to reclaim the page cache when deleting cgroup.  <br>**Minimum value** : `0`  <br>**Maximum value** : `1`|integer (int64)|
|**MemoryReservation**  <br>*optional*|Memory soft limit in bytes.|integer (int64)|
|**MemorySwap**  <br>*optional*|Total memory limit (memory + swap). Set as `-1` to enable unlimited swap.|integer (int64)|
|**MemorySwappiness**  <br>*optional*|Tune a container's memory swappiness behavior. Accepts an integer between 0 and 100.  <br>**Minimum value** : `-1`  <br>**Maximum value** : `100`|integer (int64)|
|**MemoryWmarkRatio**  <br>*optional*|MemoryWmarkRatio is an integer value representing this container's memory low water mark percentage. <br>The value of memory low water mark is memory.limit_in_bytes * MemoryWmarkRatio. The range is in [0, 100].  <br>**Minimum value** : `0`  <br>**Maximum value** : `100`|integer (int64)|
|**NanoCPUs**  <br>*optional*|CPU quota in units of 10<sup>-9</sup> CPUs.|integer (int64)|
|**NetworkMode**  <br>*optional*|Network mode to use for this container. Supported standard values are: `bridge`, `host`, `none`, and `container:<name\|id>`. Any other value is taken as a custom network's name to which this container should connect to.|string|
|**OomKillDisable**  <br>*optional*|Disable OOM Killer for the container.|boolean|
|**OomScoreAdj**  <br>*optional*|An integer value containing the score given to the container in order to tune OOM killer preferences.<br>The range is in [-1000, 1000].  <br>**Minimum value** : `-1000`  <br>**Maximum value** : `1000`|integer (int)|
|**PidMode**  <br>*optional*|Set the PID (Process) Namespace mode for the container. It can be either:<br>- `"container:<name\|id>"`: joins another container's PID namespace<br>- `"host"`: use the host's PID namespace inside the container|string|
|**PidsLimit**  <br>*optional*|Tune a container's pids limit. Set -1 for unlimited. Only on Linux 4.4 does this paramter support.|integer (int64)|
|**PortBindings**  <br>*optional*|A map of exposed container ports and the host port they should map to.|[PortMap](#portmap)|
|**Privileged**  <br>*optional*|Gives the container full access to the host.|boolean|
|**PublishAllPorts**  <br>*optional*|Allocates a random host port for all of a container's exposed ports.|boolean|
|**ReadonlyRootfs**  <br>*optional*|Mount the container's root filesystem as read only.|boolean|
|**RestartPolicy**  <br>*optional*|Restart policy to be used to manage the container|[RestartPolicy](#restartpolicy)|
|**Rich**  <br>*optional*|Whether to start container in rich container mode. (default false)|boolean|
|**RichMode**  <br>*optional*|Choose one rich container mode.(default dumb-init)|enum (dumb-init, sbin-init, systemd)|
|**Runtime**  <br>*optional*|Runtime to use with this container.|string|
|**ScheLatSwitch**  <br>*optional*|ScheLatSwitch enables scheduler latency count in cpuacct  <br>**Minimum value** : `0`  <br>**Maximum value** : `1`|integer (int64)|
|**SecurityOpt**  <br>*optional*|A list of string values to customize labels for MLS systems, such as SELinux.|< string > array|
|**ShmSize**  <br>*optional*|Size of `/dev/shm` in bytes. If omitted, the system uses 64MB.  <br>**Minimum value** : `0`|integer|
|**StorageOpt**  <br>*optional*|Storage driver options for this container, in the form `{"size": "120G"}`.|< string, string > map|
|**Sysctls**  <br>*optional*|A list of kernel parameters (sysctls) to set in the container. For example: `{"net.ipv4.ip_forward": "1"}`|< string, string > map|
|**Tmpfs**  <br>*optional*|A map of container directories which should be replaced by tmpfs mounts, and their corresponding mount options. For example: `{ "/run": "rw,noexec,nosuid,size=65536k" }`.|< string, string > map|
|**UTSMode**  <br>*optional*|UTS namespace to use for the container.|string|
|**Ulimits**  <br>*optional*|A list of resource limits to set in the container. For example: `{"Name": "nofile", "Soft": 1024, "Hard": 2048}`"|< [Ulimits](#hostconfig-ulimits) > array|
|**UsernsMode**  <br>*optional*|Sets the usernamespace mode for the container when usernamespace remapping option is enabled.|string|
|**VolumeDriver**  <br>*optional*|Driver that this container uses to mount volumes.|string|
|**VolumesFrom**  <br>*optional*|A list of volumes to inherit from another container, specified in the form `<container name>[:<ro\|rw>]`.|< string > array|

<a name="hostconfig-logconfig"></a>
**LogConfig**

|Name|Schema|
|---|---|
|**Config**  <br>*optional*|< string, string > map|
|**Type**  <br>*optional*|enum (json-file, syslog, journald, gelf, fluentd, awslogs, splunk, etwlogs, none)|

<a name="hostconfig-ulimits"></a>
**Ulimits**

|Name|Description|Schema|
|---|---|---|
|**Hard**  <br>*optional*|Hard limit|integer|
|**Name**  <br>*optional*|Name of ulimit|string|
|**Soft**  <br>*optional*|Soft limit|integer|


<a name="ipam"></a>
### IPAM
represents IP Address Management


|Name|Schema|
|---|---|
|**Config**  <br>*optional*|< [IPAMConfig](#ipamconfig) > array|
|**Driver**  <br>*optional*|string|
|**Options**  <br>*optional*|< string, string > map|


<a name="ipamconfig"></a>
### IPAMConfig
represents IPAM configurations


|Name|Schema|
|---|---|
|**AuxAddress**  <br>*optional*|< string, string > map|
|**Gateway**  <br>*optional*|string|
|**IPRange**  <br>*optional*|string|
|**Subnet**  <br>*optional*|string|


<a name="ipaddress"></a>
### IPAddress
Address represents an IPv4 or IPv6 IP address.


|Name|Description|Schema|
|---|---|---|
|**Addr**  <br>*optional*|IP address.|string|
|**PrefixLen**  <br>*optional*|Mask length of the IP address.|integer|


<a name="imageinfo"></a>
### ImageInfo
An object containing all details of an image at API side


|Name|Description|Schema|
|---|---|---|
|**Architecture**  <br>*optional*|the CPU architecture.|string|
|**Config**  <br>*optional*||[ContainerConfig](#containerconfig)|
|**CreatedAt**  <br>*optional*|time of image creation.|string|
|**Id**  <br>*optional*|ID of an image.|string|
|**Os**  <br>*optional*|the name of the operating system.|string|
|**RepoDigests**  <br>*optional*|repository with digest.|< string > array|
|**RepoTags**  <br>*optional*|repository with tag.|< string > array|
|**RootFS**  <br>*optional*|the rootfs key references the layer content addresses used by the image.|[RootFS](#imageinfo-rootfs)|
|**Size**  <br>*optional*|size of image's taking disk space.|integer|

<a name="imageinfo-rootfs"></a>
**RootFS**

|Name|Description|Schema|
|---|---|---|
|**BaseLayer**  <br>*optional*|the base layer content hash.|string|
|**Layers**  <br>*optional*|an array of layer content hashes|< string > array|
|**Type**  <br>*required*|type of the rootfs|string|


<a name="indexinfo"></a>
### IndexInfo
IndexInfo contains information about a registry.


|Name|Description|Schema|
|---|---|---|
|**Mirrors**  <br>*optional*|List of mirrors, expressed as URIs.  <br>**Example** : `[ "https://hub-mirror.corp.example.com:5000/" ]`|< string > array|
|**Name**  <br>*optional*|Name of the registry.|string|
|**Official**  <br>*optional*|Indicates whether this is an official registry.  <br>**Example** : `true`|boolean|
|**Secure**  <br>*optional*|Indicates if the the registry is part of the list of insecure<br>registries.<br><br>If `false`, the registry is insecure. Insecure registries accept<br>un-encrypted (HTTP) and/or untrusted (HTTPS with certificates from<br>unknown CAs) communication.<br><br>> **Warning**: Insecure registries can be useful when running a local<br>> registry. However, because its use creates security vulnerabilities<br>> it should ONLY be enabled for testing purposes. For increased<br>> security, users should add their CA to their system's list of<br>> trusted CAs instead of enabling this option.  <br>**Example** : `true`|boolean|


<a name="mountpoint"></a>
### MountPoint
A mount point inside a container


|Name|Schema|
|---|---|
|**CopyData**  <br>*optional*|boolean|
|**Destination**  <br>*optional*|string|
|**Driver**  <br>*optional*|string|
|**ID**  <br>*optional*|string|
|**Mode**  <br>*optional*|string|
|**Name**  <br>*optional*|string|
|**Named**  <br>*optional*|boolean|
|**Propagation**  <br>*optional*|string|
|**RW**  <br>*optional*|boolean|
|**Replace**  <br>*optional*|string|
|**Source**  <br>*optional*|string|
|**Type**  <br>*optional*|string|


<a name="networkcreate"></a>
### NetworkCreate
is the expected body of the "create network" http request message


|Name|Description|Schema|
|---|---|---|
|**CheckDuplicate**  <br>*optional*|CheckDuplicate is used to check the network is duplicate or not.|boolean|
|**Driver**  <br>*optional*|Driver means the network's driver.|string|
|**EnableIPv6**  <br>*optional*||boolean|
|**IPAM**  <br>*optional*||[IPAM](#ipam)|
|**Internal**  <br>*optional*|Internal checks the network is internal network or not.|boolean|
|**Labels**  <br>*optional*||< string, string > map|
|**Options**  <br>*optional*||< string, string > map|


<a name="networkcreateconfig"></a>
### NetworkCreateConfig
contains the request for the remote API: POST /networks/create

*Polymorphism* : Composition


|Name|Description|Schema|
|---|---|---|
|**CheckDuplicate**  <br>*optional*|CheckDuplicate is used to check the network is duplicate or not.|boolean|
|**Driver**  <br>*optional*|Driver means the network's driver.|string|
|**EnableIPv6**  <br>*optional*||boolean|
|**IPAM**  <br>*optional*||[IPAM](#ipam)|
|**Internal**  <br>*optional*|Internal checks the network is internal network or not.|boolean|
|**Labels**  <br>*optional*||< string, string > map|
|**Name**  <br>*optional*|Name is the name of the network.|string|
|**Options**  <br>*optional*||< string, string > map|


<a name="networkcreateresp"></a>
### NetworkCreateResp
contains the response for the remote API: POST /networks/create


|Name|Description|Schema|
|---|---|---|
|**Id**  <br>*optional*|ID is the id of the network.|string|
|**Warning**  <br>*optional*|Warning means the message of create network result.|string|


<a name="networkinfo"></a>
### NetworkInfo
NetworkInfo represents the configuration of a network


|Name|Description|Schema|
|---|---|---|
|**Driver**  <br>*optional*|Driver is the Driver name used to create the network|string|
|**Id**  <br>*optional*|ID uniquely identifies a network on a single machine|string|
|**Name**  <br>*optional*|Name is the name of the network.|string|
|**Scope**  <br>*optional*|Scope describes the level at which the network exists|string|


<a name="networkinspectresp"></a>
### NetworkInspectResp
is the expected body of the 'GET networks/{id}'' http request message


|Name|Description|Schema|
|---|---|---|
|**Driver**  <br>*optional*|Driver means the network's driver.|string|
|**EnableIPv6**  <br>*optional*|EnableIPv6 represents whether to enable IPv6.|boolean|
|**IPAM**  <br>*optional*|IPAM is the network's IP Address Management.|[IPAM](#ipam)|
|**Id**  <br>*optional*|ID uniquely identifies a network on a single machine|string|
|**Internal**  <br>*optional*|Internal checks the network is internal network or not.|boolean|
|**Labels**  <br>*optional*|Labels holds metadata specific to the network being created.|< string, string > map|
|**Name**  <br>*optional*|Name is the requested name of the network|string|
|**Options**  <br>*optional*|Options holds the network specific options to use for when creating the network.|< string, string > map|
|**Scope**  <br>*optional*|Scope describes the level at which the network exists.|string|


<a name="networklistresp"></a>
### NetworkListResp

|Name|Description|Schema|
|---|---|---|
|**Networks**  <br>*required*|List of networks|< [NetworkInfo](#networkinfo) > array|
|**Warnings**  <br>*required*|Warnings that occurred when fetching the list of networks|< string > array|


<a name="networksettings"></a>
### NetworkSettings
NetworkSettings exposes the network settings in the API.


|Name|Description|Schema|
|---|---|---|
|**Bridge**  <br>*optional*|Name of the network'a bridge (for example, `pouch-br`).  <br>**Example** : `"pouch-br"`|string|
|**HairpinMode**  <br>*optional*|Indicates if hairpin NAT should be enabled on the virtual interface  <br>**Example** : `false`|boolean|
|**LinkLocalIPv6Address**  <br>*optional*|IPv6 unicast address using the link-local prefix  <br>**Example** : `"fe80::42:acff:fe11:1"`|string|
|**LinkLocalIPv6PrefixLen**  <br>*optional*|Prefix length of the IPv6 unicast address.  <br>**Example** : `64`|integer|
|**Networks**  <br>*optional*|Information about all networks that the container is connected to|< string, [EndpointSettings](#endpointsettings) > map|
|**Ports**  <br>*optional*||[PortMap](#portmap)|
|**SandboxID**  <br>*optional*|SandboxID uniquely represents a container's network stack.  <br>**Example** : `"9d12daf2c33f5959c8bf90aa513e4f65b561738661003029ec84830cd503a0c3"`|string|
|**SandboxKey**  <br>*optional*|SandboxKey identifies the sandbox  <br>**Example** : `"/var/run/pouch/netns/8ab54b426c38"`|string|
|**SecondaryIPAddresses**  <br>*optional*||< [IPAddress](#ipaddress) > array|
|**SecondaryIPv6Addresses**  <br>*optional*||< [IPAddress](#ipaddress) > array|


<a name="networkingconfig"></a>
### NetworkingConfig
Configuration for a network used to create a container.

*Type* : object


<a name="portbinding"></a>
### PortBinding
PortBinding represents a binding between a host IP address and a host port


|Name|Description|Schema|
|---|---|---|
|**HostIp**  <br>*optional*|Host IP address that the container's port is mapped to.  <br>**Example** : `"127.0.0.1"`|string|
|**HostPort**  <br>*optional*|Host port number that the container's port is mapped to.  <br>**Example** : `"4443"`|string|


<a name="portmap"></a>
### PortMap
PortMap describes the mapping of container ports to host ports, using the
container's port-number and protocol as key in the format `<port>/<protocol>`,
for example, `80/udp`.

If a container's port is mapped for both `tcp` and `udp`, two separate
entries are added to the mapping table.

*Type* : < string, < [PortBinding](#portbinding) > array > map


<a name="processconfig"></a>
### ProcessConfig
ExecProcessConfig holds information about the exec process.


|Name|Schema|
|---|---|
|**arguments**  <br>*optional*|< string > array|
|**entrypoint**  <br>*optional*|string|
|**privileged**  <br>*optional*|boolean|
|**tty**  <br>*optional*|boolean|
|**user**  <br>*optional*|string|


<a name="registryserviceconfig"></a>
### RegistryServiceConfig
RegistryServiceConfig stores daemon registry services configuration.


|Name|Description|Schema|
|---|---|---|
|**AllowNondistributableArtifactsCIDRs**  <br>*optional*|List of IP ranges to which nondistributable artifacts can be pushed,<br>using the CIDR syntax [RFC 4632](https://tools.ietf.org/html/4632).<br><br>Some images contain artifacts whose distribution is restricted by license.<br>When these images are pushed to a registry, restricted artifacts are not<br>included.<br><br>This configuration override this behavior, and enables the daemon to<br>push nondistributable artifacts to all registries whose resolved IP<br>address is within the subnet described by the CIDR syntax.<br><br>This option is useful when pushing images containing<br>nondistributable artifacts to a registry on an air-gapped network so<br>hosts on that network can pull the images without connecting to<br>another server.<br><br>> **Warning**: Nondistributable artifacts typically have restrictions<br>> on how and where they can be distributed and shared. Only use this<br>> feature to push artifacts to private registries and ensure that you<br>> are in compliance with any terms that cover redistributing<br>> nondistributable artifacts.  <br>**Example** : `[ "::1/128", "127.0.0.0/8" ]`|< string > array|
|**AllowNondistributableArtifactsHostnames**  <br>*optional*|List of registry hostnames to which nondistributable artifacts can be<br>pushed, using the format `<hostname>[:<port>]` or `<IP address>[:<port>]`.<br><br>Some images (for example, Windows base images) contain artifacts<br>whose distribution is restricted by license. When these images are<br>pushed to a registry, restricted artifacts are not included.<br><br>This configuration override this behavior for the specified<br>registries.<br><br>This option is useful when pushing images containing<br>nondistributable artifacts to a registry on an air-gapped network so<br>hosts on that network can pull the images without connecting to<br>another server.<br><br>> **Warning**: Nondistributable artifacts typically have restrictions<br>> on how and where they can be distributed and shared. Only use this<br>> feature to push artifacts to private registries and ensure that you<br>> are in compliance with any terms that cover redistributing<br>> nondistributable artifacts.  <br>**Example** : `[ "registry.internal.corp.example.com:3000", "[2001:db8:a0b:12f0::1]:443" ]`|< string > array|
|**IndexConfigs**  <br>*optional*|**Example** : `{<br>  "127.0.0.1:5000" : {<br>    "Name" : "127.0.0.1:5000",<br>    "Mirrors" : [ ],<br>    "Secure" : false,<br>    "Official" : false<br>  },<br>  "[2001:db8:a0b:12f0::1]:80" : {<br>    "Name" : "[2001:db8:a0b:12f0::1]:80",<br>    "Mirrors" : [ ],<br>    "Secure" : false,<br>    "Official" : false<br>  },<br>  "registry.internal.corp.example.com:3000" : {<br>    "Name" : "registry.internal.corp.example.com:3000",<br>    "Mirrors" : [ ],<br>    "Secure" : false,<br>    "Official" : false<br>  }<br>}`|< string, [IndexInfo](#indexinfo) > map|
|**InsecureRegistryCIDRs**  <br>*optional*|List of IP ranges of insecure registries, using the CIDR syntax<br>([RFC 4632](https://tools.ietf.org/html/4632)). Insecure registries<br>accept un-encrypted (HTTP) and/or untrusted (HTTPS with certificates<br>from unknown CAs) communication.<br><br>By default, local registries (`127.0.0.0/8`) are configured as<br>insecure. All other registries are secure. Communicating with an<br>insecure registry is not possible if the daemon assumes that registry<br>is secure.<br><br>This configuration override this behavior, insecure communication with<br>registries whose resolved IP address is within the subnet described by<br>the CIDR syntax.<br><br>Registries can also be marked insecure by hostname. Those registries<br>are listed under `IndexConfigs` and have their `Secure` field set to<br>`false`.<br><br>> **Warning**: Using this option can be useful when running a local<br>> registry, but introduces security vulnerabilities. This option<br>> should therefore ONLY be used for testing purposes. For increased<br>> security, users should add their CA to their system's list of trusted<br>> CAs instead of enabling this option.  <br>**Example** : `[ "::1/128", "127.0.0.0/8" ]`|< string > array|
|**Mirrors**  <br>*optional*|List of registry URLs that act as a mirror for the official registry.  <br>**Example** : `[ "https://hub-mirror.corp.example.com:5000/", "https://[2001:db8:a0b:12f0::1]/" ]`|< string > array|


<a name="resizeoptions"></a>
### ResizeOptions
options of resizing container tty size


|Name|Schema|
|---|---|
|**Height**  <br>*optional*|integer|
|**Width**  <br>*optional*|integer|


<a name="resources"></a>
### Resources
A container's resources (cgroups config, ulimits, etc)


|Name|Description|Schema|
|---|---|---|
|**BlkioDeviceReadBps**  <br>*optional*|Limit read rate (bytes per second) from a device, in the form `[{"Path": "device_path", "Rate": rate}]`.|< [ThrottleDevice](#throttledevice) > array|
|**BlkioDeviceReadIOps**  <br>*optional*|Limit read rate (IO per second) from a device, in the form `[{"Path": "device_path", "Rate": rate}]`.|< [ThrottleDevice](#throttledevice) > array|
|**BlkioDeviceWriteBps**  <br>*optional*|Limit write rate (bytes per second) to a device, in the form `[{"Path": "device_path", "Rate": rate}]`.|< [ThrottleDevice](#throttledevice) > array|
|**BlkioDeviceWriteIOps**  <br>*optional*|Limit write rate (IO per second) to a device, in the form `[{"Path": "device_path", "Rate": rate}]`.|< [ThrottleDevice](#throttledevice) > array|
|**BlkioWeight**  <br>*optional*|Block IO weight (relative weight).  <br>**Minimum value** : `0`  <br>**Maximum value** : `1000`|integer (uint16)|
|**BlkioWeightDevice**  <br>*optional*|Block IO weight (relative device weight) in the form `[{"Path": "device_path", "Weight": weight}]`.|< [WeightDevice](#weightdevice) > array|
|**CgroupParent**  <br>*optional*|Path to `cgroups` under which the container's `cgroup` is created. If the path is not absolute, the path is considered to be relative to the `cgroups` path of the init process. Cgroups are created if they do not already exist.|string|
|**CpuCount**  <br>*optional*|The number of usable CPUs (Windows only).<br>On Windows Server containers, the processor resource controls are mutually exclusive. The order of precedence is `CPUCount` first, then `CPUShares`, and `CPUPercent` last.|integer (int64)|
|**CpuPercent**  <br>*optional*|The usable percentage of the available CPUs (Windows only).<br>On Windows Server containers, the processor resource controls are mutually exclusive. The order of precedence is `CPUCount` first, then `CPUShares`, and `CPUPercent` last.|integer (int64)|
|**CpuPeriod**  <br>*optional*|CPU CFS (Completely Fair Scheduler) period.<br>The length of a CPU period in microseconds.  <br>**Minimum value** : `1000`  <br>**Maximum value** : `1000000`|integer (int64)|
|**CpuQuota**  <br>*optional*|CPU CFS (Completely Fair Scheduler) quota.<br>Microseconds of CPU time that the container can get in a CPU period."  <br>**Minimum value** : `1000`|integer (int64)|
|**CpuRealtimePeriod**  <br>*optional*|The length of a CPU real-time period in microseconds. Set to 0 to allocate no time allocated to real-time tasks.|integer (int64)|
|**CpuRealtimeRuntime**  <br>*optional*|The length of a CPU real-time runtime in microseconds. Set to 0 to allocate no time allocated to real-time tasks.|integer (int64)|
|**CpuShares**  <br>*optional*|An integer value representing this container's relative CPU weight versus other containers.|integer|
|**CpusetCpus**  <br>*optional*|CPUs in which to allow execution (e.g., `0-3`, `0,1`)  <br>**Example** : `"0-3"`|string|
|**CpusetMems**  <br>*optional*|Memory nodes (MEMs) in which to allow execution (0-3, 0,1). Only effective on NUMA systems.|string|
|**DeviceCgroupRules**  <br>*optional*|a list of cgroup rules to apply to the container|< string > array|
|**Devices**  <br>*optional*|A list of devices to add to the container.|< [DeviceMapping](#devicemapping) > array|
|**DiskQuota**  <br>*optional*|Disk limit (in bytes).|integer (int64)|
|**IOMaximumBandwidth**  <br>*optional*|Maximum IO in bytes per second for the container system drive (Windows only)|integer (uint64)|
|**IOMaximumIOps**  <br>*optional*|Maximum IOps for the container system drive (Windows only)|integer (uint64)|
|**IntelRdtL3Cbm**  <br>*optional*|IntelRdtL3Cbm specifies settings for Intel RDT/CAT group that the container is placed into to limit the resources (e.g., L3 cache) the container has available.|string|
|**KernelMemory**  <br>*optional*|Kernel memory limit in bytes.|integer (int64)|
|**Memory**  <br>*optional*|Memory limit in bytes.|integer|
|**MemoryExtra**  <br>*optional*|MemoryExtra is an integer value representing this container's memory high water mark percentage.<br>The range is in [0, 100].  <br>**Minimum value** : `0`  <br>**Maximum value** : `100`|integer (int64)|
|**MemoryForceEmptyCtl**  <br>*optional*|MemoryForceEmptyCtl represents whether to reclaim the page cache when deleting cgroup.  <br>**Minimum value** : `0`  <br>**Maximum value** : `1`|integer (int64)|
|**MemoryReservation**  <br>*optional*|Memory soft limit in bytes.|integer (int64)|
|**MemorySwap**  <br>*optional*|Total memory limit (memory + swap). Set as `-1` to enable unlimited swap.|integer (int64)|
|**MemorySwappiness**  <br>*optional*|Tune a container's memory swappiness behavior. Accepts an integer between 0 and 100.  <br>**Minimum value** : `-1`  <br>**Maximum value** : `100`|integer (int64)|
|**MemoryWmarkRatio**  <br>*optional*|MemoryWmarkRatio is an integer value representing this container's memory low water mark percentage. <br>The value of memory low water mark is memory.limit_in_bytes * MemoryWmarkRatio. The range is in [0, 100].  <br>**Minimum value** : `0`  <br>**Maximum value** : `100`|integer (int64)|
|**NanoCPUs**  <br>*optional*|CPU quota in units of 10<sup>-9</sup> CPUs.|integer (int64)|
|**OomKillDisable**  <br>*optional*|Disable OOM Killer for the container.|boolean|
|**PidsLimit**  <br>*optional*|Tune a container's pids limit. Set -1 for unlimited. Only on Linux 4.4 does this paramter support.|integer (int64)|
|**ScheLatSwitch**  <br>*optional*|ScheLatSwitch enables scheduler latency count in cpuacct  <br>**Minimum value** : `0`  <br>**Maximum value** : `1`|integer (int64)|
|**Ulimits**  <br>*optional*|A list of resource limits to set in the container. For example: `{"Name": "nofile", "Soft": 1024, "Hard": 2048}`"|< [Ulimits](#resources-ulimits) > array|

<a name="resources-ulimits"></a>
**Ulimits**

|Name|Description|Schema|
|---|---|---|
|**Hard**  <br>*optional*|Hard limit|integer|
|**Name**  <br>*optional*|Name of ulimit|string|
|**Soft**  <br>*optional*|Soft limit|integer|


<a name="restartpolicy"></a>
### RestartPolicy
Define container's restart policy


|Name|Schema|
|---|---|
|**MaximumRetryCount**  <br>*optional*|integer|
|**Name**  <br>*optional*|string|


<a name="runtime"></a>
### Runtime
Runtime describes an [OCI compliant](https://github.com/opencontainers/runtime-spec)
runtime.

The runtime is invoked by the daemon via the `containerd` daemon. OCI
runtimes act as an interface to the Linux kernel namespaces, cgroups,
and SELinux.


|Name|Description|Schema|
|---|---|---|
|**path**  <br>*optional*|Name and, optional, path, of the OCI executable binary.<br><br>If the path is omitted, the daemon searches the host's `$PATH` for the<br>binary and uses the first result.  <br>**Example** : `"/usr/local/bin/my-oci-runtime"`|string|
|**runtimeArgs**  <br>*optional*|List of command-line arguments to pass to the runtime when invoked.  <br>**Example** : `[ "--debug", "--systemd-cgroup=false" ]`|< string > array|


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

|Name|Description|Schema|
|---|---|---|
|**Architecture**  <br>*optional*|Hardware architecture of the host, as returned by the Go runtime<br>(`GOARCH`).<br><br>A full list of possible values can be found in the [Go documentation](https://golang.org/doc/install/source#environment).  <br>**Example** : `"x86_64"`|string|
|**CgroupDriver**  <br>*optional*|The driver to use for managing cgroups.  <br>**Default** : `"cgroupfs"`  <br>**Example** : `"cgroupfs"`|enum (cgroupfs, systemd)|
|**ContainerdCommit**  <br>*optional*||[Commit](#commit)|
|**Containers**  <br>*optional*|Total number of containers on the host.  <br>**Example** : `14`|integer|
|**ContainersPaused**  <br>*optional*|Number of containers with status `"paused"`.  <br>**Example** : `1`|integer|
|**ContainersRunning**  <br>*optional*|Number of containers with status `"running"`.  <br>**Example** : `3`|integer|
|**ContainersStopped**  <br>*optional*|Number of containers with status `"stopped"`.  <br>**Example** : `10`|integer|
|**Debug**  <br>*optional*|Indicates if the daemon is running in debug-mode / with debug-level logging enabled.  <br>**Example** : `true`|boolean|
|**DefaultRegistry**  <br>*optional*|default registry can be defined by user.|string|
|**DefaultRuntime**  <br>*optional*|Name of the default OCI runtime that is used when starting containers.<br>The default can be overridden per-container at create time.  <br>**Default** : `"runc"`  <br>**Example** : `"runc"`|string|
|**Driver**  <br>*optional*|Name of the storage driver in use.  <br>**Example** : `"overlay2"`|string|
|**DriverStatus**  <br>*optional*|Information specific to the storage driver, provided as<br>"label" / "value" pairs.<br><br>This information is provided by the storage driver, and formatted<br>in a way consistent with the output of `pouch info` on the command<br>line.<br><br><p><br /></p><br><br>> **Note**: The information returned in this field, including the<br>> formatting of values and labels, should not be considered stable,<br>> and may change without notice.  <br>**Example** : `[ [ "Backing Filesystem", "extfs" ], [ "Supports d_type", "true" ], [ "Native Overlay Diff", "true" ] ]`|< < string > array > array|
|**ExperimentalBuild**  <br>*optional*|Indicates if experimental features are enabled on the daemon.  <br>**Example** : `true`|boolean|
|**HttpProxy**  <br>*optional*|HTTP-proxy configured for the daemon. This value is obtained from the<br>[`HTTP_PROXY`](https://www.gnu.org/software/wget/manual/html_node/Proxies.html) environment variable.<br><br>Containers do not automatically inherit this configuration.  <br>**Example** : `"http://user:pass@proxy.corp.example.com:8080"`|string|
|**HttpsProxy**  <br>*optional*|HTTPS-proxy configured for the daemon. This value is obtained from the<br>[`HTTPS_PROXY`](https://www.gnu.org/software/wget/manual/html_node/Proxies.html) environment variable.<br><br>Containers do not automatically inherit this configuration.  <br>**Example** : `"https://user:pass@proxy.corp.example.com:4443"`|string|
|**ID**  <br>*optional*|Unique identifier of the daemon.<br><br><p><br /></p><br><br>> **Note**: The format of the ID itself is not part of the API, and<br>> should not be considered stable.  <br>**Example** : `"7TRN:IPZB:QYBB:VPBQ:UMPP:KARE:6ZNR:XE6T:7EWV:PKF4:ZOJD:TPYS"`|string|
|**Images**  <br>*optional*|Total number of images on the host.<br><br>Both _tagged_ and _untagged_ (dangling) images are counted.  <br>**Example** : `508`|integer|
|**IndexServerAddress**  <br>*optional*|Address / URL of the index server that is used for image search,<br>and as a default for user authentication.|string|
|**KernelVersion**  <br>*optional*|Kernel version of the host.<br>On Linux, this information obtained from `uname`.|string|
|**Labels**  <br>*optional*|User-defined labels (key/value metadata) as set on the daemon.  <br>**Example** : `[ "storage=ssd", "production" ]`|< string > array|
|**ListenAddresses**  <br>*optional*|List of addresses the pouchd listens on  <br>**Example** : `[ [ "unix:///var/run/pouchd.sock", "tcp://0.0.0.0:4243" ] ]`|< string > array|
|**LiveRestoreEnabled**  <br>*optional*|Indicates if live restore is enabled.<br>If enabled, containers are kept running when the daemon is shutdown<br>or upon daemon start if running containers are detected.  <br>**Default** : `false`  <br>**Example** : `false`|boolean|
|**LoggingDriver**  <br>*optional*|The logging driver to use as a default for new containers.|string|
|**MemTotal**  <br>*optional*|Total amount of physical memory available on the host, in kilobytes (kB).  <br>**Example** : `2095882240`|integer (int64)|
|**NCPU**  <br>*optional*|The number of logical CPUs usable by the daemon.<br><br>The number of available CPUs is checked by querying the operating<br>system when the daemon starts. Changes to operating system CPU<br>allocation after the daemon is started are not reflected.  <br>**Example** : `4`|integer|
|**Name**  <br>*optional*|Hostname of the host.  <br>**Example** : `"node5.corp.example.com"`|string|
|**OSType**  <br>*optional*|Generic type of the operating system of the host, as returned by the<br>Go runtime (`GOOS`).<br><br>Currently returned value is "linux". A full list of<br>possible values can be found in the [Go documentation](https://golang.org/doc/install/source#environment).  <br>**Example** : `"linux"`|string|
|**OperatingSystem**  <br>*optional*|Name of the host's operating system, for example: "Ubuntu 16.04.2 LTS".  <br>**Example** : `"Alpine Linux v3.5"`|string|
|**PouchRootDir**  <br>*optional*|Root directory of persistent Pouch state.<br><br>Defaults to `/var/lib/pouch` on Linux.  <br>**Example** : `"/var/lib/pouch"`|string|
|**RegistryConfig**  <br>*optional*||[RegistryServiceConfig](#registryserviceconfig)|
|**RuncCommit**  <br>*optional*||[Commit](#commit)|
|**Runtimes**  <br>*optional*|List of [OCI compliant](https://github.com/opencontainers/runtime-spec)<br>runtimes configured on the daemon. Keys hold the "name" used to<br>reference the runtime.<br><br>The Pouch daemon relies on an OCI compliant runtime (invoked via the<br>`containerd` daemon) as its interface to the Linux kernel namespaces,<br>cgroups, and SELinux.<br><br>The default runtime is `runc`, and automatically configured. Additional<br>runtimes can be configured by the user and will be listed here.  <br>**Example** : `{<br>  "runc" : {<br>    "path" : "pouch-runc"<br>  },<br>  "runc-master" : {<br>    "path" : "/go/bin/runc"<br>  },<br>  "custom" : {<br>    "path" : "/usr/local/bin/my-oci-runtime",<br>    "runtimeArgs" : [ "--debug", "--systemd-cgroup=false" ]<br>  }<br>}`|< string, [Runtime](#runtime) > map|
|**SecurityOptions**  <br>*optional*|List of security features that are enabled on the daemon, such as<br>apparmor, seccomp, SELinux, and user-namespaces (userns).<br><br>Additional configuration options for each security feature may<br>be present, and are included as a comma-separated list of key/value<br>pairs.  <br>**Example** : `[ "name=apparmor", "name=seccomp,profile=default", "name=selinux", "name=userns" ]`|< string > array|
|**ServerVersion**  <br>*optional*|Version string of the daemon.  <br>**Example** : `"17.06.0-ce"`|string|


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


<a name="throttledevice"></a>
### ThrottleDevice

|Name|Description|Schema|
|---|---|---|
|**Path**  <br>*optional*|Device path|string|
|**Rate**  <br>*optional*|Rate  <br>**Minimum value** : `0`|integer (uint64)|


<a name="updateconfig"></a>
### UpdateConfig
UpdateConfig holds the mutable attributes of a Container. Those attributes can be updated at runtime.

*Polymorphism* : Composition


|Name|Description|Schema|
|---|---|---|
|**BlkioDeviceReadBps**  <br>*optional*|Limit read rate (bytes per second) from a device, in the form `[{"Path": "device_path", "Rate": rate}]`.|< [ThrottleDevice](#throttledevice) > array|
|**BlkioDeviceReadIOps**  <br>*optional*|Limit read rate (IO per second) from a device, in the form `[{"Path": "device_path", "Rate": rate}]`.|< [ThrottleDevice](#throttledevice) > array|
|**BlkioDeviceWriteBps**  <br>*optional*|Limit write rate (bytes per second) to a device, in the form `[{"Path": "device_path", "Rate": rate}]`.|< [ThrottleDevice](#throttledevice) > array|
|**BlkioDeviceWriteIOps**  <br>*optional*|Limit write rate (IO per second) to a device, in the form `[{"Path": "device_path", "Rate": rate}]`.|< [ThrottleDevice](#throttledevice) > array|
|**BlkioWeight**  <br>*optional*|Block IO weight (relative weight).  <br>**Minimum value** : `0`  <br>**Maximum value** : `1000`|integer (uint16)|
|**BlkioWeightDevice**  <br>*optional*|Block IO weight (relative device weight) in the form `[{"Path": "device_path", "Weight": weight}]`.|< [WeightDevice](#weightdevice) > array|
|**CgroupParent**  <br>*optional*|Path to `cgroups` under which the container's `cgroup` is created. If the path is not absolute, the path is considered to be relative to the `cgroups` path of the init process. Cgroups are created if they do not already exist.|string|
|**CpuCount**  <br>*optional*|The number of usable CPUs (Windows only).<br>On Windows Server containers, the processor resource controls are mutually exclusive. The order of precedence is `CPUCount` first, then `CPUShares`, and `CPUPercent` last.|integer (int64)|
|**CpuPercent**  <br>*optional*|The usable percentage of the available CPUs (Windows only).<br>On Windows Server containers, the processor resource controls are mutually exclusive. The order of precedence is `CPUCount` first, then `CPUShares`, and `CPUPercent` last.|integer (int64)|
|**CpuPeriod**  <br>*optional*|CPU CFS (Completely Fair Scheduler) period.<br>The length of a CPU period in microseconds.  <br>**Minimum value** : `1000`  <br>**Maximum value** : `1000000`|integer (int64)|
|**CpuQuota**  <br>*optional*|CPU CFS (Completely Fair Scheduler) quota.<br>Microseconds of CPU time that the container can get in a CPU period."  <br>**Minimum value** : `1000`|integer (int64)|
|**CpuRealtimePeriod**  <br>*optional*|The length of a CPU real-time period in microseconds. Set to 0 to allocate no time allocated to real-time tasks.|integer (int64)|
|**CpuRealtimeRuntime**  <br>*optional*|The length of a CPU real-time runtime in microseconds. Set to 0 to allocate no time allocated to real-time tasks.|integer (int64)|
|**CpuShares**  <br>*optional*|An integer value representing this container's relative CPU weight versus other containers.|integer|
|**CpusetCpus**  <br>*optional*|CPUs in which to allow execution (e.g., `0-3`, `0,1`)  <br>**Example** : `"0-3"`|string|
|**CpusetMems**  <br>*optional*|Memory nodes (MEMs) in which to allow execution (0-3, 0,1). Only effective on NUMA systems.|string|
|**DeviceCgroupRules**  <br>*optional*|a list of cgroup rules to apply to the container|< string > array|
|**Devices**  <br>*optional*|A list of devices to add to the container.|< [DeviceMapping](#devicemapping) > array|
|**DiskQuota**  <br>*optional*|Disk limit (in bytes).|integer (int64)|
|**Env**  <br>*optional*|A list of environment variables to set inside the container in the form `["VAR=value", ...]`. A variable without `=` is removed from the environment, rather than to have an empty value.|< string > array|
|**IOMaximumBandwidth**  <br>*optional*|Maximum IO in bytes per second for the container system drive (Windows only)|integer (uint64)|
|**IOMaximumIOps**  <br>*optional*|Maximum IOps for the container system drive (Windows only)|integer (uint64)|
|**Image**  <br>*optional*|Image ID or Name|string|
|**IntelRdtL3Cbm**  <br>*optional*|IntelRdtL3Cbm specifies settings for Intel RDT/CAT group that the container is placed into to limit the resources (e.g., L3 cache) the container has available.|string|
|**KernelMemory**  <br>*optional*|Kernel memory limit in bytes.|integer (int64)|
|**Labels**  <br>*optional*|List of labels set to container.|< string, string > map|
|**Memory**  <br>*optional*|Memory limit in bytes.|integer|
|**MemoryExtra**  <br>*optional*|MemoryExtra is an integer value representing this container's memory high water mark percentage.<br>The range is in [0, 100].  <br>**Minimum value** : `0`  <br>**Maximum value** : `100`|integer (int64)|
|**MemoryForceEmptyCtl**  <br>*optional*|MemoryForceEmptyCtl represents whether to reclaim the page cache when deleting cgroup.  <br>**Minimum value** : `0`  <br>**Maximum value** : `1`|integer (int64)|
|**MemoryReservation**  <br>*optional*|Memory soft limit in bytes.|integer (int64)|
|**MemorySwap**  <br>*optional*|Total memory limit (memory + swap). Set as `-1` to enable unlimited swap.|integer (int64)|
|**MemorySwappiness**  <br>*optional*|Tune a container's memory swappiness behavior. Accepts an integer between 0 and 100.  <br>**Minimum value** : `-1`  <br>**Maximum value** : `100`|integer (int64)|
|**MemoryWmarkRatio**  <br>*optional*|MemoryWmarkRatio is an integer value representing this container's memory low water mark percentage. <br>The value of memory low water mark is memory.limit_in_bytes * MemoryWmarkRatio. The range is in [0, 100].  <br>**Minimum value** : `0`  <br>**Maximum value** : `100`|integer (int64)|
|**NanoCPUs**  <br>*optional*|CPU quota in units of 10<sup>-9</sup> CPUs.|integer (int64)|
|**OomKillDisable**  <br>*optional*|Disable OOM Killer for the container.|boolean|
|**PidsLimit**  <br>*optional*|Tune a container's pids limit. Set -1 for unlimited. Only on Linux 4.4 does this paramter support.|integer (int64)|
|**RestartPolicy**  <br>*optional*||[RestartPolicy](#restartpolicy)|
|**ScheLatSwitch**  <br>*optional*|ScheLatSwitch enables scheduler latency count in cpuacct  <br>**Minimum value** : `0`  <br>**Maximum value** : `1`|integer (int64)|
|**Ulimits**  <br>*optional*|A list of resource limits to set in the container. For example: `{"Name": "nofile", "Soft": 1024, "Hard": 2048}`"|< [Ulimits](#updateconfig-ulimits) > array|

<a name="updateconfig-ulimits"></a>
**Ulimits**

|Name|Description|Schema|
|---|---|---|
|**Hard**  <br>*optional*|Hard limit|integer|
|**Name**  <br>*optional*|Name of ulimit|string|
|**Soft**  <br>*optional*|Soft limit|integer|


<a name="volumecreateconfig"></a>
### VolumeCreateConfig
config used to create a volume


|Name|Description|Schema|
|---|---|---|
|**Driver**  <br>*optional*|Name of the volume driver to use.  <br>**Default** : `"local"`|string|
|**DriverOpts**  <br>*optional*|A mapping of driver options and values. These options are passed directly to the driver and are driver specific.|< string, string > map|
|**Labels**  <br>*optional*|User-defined key/value metadata.|< string, string > map|
|**Name**  <br>*optional*|The new volume's name. If not specified, Pouch generates a name.|string|


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


<a name="weightdevice"></a>
### WeightDevice
Weight for BlockIO Device


|Name|Description|Schema|
|---|---|---|
|**Path**  <br>*optional*|Weight Device|string|
|**Weight**  <br>*optional*|**Minimum value** : `0`|integer (uint16)|





