## Deploy en k8s con Helm
Tener ya instalados docker, minikube, kubectl y helm \
**kubectl**: https://kubernetes.io/docs/tasks/tools/
**minikube**: https://minikube.sigs.k8s.io/docs/start/
**helm**: `sudo snap install helm --classic` (https://helm.sh/docs/intro/install/)
**helmfile**: https://github.com/helmfile/helmfile

Los charts de helm se pueden conseguir en: https://artifacthub.io/

```
minikube start
helmfile init # Poner 'y' a todo
helmfile apply
helmfile destroy
```

