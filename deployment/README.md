# Deploying with Secret File
In order for the controller to run and access a remote cluster, it needs to be able to mount a secret file with the remote kubeconfig settings. 

For clarity/easy reading, we're going to assume we have Cluster A and Cluster B.

First, apply the serviceaccount.yaml to both Cluster A and Cluster B (NOTE: this will create two service accounts per cluster, 1 for local, 1 for remote "read only" access). 

`kubectl create -f serviceaccount.yaml`

Then, for Cluster A grab the "read only" serviceaccount token in Cluster B, base64 decode it, and add it to the kubeconfig_template.yaml. Then create the secret. Do the same for Cluster B using the Cluster A "read only" serviceaccount token.

`kubectl create secret generic cross-cluster-controller --from-file=./kubeconfig_template.yaml`
