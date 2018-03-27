# Cross Cluster Controller

## Background
Kubernetes clusters using the [Amazon VPC CNI plugin](https://github.com/aws/amazon-vpc-cni-k8s) allow pods to be assigned and reachable on an IP address within the AWS VPC CIDR range. This means that two K8 clusters spun up in different VPCs that are peered to each other with security group policies to allow for ingress between the two clusters are able to communicate with pods from the other cluster via their IPs (so long as their VPC CIDRs don't clobber). 

Previously, our K8 architecture setup cross-cluster traffic through AWS's ELBs. This involves creating ELBs in both clusters, with a service on one side pointing at the DNS of the ELB of the service on the other side in order to actually reach the services in the other cluster. This solution is unfortunately manual and cost ineffective. This cross cluster controller automates the process of creating "pointer services" and removes the need for ELBs.

## Requirements to Run
- 2 K8 clusters that are peered to each other with security groups to allow cross-cluster traffic between the nodes
- K8 clusters running Amazon VPC CNI plugin
- K8 service account tokens scoped to allow listing/getting services and endpoints

## Architecture
A cross cluster controller needs to run on both clusters with a service token from the other cluster. The cross cluster controller watches for put and delete events on K8 services on the other cluster.

The cross cluster controller will only track services with the label `fair.com/cross-cluster=true`. This will then grab the endpoints associated with the service and create a service in its own cluster with the same set of endpoints (since those pod IPs are reachable on both sides). On service deletions, it will delete the service on its own end as well.

Example steps:
- Cross cluster controller A is running on Cluster A. Cross cluster controller B is running on Cluster B.
- Service Foo comes up on Cluster A with the fair.com/cross-cluster=true label.
- Cross cluster controller B will create a Service Foo on Cluster B with the same set of endpoints as Service A.
- Service Foo is deleted from Cluster B.
- Cross cluster controller B will delete Service Foo in Cluster B.

## Running Locally
The controller can run in development mode, which will run using the default kubeconfig file ($HOME/.kube/config). This flag can be set by setting the DEV_MODE var to true or by passing in the flag.

```
export DEV_MODE=true
go run main.go

# OR
go run main.go --devmode=true
```
