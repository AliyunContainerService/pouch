#!/bin/bash

set -x
which openssl
if (( $? != 0 )) ; then
	echo "Fail to find openssl tool"
	exit 1
fi

# Notes: To generate a x509 key pair for testing,
# please add a config 'subjectAltName = IP:127.0.0.1'
# under the '[ v3_ca ]' entry in /etc/pki/tls/openssl.cnf.
#
# SSL needs identification of the peer, otherwise your connection
# might be against a man-in-the-middle. Usually the target is given
# as a hostname and this is checked against the subject and subject
# alternative names of the certificate. In our test case, the server
# target is an IP. That's why we do this trick.

# Generate CA
openssl req -new -newkey rsa:2048 -days 3650 -nodes -x509 -subj "/C=CN/ST=ZheJiang/L=HangZhou/O=Company/OU=Department" -keyout ca-key.pem -out ca.pem
