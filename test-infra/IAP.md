# IAP for Airflow

Instructions for setting up IAP for our Airflow instances used for testing/releasing.

Create a global static IP.

```
PROJECT=mlkube-testing
NAME=airflow
gcloud compute --project=$PROJECT addresses create $NAME --global
```

Create self-signed certificates.

```
PROJECT=mlkube-testing
ENDPOINT_URL="airflow-testing.kubeflow.io"

openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -subj "/CN=${ENDPOINT_URL}/Google/C=US" \
  -keyout ./tls.key -out ./tls.crt
```

Create the secrets in the cluster

```
kubectl create secret generic airflow-ingress-ssl \
  --from-file=./tls.crt --from-file=./tls.key
```