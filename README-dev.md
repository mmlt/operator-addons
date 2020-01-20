

# Development

Run controller outside cluster:
1. Install CRD in k8s: `make install`
2. Run manager: `make run`

Run controller in cluster:
1. Create image: `IMG=localhost:32000/operator-addons make docker-build docker-push`
2. Install CRD in k8s: `make install`
3. Deploy manager: `make deploy` 

Run controller in IDE:
1. Generate CRD and run tests: `make test`
2. Apply CRD: `make crd-apply`
3. Run in IDE

Run controller in cluster using `k8s-clusterops` repo:
1. Generate CRD and run tests: `make test`
2. Create image: `IMG=localhost:32000/operator-addons make docker-build docker-push`
3. Copy CRD to tpl/: `cp config/crd/bases/clusterops.mmlt.nl_clusteraddons.yaml tpl/op-addons-crd.yaml` 
4. Deploy CRD and operator to k8s: `tmplt -a cluster/this/01-apps.yaml | kubectl apply -f -`
5. Show pod: `kubectl -n cpe get po -l control-plane=op-addons`
6. Tail log: `kubectl -n cpe logs -l control-plane=op-addons -f`

Apply a CR: `kubectl apply -f config/samples/clusterops.yaml`

Show CR Status and Events: `kubectl describe clusteraddon`

Show the cluster state ConfigMap: `kubectl -n kube-system describe cm clusterops-state`


## Tips
Help on +kubebuilder annotations:
`controller-gen object:headerFile=./hack/boilerplate.go.txt paths=./... -w`

Help on code generator:
`controller-gen object:headerFile=./hack/boilerplate.go.txt paths=./... -h`

Add CRD (`config/crd/bases/clusterops.mmlt.nl_clusteraddons.yaml`) to IDE for easy ClusterAddon CR creation:
- IDEA Settings | Language & Frameworks | Kubernetes CRD files
- VC ??


Operator best practices:
- https://github.com/kubernetes-sigs/controller-runtime/blob/master/FAQ.md
- https://blog.openshift.com/kubernetes-operators-best-practices/

Discussions around best practices:
- [client-go Q&A for KubeCon EU 2018](https://github.com/kubernetes/community/blob/b3349d5b1354df814b67bbdee6890477f3c250cb/events/2018/05-contributor-summit/clientgo-notes.md)


### [WIP] Providing feedback
How to provide feedback from an operator to an user or service.

#### Status fields
Status fields provide an API with the current state of the world.

For example:
status.ok (true, false, unknown) (green, yellow, red, unknown) - does the operator work ok? false/red=needs attention, yellow=still operational but do not make more changes, operator is working on getting it green again. Yellow state is (should be) temporary.

#### Status Conditions
Status conditions provide insights on major operations or dependencies of the operator.
Conditions need more interpretation (than status fields). Usefull when troubleshooting.

#### Events
Events provide feedback on when and how often major operations happen.

#### Log
Log shows in detail what the operator is doing. 

Levels:
0. (info) shows events
1. shows flow within the operator
2. shows outgoing calls (to external resources)
3. shows the responses of 2



## Notes on controller-runtime

MaxConcurrentReconciles = 1
Before increasing the number a couple of things need to be implemented:
1. Serializing access to repo; 
  - Multiple processes can read the files. 
  - Only one process can update the files when no one is reading.
2. Serializing access to a target cluster. 

RateLimiter
Currently the default Ratelimiter is used with retry times between 15mS and 10m
For our puposes times between 1m and 10m would be more appropriate.
See https://github.com/kubernetes-sigs/controller-runtime/issues/631

Use resync period when watching/reconciling external resources.
