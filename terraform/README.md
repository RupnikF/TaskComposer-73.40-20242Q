# Pasos de instalacion
```
# La primera vez que corremos el terraform, esperar a que el nodo master complete el script (se reinicia solo al final)
# Correr sobre nodo master
sudo kubeadm token create --print-join-command

# Produce un output con este resultado kubeadm join <ip> --token <token> --discovery-token-ca-cert-hash <hash>
# Correr este comando con sudo en cada uno de los nodos hijo

# Rebootear las instancias desde la consola de AWS de forma manual

```

# Conectar kubectl
Podemos chequear las IPs de los nodos con `terraform output`
1- Tenemos que modificar el archivo local `~/.kube/config`. Cambiarle el nombre a otra cosa para no sobreescribirlo.
2- Ejecutamos `./connect_kubectl.sh 192 168 0 1` reemplazando por IP publica del master node (separada por espacios)
3- Testear con `kubectl get nodes`
4- Agregar el ip privada o dns name del nodo nfs al .env `"NFS_SERVER"`
5- REINICIAR A MANO DESDE LA CONSOLA DE AWS TODOS LOS NODOS ANTES DE PASAR AL DEPLOY DE HELM