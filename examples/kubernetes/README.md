# converge-kubernetes

A converge example that sets up an HA Kubernetes cluster based
on
[kubernetes-the-hard-way](https://github.com/kelseyhightower/kubernetes-the-hard-way).

## Usage

This example is supported on Linux and OSX hosts.

### Vagrant

Just run `vagrant up`!

The first time you run `vagrant up`, converge will be downloaded and executed
from your `/tmp` directory (by default). Converge will generate a certificate
authority on your local machine and upload it to the vagrant instances. From
that point forward, a vagrant provisioner will run converge to configure the
instances and install all of the necessary kubernetes components. By default, a
kubernetes cluster consisting of a single controller and 2 nodes will be
created. You can configure the `CONTROLLER_COUNT` to 3+ nodes to run an HA
cluster. That number of worker nodes can be configured using the `NODE_COUNT`
variable.

### Configuring kubectl

You can SSH into a controller instance (`vagrant ssh controller-1`, for example)
and run kubectl from there. If you want to run kubectl from your host machine,
you can run the following steps from the `examples/kubernetes` directory.

First, you need to get a copy of the CA certificate from one of the vagrant
instances:

```shell
vagrant ssh controller-1 -c "sudo cat /etc/kubernetes/ssl/ca.pem" > ca.pem
```

```shell
kubectl config set-cluster converge-kubernetes \
  --certificate-authority=./ca.pem \
  --embed-certs=true \
  --server=https://172.19.9.21:6443

kubectl config set-credentials converge-kubernetes-admin --token chAng3m3

kubectl config set-context converge-kubernetes \
  --cluster=converge-kubernetes \
  --user=converge-kubernetes-admin

kubectl config use-context converge-kubernetes
```

If you have customized the IP addresses or admin token defined in
the [Vagrantfile](./Vagrantfile), you should change the `--server` and `--token`
options respectively.
