#!/bin/bash

# Install the NFS server
sudo apt update
sudo apt install nfs-kernel-server -y

# Create the directory to be shared and set the permissions
sudo mkdir -p /mnt/nfs
sudo chown -R nobody:nogroup /mnt/nfs
sudo chmod 777 /mnt/nfs

# Update the exports file
echo "/mnt/nfs 172.31.0.0/16 (rw,sync,no_subtree_check,no_root_squash)" | sudo tee -a /etc/exports

# Export the config and restart the NFS server
sudo exportfs -a
sudo systemctl restart nfs-kernel-server

# Disable the firewall
sudo ufw disable
sudo reboot


