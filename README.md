# Hoster Intro

![Hoster Core Logo](https://github.com/yaroslav-gwit/HosterCore/raw/main/screenshots/hoster-core-cropped.png)
Introducing `Hoster` 🚀 - FreeBSD's VM and Jail management framework that will make your life easier.
Whether you're an experienced sysadmin or just starting out, `Hoster` has got you covered.

Built using modern, rock solid and battle tested technologies like Go, FreeBSD, bhyve, ZFS, and PF, `Hoster` is a highly opinionated system that puts an emphasis on ease of use and speed of VM deployments.
Whether you're managing a small home lab or a large-scale production, `Hoster` can easily accommodate your environment 🧑🏼‍💻
![HosterCore CLI](https://github.com/yaroslav-gwit/HosterCore/raw/main/screenshots/hoster-v03.png)

Here are some of the features you'll be able to use:

- PF firewall that works with every VM or Jail individually (you can use the VM and/or Jail names in the PF config directly - no need to explicitly implement VLANs to isolate VMs/Jails from one another)
- Native NAT management support using PF
- Bare-metal cloud-friendly deployment options (tested using Hetzner bare-metal cloud)
- Storage Dataset Encryption - your data is safe in the co-location, or on the bare-metal cloud
- Instant VM deployments - a new VM can be deployed in less than 1 second
- Cloud based VM images to avoid spending time with ISOs - it's very similar to `docker pull`, simply run `hoster image download debian12` and it'll be ready for use in minutes (depending on your internet connection speed)
- CloudInit integration to help you forget about the manual VM configurations - simply deploy the VM, start it and it's immediately ready to be used (with the IP address configured, hostname set, and your own custom scripts executed upon first boot)
- Built-in and easy to use OpenZFS replication (based on OpenZFS's send/receive features), which gives you the ability to perform continuous asynchronous VM replication between the two or more hosts, to ensure the data safety and availability 🛡️
- RestAPI for ease of management, and to support the integration with 3rd party systems, or your own home-grown solutions
- An automated HA failover using the underlying RestAPI, so you can avoid the complex network configurations - it's all based on the HTTP protocol, which is easy to firewall and troubleshoot if there is a need for it
- PCI/GPU passthrough is supported, but considered experimental

To avoid any frustrations, here is the list of things NOT currently supported:

- Custom `bhyve` flags are not supported - I want to make sure every flag introduced has it's own config option
- File systems other than ZFS are not supported
- Linux as a host OS is not supported
- Very niche OSes are not supported due to some `bhyve`+`Hoster` limitations
- Only UEFI booting is officially supported, with just few exceptions for BIOS based Linux VMs
- Terraform is not supported - `Hoster` is too young to have any IaaC integrations at this point
- Custom binary files location and custom config files location is not supported - everything must reside within `/opt/hoster-core` to work properly (there is some WIP to overcome this limitation, but it's not ready yet)
- IPv6 is not supported yet (unless you want to manage it by hand, or sponsor us to speed up the IPv6 dev integration process)
- Nested virtualization is not supported by `bhyve`
- Code is not cross-platform - you can't run it on Illumos or any other BSD system, it only works on FreeBSD (and possibly some FreeBSD derivatives, like GhostBSD, HardenedBSD, etc, but it needs testing)
- `bhyve` doesn't support the live VM migration yet

Coming soon (or at least at some point 😄):

- Hyper Converged setup using `HAST` and `OpenZFS` (sync vs. async replication)
- Generally available WebUI for VM and Jail management
- The ability to specify a `docker-compose.yaml` file location during the VM deployment (`hoster vm deploy --from-docker-compose compose.yaml`) - it will be automatically picked up by `docker` and executed in a background `tmux` session (you can easily check if the `docker` deployment was successful using the `tmux a` command later on, when you SSH into the VM)
- Publicly available `Grafana` dashboards
- `Prometheus` integration - all VMs and Jails will be discovered and monitored automatically
- Automated, full mesh `WireGuard` config generation based on your internal cluster topology, with the support for multi-tenancy and multiple cluster groups
- `Tailscale` installation and configuration guide (specific to `Hoster`), which will provide an easy way to access and manage your cluster and it's resources
- `LazyHoster` (an interface similar to `k9s` or `lazygit`) - a CLI utility to control the whole fleet of machines over the `REST API` (for the WebUI haters 😎)
- More VM and Jail templates
- The ability to run `Hoster` in a Jail - this will allow to split a single hardware server into a multi-user isolated platform, with the ability to apply a highly granular control of just how much resources each user can utilize within their own, jailed, virtualization environment

## Are there any plans to develop a WebUI?

Yes.
Part of the project roadmap includes the development of a WebUI.
The WebUI will serve as a user-friendly interface to interact with the system and control multiple hoster nodes simultaneously.
While currently not the highest priority due to time constraints, I am open to exploring this feature further with increased community engagement and potential investment.

Our paying customers already have access to an early version of the WebUI, that looks like this:
![Hoster Core WebUI 1](https://github.com/yaroslav-gwit/HosterCore/raw/main/screenshots/hoster-web-ui-1.png)
![Hoster Core WebUI 2](https://github.com/yaroslav-gwit/HosterCore/raw/main/screenshots/hoster-web-ui-2.png)
![Hoster Core WebUI 3](https://github.com/yaroslav-gwit/HosterCore/raw/main/screenshots/hoster-web-ui-3.png)
![Hoster Core WebUI 4](https://github.com/yaroslav-gwit/HosterCore/raw/main/screenshots/hoster-web-ui-4.png)
![Hoster Core WebUI 5](https://github.com/yaroslav-gwit/HosterCore/raw/main/screenshots/hoster-web-ui-5.png)
![Hoster Core WebUI 6](https://github.com/yaroslav-gwit/HosterCore/raw/main/screenshots/hoster-web-ui-6.png)

The main idea behind our WebUI is to keep things simple.
We are not aiming to be yet another XenSever/Proxmox feature clone: the WebUI will do basic things like managing and deploying new VMs, displaying monitoring information for the VMs and Hosts, managing VM snapshots, connecting to VNC, etc.
Everything else in terms of configuration and `Hoster` management still happens on the CLI.

## Cheatsheet - VM Status (state) icons

| Icon | Meaning                                    |
| :--: | :----------------------------------------- |
|  🟢  | VM is running                              |
|  🔴  | VM is stopped                              |
|  🔁  | Production VM\*\*                          |
|  🔒  | VM is located on the encrypted ZFS Dataset |
|  💾  | VM is a backup from another node           |

\*\* Only production VMs will be included in the `start-all`, `snapshot-all`, `replicate-all`, etc

### OS Support

#### List of supported OSes

| OS                  |       State       | Notes                                                                                               |
| :------------------ | :---------------: | :-------------------------------------------------------------------------------------------------- |
| Debian 11           |     🟢 Ready      | VM image is ready to be downloaded directly from our public image server                            |
| Debian 12           |     🟢 Ready      | VM image is ready to be downloaded directly from our public image server                            |
| AlmaLinux 8         |     🟢 Ready      | VM image is ready to be downloaded directly from our public image server                            |
| AlmaLinux 9         |     🟢 Ready      | VM image is ready to be downloaded directly from our public image server                            |
| RockyLinux 8        |     🟢 Ready      | VM image is ready to be downloaded directly from our public image server                            |
| RockyLinux 9        |     🟢 Ready      | VM image is ready to be downloaded directly from our public image server                            |
| Ubuntu 20.04        |     🟢 Ready      | VM image is ready to be downloaded directly from our public image server                            |
| Ubuntu 22.04        |     🟢 Ready      | VM image is ready to be downloaded directly from our public image server                            |
| Ubuntu 24.04        |     🟢 Ready      | VM image is ready to be downloaded directly from our public image server                            |
| RHEL 8              |   🟡 Compatible   | VM image will have to be built manually by the end user due to licensing issues                     |
| RHEL 9              |   🟡 Compatible   | VM image will have to be built manually by the end user due to licensing issues                     |
| RouterOS 7          |   🟡 Compatible   | VM image will have to be built manually by the end user due to licensing issues                     |
| Windows 10          |   🟡 Compatible   | VM image will have to be built manually by the end user due to licensing issues                     |
| Windows 11          |   🟡 Compatible   | OS requires tinkering with the registry to disable the TPM checks                                   |
| Windows Server 19   |   🟡 Compatible   | VM image will have to be built manually by the end user due to licensing issues                     |
| Windows Server 22   |   🟡 Compatible   | VM image will have to be built manually by the end user due to licensing issues                     |
| FreeBSD 13 ZFS      | 🔴 Not ready yet  | VM image will be released on our public server at some point, but it's not ready yet                |
| FreeBSD 13 UFS      | 🔴 Not ready yet  | VM image will be released on our public server at some point, but it's not ready yet                |
| Fedora (latest)     | 🔴 Not ready yet  | VM image will be released on our public server at some point, but it's not ready yet                |
| OpenSUSE Leap       | 🔴 Not ready yet  | VM image will be released on our public server at some point, but it's not ready yet                |
| OpenSUSE Tumbleweed | 🔴 Not ready yet  | VM image will be released on our public server at some point, but it's not ready yet                |
| OpenBSD             | 🚫 NOT Compatible | The OS is trying to execute an obscure CPU/Mem instruction and immediately gets terminated by Bhyve |

## Start using Hoster

Whether you need a quick start guidance, or a deeper dive into `Hoster`'s documentation, you can definitely do so by visiting this link:
[Hoster Core Docs](https://docs.hoster-core.gateway-it.com/)

## Stargazers over time

[![Stargazers over time](https://starchart.cc/yaroslav-gwit/HosterCore.svg)](https://starchart.cc/yaroslav-gwit/HosterCore)
