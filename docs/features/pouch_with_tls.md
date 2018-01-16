# Protect the pouchd using tls

If clients merely connect to pouchd from the same sever, only unix socket should be used, which means starting pouchd with argument `-l unix:///var/run/pouchd.sock`. If pouchd will be connected via http client on a remote server, pouchd should listen on a tcp port eg:`-l 0.0.0.0:4243`. In this case, if we don't put on a tls protection, pouchd will accept any connection regardless of the remote identity, which is absolutely not safe in production environment.

In order to verify client identity, a CA should be created to generate certificates for pouch daemon and clients.

## create a CA

Create a CA with common name pouch_test, as following

```shell
openssl req -new -newkey rsa:2048 -days 3650 -nodes -x509 -subj "/C=CN/ST=ZheJiang/L=HangZhou/O=Company/OU=Department/CN=pouch_test" -keyout ca-key.pem -out ca.pem
```

After applying above commands, in current directory there are two files named ca.pem and ca-key.pem, they are the CA, and keep them secret.

## create a certificate for daemon

Create a certificate for pouchd, which only can be used as a server certificate, variable name is HOSTNAME of the machine to run pouchd on.


```shell
name=$HOSTNAME
mkdir -p ${name}
# create a key
/usr/bin/openssl genrsa -out ${name}/key.pem 2048
# create a csr
/usr/bin/openssl req -subj "/C=CN/ST=ZheJiang/L=HangZhou/O=Company/OU=Department/CN=${name}" -new -key ${name}/key.pem -out ${name}/$name.csr
# generate a certificate
/usr/bin/openssl x509 -req -days 3650 -in ${name}/${name}.csr -CA ca.pem -CAkey ca-key.pem -CAcreateserial -out ${name}/cert.pem -extfile <(echo "extendedKeyUsage = serverAuth")
cp ca.pem ${name}/ca.pem
```

After applying above commands, we have a directory which contains all files we need to setup pouchd's tls protection, to set it up start pouchd with flags like (replace the variable with correct value) :

```shell
--tlsverify --tlscacert=${name}/ca.pem --tlscert=${name}/cert.pem --tlskey=${name}/key.pem
```

When pouchd started with all theses tls flags, the tcp address which it listens on can only be connected by client which use certificate published by the same CA.

## create a client certificate

```shell
name=a_client
mkdir -p ${name}
# create a key
/usr/bin/openssl genrsa -out ${name}/key.pem 2048
# create a csr
/usr/bin/openssl req -subj "/C=CN/ST=ZheJiang/L=HangZhou/O=Company/OU=Department/CN=${name}" -new -key ${name}/key.pem -out ${name}/$name.csr
# generate a certificate
/usr/bin/openssl x509 -req -days 3650 -in ${name}/${name}.csr -CA ca.pem -CAkey ca-key.pem -CAcreateserial -out ${name}/cert.pem -extfile <(echo "extendedKeyUsage = clientAuth")
cp ca.pem ${name}/ca.pem
```

After applying above commands, we have a directory which contains all files we need to setup pouchd's tls protection, to set it up start pouchd with flags like (replace the variable *name* with correct value) :

```shell
--tlsverify --tlscacert=${name}/ca.pem --tlscert=${name}/cert.pem --tlskey=${name}/key.pem
```

Then we have a directory which contains all files we needed to use with pouch client. eg: use this certificate as an identity to get pouchd service version:

```
./pouch -H ${server_hostname}:4243 --tlsverify --tlscacert=${path}/ca.pem --tlscert=${path}/cert.pem --tlskey=${path}/key.pem version
```

When a client without a certificate or with a certificate not published by the same CA tries to connect to a pouch daemon having a TLS protection, this connection will be refused.