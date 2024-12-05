# Pasos de instalacion
```
# Correr sobre nodo master
sudo kubeadm token create --print-join-command

# Produce un output con este resultado kubeadm join <ip> --token <token> --discovery-token-ca-cert-hash <hash>
# Correr este comando en cada uno de los nodos hijo

```