#!/bin/bash

# Check if the correct number of arguments are provided
if [ "$#" -ne 12 ]; then
    echo "Usage: $0 <ca> <ca_signed_server> <ca_signed_client> <self_signed_server> <self_signed_client> <self_signed_client2> <admin> <api> <integration> <remoting> <shutdown> <tektclient>"
    exit 1
fi

# Assign the arguments to variables
CA=$1
CA_SIGNED_SERVER=$2
CA_SIGNED_CLIENT=$3
SELF_SIGNED_SERVER=$4
SELF_SIGNED_CLIENT=$5
SELF_SIGNED_CLIENT2=$6
ADMIN=$7
API=$8
INTEGRATION=$9
REMOTING=${10}
SHUTDOWN=${11}
TEKTLIENT=${12}

# Check certs expiration date
for pem in $CA $CA_SIGNED_SERVER $CA_SIGNED_CLIENT $SELF_SIGNED_SERVER $SELF_SIGNED_CLIENT $SELF_SIGNED_CLIENT2 $ADMIN $API $INTEGRATION $REMOTING $SHUTDOWN $TEKTLIENT;
do
   printf '%s: %s\n' \
      "$(openssl x509 -enddate -noout -in "$pem"|cut -d= -f 2)" \
      "$pem"
done | sort