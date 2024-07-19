# TODO generate a key

# generate crt
openssl req -x509 -new -nodes -key cakey.key -sha256 -days 825 -out ca.crt -subj "/C=US/ST=/L=San Francisco, street=Golden Gate Bridge, postalCode=94016/O=acme CA/"

# TODO Combine key and crt into a cacert.pem

openssl req -new -key casignedserverkey.pem -out casignedservercert.csr -subj "/C=US/L=San Francisco, street=Golden Gate Bridge, postalCode=94016/O=acme badgers ltd./"
openssl x509 -req -in casignedservercert.csr  -CA cacert.pem  -out casignedservercert.pem -days 825 -sha256 -extfile server-v3.ext

openssl req -new -key casignedclientkey.pem -out casignedclientcert.csr -subj "/C=US/L=San Francisco, street=Golden Gate Bridge, postalCode=94016/O=acme squirrels ltd./"
openssl x509 -req -in casignedclientcert.csr  -CA cacert.pem  -out casignedclientcert.pem -days 825 -sha256 -extfile server-v3.ext



openssl genpkey -algorithm RSA -out selfsignedserverkey.pem -pkeyopt rsa_keygen_bits:4096
openssl genpkey -algorithm RSA -out selfsignedclientkey.pem -pkeyopt rsa_keygen_bits:4096
openssl genpkey -algorithm RSA -out selfsignedclientkey2.pem -pkeyopt rsa_keygen_bits:4096

openssl req -new -x509 -key selfsignedserverkey.pem -out selfsignedservercert.pem -config ../../cfg/certs/openssl.cnf -days 825 -subj "/C=US/L=San Francisco, street=Golden Gate Bridge, postalCode=94016/O=acme antelopes ltd./"

openssl req -new -x509 -key selfsignedclientkey.pem -out selfsignedclientcert.pem -config ../../cfg/certs/openssl.cnf -days 825 -subj "/C=US/L=San Francisco, street=Golden Gate Bridge, postalCode=94016/O=acme aardvarks ltd./"

openssl req -new -x509 -key selfsignedclientkey2.pem -out selfsignedclientcert2.pem -config ../../cfg/certs/openssl.cnf -days 825 -subj "/C=US/L=San Francisco, street=Golden Gate Bridge, postalCode=94016/O=acme squirrels ltd./"

################################

openssl req -new -x509 -key cli/testdata/casignedserverkey.pem -days 825 -out servercert2.pem -config cfg/certs/openssl.cnf -subj "/C=US/L=San Francisco, street=Golden Gate Bridge, postalCode=94016/O=acme badgers ltd./"\n
mv servercert2.pem servercert.pem
for i in admin/testdata/servercert.pem api/testdata/servercert.pem integration/testdata/servercert.pem remoting/testdata/servercert.pem shutdown/testdata/servercert.pem tekclient/testdata/servercert.pem; do cp servercert.pem $i; done
