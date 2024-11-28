## Deploy en k8s con Helm
Tener ya instalados docker, minikube, kubectl y helm \
**kubectl**: https://kubernetes.io/docs/tasks/tools/
**minikube**: https://minikube.sigs.k8s.io/docs/start/
**helm**: `sudo snap install helm --classic` (https://helm.sh/docs/intro/install/)
**helmfile**: https://github.com/helmfile/helmfile

Los charts de helm se pueden conseguir en: https://artifacthub.io/

```
# Inicialmente...
minikube start # yo este lo tengo q correr cada vez q prendo la pc
helmfile init # Poner 'y' a todo. Este lo corres una vez en la vida nomas
export $(cat .env | xargs) && helmfile apply # Este es para aplicar los cambios en helmfile.yaml
helmfile destroy # Este es para borrar todo (idem terraform)
```

## Comandos útiles
```
alias k="kubectl -n namespace-name" # Para no escribir 80 veces el -n namespace

helm create my-app-name # Crear un chart (crea una carpeta con el template de archivos)
kubectl delete all --all -n namespace-name (borra TODO lo que exista en el cluster bajo un namespace)

# forwardear un puerto a localhost:
export POD_NAME=$(kubectl get pods --namespace workflow-manager -l "app.kubernetes.io/name=workflow-manager,app.kubernetes.io/instance=workflow-manager" -o jsonpath="{.items[0].metadata.name}")
export CONTAINER_PORT=$(kubectl get pod --namespace workflow-manager $POD_NAME -o jsonpath="{.spec.containers[0].ports[0].containerPort}")
kubectl --namespace workflow-manager port-forward $POD_NAME 8080:$CONTAINER_PORT
```

