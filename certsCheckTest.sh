#!/bin/bash

# Check if the correct number of arguments are provided
if [ "$#" -ne 3 ]; then
    echo "Usage: $0 <ca> <ca_signed_server> <ca_signed_client>"
    exit 1
fi

# Assign the arguments to variables
CA=$1
CA_SIGNED_SERVER=$2
CA_SIGNED_CLIENT=$3

# Check certs expiration date
for pem in $CA $CA_SIGNED_SERVER $CA_SIGNED_CLIENT;
do
   printf '%s: %s\n' \
      "$(openssl x509 -enddate -noout -in "$pem"|cut -d= -f 2)" \
      "$pem"
done | sort