# Apis

Directory /apis currently contains the following three kinds of things:

* API definitions via [swagger.yml](swagger.yml);
* struct in `/apis/types` used in restful API between Pouch Client and Server, both auto generated and manual added;
* api server implemented for pouch daemon.

## Generated and Manual Types

Pouch has both generated types via tool [swagger](https://swagger.io) and mamually added types. Currently both generated and manually added are located in directory `/apis/types`.

## Hack APIs

If you wish to hack Pouch on API side, here we have several points to guide you:

* Design or update API in swagger.yml first. Please don't do direct change to source code;
* Generate API structs using with script file [generate-swagger-models.sh](../hack/generate-swagger-models.sh);
* Start to design or update source code on the basis of swagger.yml.
