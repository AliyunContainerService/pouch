# API

Directory /apis currently contains the following three kinds of things:

* API definitions via [swagger.yml](swagger.yml);
* struct in `/apis/types` used in restful API between Pouch Client and Server, both auto generated and manual added;
* api server implemented for pouch daemon.

## Generated and Manual Types

Pouch has both generated types via tool [swagger](https://swagger.io) and mamually added types. Currently both generated and manually added are located in directory `/apis/types`.

## Hack APIs

Swagger helps us to unify the API standard, and swagger.yml is the only entry to update API. If you wish to hack Pouch on API side, here we have several points to guide you.

### Install Go Swagger

Swagger is a simple yet powerful representation of your RESTful API. [go-swagger](https://github.com/go-swagger/go-swagger) is a tool implemented for Golang. We could install go-swagger by following official [go-swagger installation](https://github.com/go-swagger/go-swagger#binary-distribution). 

### Design API 

With swagger installed in you local $PATH, you could start to design or update API of Pouch. Here we need to underline that every change for API should first in [swagger.yml](swagger.yml) of Pouch porject.

Let's us take an example:

> add a new API "GET /version" to get the version of the pouchd

To finish this, we can simplify work to two parts:

* define API path, including request method, request parameters, request content type and response status code, reposonse body and so on;
* define API structs used by both pouch client and pouch daemon.

For part one, we should add the following content in `paths`:

``` 
  /version:
    get:
      summary: "Get Pouchd version"
      description: ""
      responses:
        200:
          schema:
            $ref: '#/definitions/SystemVersion'
          description: "no error"
        500:
          description: "server error"
          schema:
            $ref: '#/definitions/Error'
```

For part two, we need to add struct or object `SystemVersion` and `Error` in `definitions`. Added content could be seen like below:

```
  SystemVersion:
    type: "object"
    properties:
      Version:
        type: "string"
        description: "version of Pouch Daemon"
        example: "0.1.2"
      ApiVersion:
        type: "string"
        description: "Api Version held by daemon"
        example: ""
      ......
      KernelVersion:
        type: "string"
        description: "Operating system kernel version"
        example: "3.13.0-106-generic"
      BuildTime:
        type: "string"
        description: "The time when this binary of daemon is built"
        example: "2017-08-29T17:41:57.729792388+00:00"
```

Daemon side can use `SystemVersion` easily as well. 

### Auto-generate API struct

With above two parts finished in swagger.yml, the following step is auto-generating source code for API structs. We can do that with a script [generate-swagger-models.sh](../hack/generate-swagger-models.sh). We can execute the script via:

``` shell
pouch (master) $ ./hack/generate-swagger-models.sh
```

Then details you updated would auto generate source code in directory pouch/apis/types. With the example above, you will find file `system_version.go` existing there.

### Integrate generated struct

With API structs generated, you can integrate these structs in both client side and daemon side. In client side, you could code like this integrating `SystemVersion`:

```
// SystemVersion requests daemon for system version.
func (cli *Client) SystemVersion() (*types.SystemVersion, error) {
	resp, err := cli.get("/version", nil)
	if err != nil {
		return nil, err
	}

	version := &types.SystemVersion{}
	err = decodeBody(version, resp.Body)
	ensureCloseReader(resp)

	return version, err
}
```

## Conclusion

Swagger helps pouch to use a general way to design API among lots of different committers. A standard way can improve collaborating efficiency a lot.
