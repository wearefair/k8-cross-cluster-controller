# Deploying with Secret File
In order for the controller to run and access a remote cluster, it needs to be able to mount a secret file with the remote kubeconfig settings. 

First, apply the serviceaccount.yaml to both clusters (NOTE: this will create two service accounts per cluster, 1 for local, 1 for remote "read only" access). 

`kubectl create -f serviceaccount.yaml`

Then, grab the token associated with the serviceaccount created on the remote cluster, base64 decode it, and add it to the kubeconfig_template.yaml. Then create the secret on the local cluster side.

`kubectl create secret generic cross-cluster-controller --from-file=./kubeconfig_template.yaml`
