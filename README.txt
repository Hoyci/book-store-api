helm repo add kubernetes-dashboard https://kubernetes.github.io/dashboard/
helm repo update
helm install kubernetes-dashboard kubernetes-dashboard/kubernetes-dashboard --namespace kubernetes-dashboard --create-namespace
kubectl apply -f dashboard-admin.yaml
kubectl -n kubernetes-dashboard create token dashboard-admin
kubectl get pods -n kubernetes-dashboard
kubectl port-forward -n kubernetes-dashboard service/kubernetes-dashboard-kong-proxy 8443:443

1 - kind create cluster --name lgtm --config kind-config.yaml --> Criar o cluster usando kind
    kind load docker-image book-store-api:1.0.1 --name lgtm
2 - helm repo add grafana https://grafana.github.io/helm-charts
    helm repo update
    --> Configurar o Helm para buscar os charts da Grafana
3 - helm install loki grafana/loki-stack \
  --set grafana.enabled=false \
  --set promtail.enabled=false \
  --namespace logging --create-namespace
4 - helm install tempo grafana/tempo \
  --namespace tracing --create-namespace
5 - helm install mimir grafana/mimir-distributed \
  --namespace metrics --create-namespace \
  --set auth.enabled=false
6 - helm install grafana grafana/grafana \
  --namespace monitoring --create-namespace \
  --set adminPassword="admin" \
  --set service.type=NodePort
7 -  kubectl get --namespace monitoring services grafana --> Para ver em qual porta o grafana está rodando (está rodando na porta 80)
    NAME      TYPE       CLUSTER-IP     EXTERNAL-IP   PORT(S)        AGE
    grafana   NodePort   10.96.69.222   <none>        80:30865/TCP   6m1s
8 - kubectl port-forward --namespace monitoring service/grafana 3000:80 --> Para expor o grafana para o meu localhost (http://127.0.0.1:3000/login)
9 - kubectl create namespace observability
kubectl apply -f otel-configmap.yaml
kubectl apply -f otel-service.yaml
kubectl apply -f otel-deployment.yaml
10 - kubectl create namespace app
kubectl apply -f app-configmap.yaml
kubectl apply -f app-secrets.yaml
kubectl apply -f app-service.yaml
kubectl apply -f app-deployment.yaml


alertmanager -> http://mimir-nginx.metrics.svc:80/alertmanager (colocar como prometheus)
loki -> http://loki.logging.svc.cluster.local:3100
temppo -> http://tempo.tracing.svc.cluster.local:3100