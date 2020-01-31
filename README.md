# AddOns Operator
The addons operator is responsible for making sure a set of kubernetes resources is available in a target cluster.

Its main pupose is to deploy resources like policies, ingresses, RBAC role(binding)s etc. after a new cluster has been
create.

## Usage
Create a resource of kind ClusterAddon and kubectl apply it.

Do a `kubectl get clusteraddon`.
The `Synced` column shows `True` when the operator has successfully applied the action on the target cluster.

For more info do a `kubectl describe clusteraddon`.
The conditions show if the source repo and target cluster are accessible. 
The `ActionOK` condition shows if the last action succeeded.
In addition an Event is recorded for every action performed.

For even more detail check the log, see [README-dev](./README-dev.md)


## How it works
Each environment (test, prod) gets a namespace and an operator to watch CR's in that namespace.
(production and test environments can have different operator versions)

A ClusterAddon CR is created in for each cluster to maintain.

The ClusterAddon CR contains a list of sources with an action to perform.
Typically the source is a GIT repository and the action is a shell command.

The reconciler watches ClusterAddons CR and the repositories for changes.
When something changes it performs the specified action.

### Actions

#### shell
An action of `type: shell` runs a `cmd` when the ClusterAddon CR or the source changes.

When the command runs the `pwd` is a directory that's CR namespace-name specific.
This direcory containts a `values.yaml` file with the `values:` from the ClusterAddon CR.

The environment contains `$SOURCEDIR` with the path to source repository clone.
In addition the environment contains the `values:` from the ClusterAddon CR prefixed with `VALUE_` and converted to uppercase.
For example: `k8sEnvironment: test` results in `VALUE_K8SENVIRONMENT=test`

The `$HOME` of the user that runs the command contains a `.kube/config` that allows access to the target cluster.
 

## CRD
See [clusteraddon_types](api/v1alpha1/clusteraddon_types.go) source or the generated [CRD](config/crd/bases/clusterops.mmlt.nl_clusteraddons.yaml)


## Comparison with similar solutions
### Weave Flux
Runs in the target cluster.