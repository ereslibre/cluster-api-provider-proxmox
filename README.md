# cluster-api-provider-proxmox

This repository contains a `cluster-api` provider for [`Proxmox VE`](https://www.proxmox.com/en/proxmox-ve).

Please, check the [`cluster-api` project](https://github.com/kubernetes-sigs/cluster-api).

## Prerequisites

### Patched `Proxmox VE`

This project leverages `Proxmox VE` snippets to pass arbitrary
`cloud-init` contents to machines upon booting. While `Proxmox VE`
allows you to create `snippets` either by SSH'ing into the hypervisor,
or by using the UI, none of this solutions is optimal for creating
VM's: SSH'ing into the hypervisors is not desired, and snippet
creation through the UI is not programmatic.

Because of this reason, [a small patch needs to be applied to your
`Proxmox VE` installation](https://bugzilla.proxmox.com/show_bug.cgi?id=2208).

**Without this patch, this provider won't work**. If `Proxmox`
provides an official way of uploading snippets using the API, that
solution will be used instead of this patch.

### Create a VM template

Machines are created by cloning a template, and then booting the
cloned VM with some specific custom `cloud-init` scripts.

Create a VM template that will contain the minimum requirements for
deploying Kubernetes. While this is not strictly required, it's a good
idea so new machine creation and join will be much faster and less
exposed to temporary errors when installing packages on first boot.

You can follow this instructions:

* [Installing kubeadm toolbox](https://kubernetes.io/docs/setup/production-environment/tools/kubeadm/install-kubeadm/)
* [Creating a cloud-init template on `Proxmox VE`](https://pve.proxmox.com/wiki/Cloud-Init_Support#_preparing_cloud_init_templates)

### Environment variables

The manager needs to be started with certain environment variables set:

* `PROXMOX_HOSTPORT`
  * Example: `a-proxmox.some-company.intra.net:8006`
* `PROXMOX_USERNAME`
  * Example: `root@pam`
* `PROXMOX_PASSWORD`
  * Example: `mypassword`
* `PROXMOX_HYPERVISOR_NAME`
  * Example: `proxmox`
* `PROXMOX_HYPERVISOR_SNIPPETS_STORAGE`
  * Example: `ci-snippets`
  * Important: make sure you create this volume using the Proxmox UI
    with `Snippets` content.
* `VM_TEMPLATE_ID`
  * Example: `9000`
  * Important: make sure you have created this template beforehand.

## `TODO`

All the previous environment variables should be removed as
development on this provider evolves.

- [ ] Support a list of hypervisors
  * Removes the need for `PROXMOX_HOSTPORT`, `PROXMOX_USERNAME`, `PROXMOX_PASSWORD`, `PROXMOX_HYPERVISOR_NAME`, `PROXMOX_HYPERVISOR_SNIPPETS_STORAGE` envvars
- [ ] Support machine classes
  * Removes the need for `VM_TEMPLATE_ID` envvar
  * Makes it possible to have different machine specs instead of
    hardcoded ones
- [ ] Support SSH keys when deploying machines on `cloud-init`
      configuration
- [ ] Do not always ignore certificate errors when talking to the
      `Proxmox VE` api, make it configurable
- [ ] Make kubeadm join token generic on `cloud-init` configuration
- [ ] Support openSUSE Leap
- [ ] Support openSUSE Tumbleweed
- [ ] Investigate what network settings can be tweaked (bridge, vlan tagging...)
- [ ] Test `clusterctl`

## License

```
Copyright 2019 Rafael Fernández López <ereslibre@ereslibre.es>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
```
