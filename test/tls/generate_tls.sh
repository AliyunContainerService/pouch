#!/bin/bash

set -x
which openssl
if (( $? != 0 )) ; then
	echo "Fail to find openssl tool"
	exit 1
fi

# Generate CA
openssl req -new -newkey rsa:2048 -days 3650 -nodes -x509 -subj "/C=CN/ST=ZheJiang/L=HangZhou/O=Company/OU=Department/CN=pouch_test" -keyout ca-key.pem -out ca.pem

# Generate private key for server
name="server"
mkdir -p ${name}
openssl genrsa -out ${name}/key.pem 2048
# Generate CSR
openssl req -subj "/C=CN/ST=ZheJiang/L=HangZhou/O=Company/OU=Department/CN=${name}" -new -key ${name}/key.pem -out ${name}/$name.csr
# Generate CRT
echo "extendedKeyUsage = serverAuth" >./extfile.out
openssl x509 -req -days 3650 -in ${name}/${name}.csr -CA ca.pem -CAkey ca-key.pem -CAcreateserial -out ${name}/cert.pem -extfile extfile.out
cp ca.pem ${name}/ca.pem

# Client
name=a_client
mkdir -p ${name}
# create a key
openssl genrsa -out ${name}/key.pem 2048
# create a csr
openssl req -subj "/C=CN/ST=ZheJiang/L=HangZhou/O=Company/OU=Department/CN=${name}" -new -key ${name}/key.pem -out ${name}/$name.csr
# generate a certificate
echo "extendedKeyUsage = clientAuth" >./extfile.out
openssl x509 -req -days 3650 -in ${name}/${name}.csr -CA ca.pem -CAkey ca-key.pem -CAcreateserial -out ${name}/cert.pem -extfile ./extfile.out

cp ca.pem ${name}/ca.pem
