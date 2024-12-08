# Pasos de instalacion
```
# Correr sobre nodo master
sudo kubeadm token create --print-join-command

# Produce un output con este resultado kubeadm join <ip> --token <token> --discovery-token-ca-cert-hash <hash>
# Correr este comando en cada uno de los nodos hijo

```

# Conectar kubectl
1- Tenemos que modificar el archivo local `~/.kube/config`. Cambiarle el nombre a otra cosa para no sobreescribirlo.
2- Traer el `~/.kube/config` del nodo master por `ssh`: `scp ubuntu@ec2-3-91-191-2.compute-1.amazonaws.com:~/.kube/config ~/.kube/config`
3- Modificar el archivo `~/.kube/config`:
```yaml
# Cambiar la seccion de clusters a lo siguiente
apiVersion: v1
clusters:
- cluster:
    server: https://<ip-publica-del-ec2>:6443 # cambiar la ip acá
    insecure-skip-tls-verify: true # agregar esta linea
    # borrar la linea de client-certificate-data
  name: kubernetes
contexts:
...
```
4- Testear con `kubectl get nodes`