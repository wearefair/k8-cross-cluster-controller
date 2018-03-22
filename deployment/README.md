# Deploying with Secret File
In order for the controller to run and access a remote cluster, it needs to be able to mount a secret file with the remote Kubeconfig settings. 

`kubectl create secret generic cross-cluster-controller --from-file=./kubeconfig_template.yaml`
