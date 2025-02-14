#!/bin/bash

# Check if exactly 4 arguments are provided
if [ "$#" -ne 4 ]; then
    echo "Error: You must provide exactly 4 numbers as arguments for the master node's IP address."
    exit 1
fi

# Validate that each argument is a number between 0 and 255
for arg in "$@"; do
    if ! [[ "$arg" =~ ^[0-9]+$ ]] || [ "$arg" -lt 0 ] || [ "$arg" -gt 255 ]; then
        echo "Error: Each argument must be a number between 0 and 255 to form a valid master node's IP address."
        exit 1
    fi
done

# Construct the IP address
ip_address="$1.$2.$3.$4"
ip_address_dashed="$1-$2-$3-$4"

# Output the IP address
echo "Master node's IP address: $ip_address"
echo "Copying the kubeconfig file to the local machine..."
#scp ubuntu@ec2-3-91-191-2.compute-1.amazonaws.com:~/.kube/config ~/.kube/config
scp ubuntu@ec2-$ip_address_dashed.compute-1.amazonaws.com:~/.kube/config ~/.kube/config

echo "Modifying the kubeconfig file to use the master node's IP address..."
# delete line 4
sed -i '4d' ~/.kube/config
# replace the server address with the master node's IP address and disable tls verification
sed -i "s/server: https:\/\/.*:6443/server: https:\/\/$ip_address:6443\n    insecure-skip-tls-verify: true/g" ~/.kube/config