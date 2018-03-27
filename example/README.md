# Examples
These are quick example files to test out the functionality of the cross cluster controller.

First, ensure that no nginx service or deployments exist in the cluster you're deploying to. Then, deploy the two Kubernetes objects to the cluster considered "remote".

```
# Create example deploy
kubectl create -f example-deployment.yaml

# Create example service
kubectl create -f example-service.yaml
```

If you're running the cross-cluster controller locally (or in the cluster), you should see the cluster that's considered the "local" side pick up the change and attempt to create a service and endpoint.
