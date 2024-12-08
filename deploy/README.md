## Deploy en k8s con Helm
Tener ya instalados docker, minikube, kubectl y helm \
**kubectl**: https://kubernetes.io/docs/tasks/tools/
**helm**: `sudo snap install helm --classic` (https://helm.sh/docs/intro/install/)
**helmfile**: https://github.com/helmfile/helmfile

Los charts de helm se pueden conseguir en: https://artifacthub.io/

```
# Inicialmente...
helmfile init # Poner 'y' a todo. Este lo corres una vez en la vida nomas

# Install Traefik Resource Definitions:
kubectl apply -f https://raw.githubusercontent.com/traefik/traefik/v2.10/docs/content/reference/dynamic-configuration/kubernetes-crd-definition-v1.yml

echo "<token>" | docker login -u <user> --password-stdin registry.gitlab.com
export DOCKER_CONFIG_B64=$(cat ~/.docker/config.json | base64)

#Install Crds
kubectl apply -f https://raw.githubusercontent.com/Altinity/clickhouse-operator/master/deploy/operator/clickhouse-operator-install-bundle.yaml

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

# Para levantar el traefik
helm install -f traefik.yaml default traefik/traefik

# Para agregar el host a /etc/hosts
echo "$(minikube ip) workflow-manager.local" | sudo tee -a /etc/hosts
echo "$(minikube ip) dashboard.local" | sudo tee -a /etc/hosts


# El endpoint a usar es workflow-manager.local:32080, por lo menos para uso local
# El endpoint de dashboard es dashboard.local:32080/dashboard para visualizar

```