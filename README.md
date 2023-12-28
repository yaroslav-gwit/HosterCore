# Hoster Intro

![Hoster Core Logo](https://github.com/yaroslav-gwit/HosterCore/raw/main/screenshots/hoster-core-cropped.png)
Introducing `Hoster` 🚀 - FreeBSD's VM and Jail management framework that will make your life easier.
Whether you're an experienced sysadmin or just starting out, `Hoster` has got you covered.

Built using modern, rock solid and battle tested technologies like Go, FreeBSD, bhyve, ZFS, and PF, `Hoster` is a highly opinionated system that puts an emphasis on ease of use and speed of VM deployments.
Whether you're managing a small home lab or a large-scale production, `Hoster` can easily accommodate your environment 🧑🏼‍💻

Here are some of the features you'll be able to use:

- PF firewall that works with every VM individually, without the need for VLANs
- Cloud friendly deployment options, with native support for NAT (with Hetzner as a main focus)
- ZFS Dataset Encryption - your data is safe in the co-location, or on the bare-metal cloud
- Instant VM deployments - a new VM can be deployed in less than 1 second
- Cloud based VM images to avoid spending time with ISOs - it's very similar to `docker pull`, simply run `hoster image download debian12` and it'll be ready for use in minutes
- CloudInit integration to help you forget about the manual VM configurations - simply deploy the VM, start it and it's immediately ready to be used, with the IP address configured, hostname set, and your own custom scripts executed upon first boot
- Built-in and easy to use OpenZFS replication (based on OpenZFS's send/receive), which gives you the ability to perform continuous asynchronous VM replication between 2 or more hosts, to ensure data safety and availability 🛡️
- RestAPI for ease of management, and to support the integration with 3rd party systems, or your home-grown solutions
- HA clustering using the underlying RestAPI with an automated VM failover, so you can avoid complex network configurations - it's just an HTTP, which is easy to firewall and troubleshoot if there is a need for it
- (not ready, but coming at some point) Hyper Converged setup using Gluster, iSCSI, and ZFS - you'll be able to cluster together Petabytes of storage, dynamically provision iSCSI targets, and mount them on a specific node of your choice
- (not ready, but coming at some point) Generally available WebUI for VM and Jail management, that supports hundreds of hosts at the same time
- (not ready, but coming at some point) Specify a `docker-compose.yaml` file location during the VM deployment - it will be automatically picked up by `docker` and executed in a background `tmux` session (you can easily check if the `docker` deployment was successful using the `tmux a` command later on, when you SSH into the VM)
- PCI/GPU passthrough is supported, but considered experimental

To avoid any frustrations, here is the list of things NOT currently supported:

- Custom `bhyve` flags - I want to make sure every flag introduced has it's own config option
- File systems other than ZFS
- Support for Linux as a host (this will never happen with HosterCore)
- Very niche OSes are not supported due to some `bhyve`+`Hoster` limitations
- Only UEFI booting is officially supported, with just few exceptions for BIOS based Linux VMs
- Terraform - `Hoster` is too young to have any IaaC integrations at this point
- Custom binary and config location - everything must reside within `/opt/hoster-core` to work properly
- IPv6 - my ISP doesn't fully support it, so it's hard to implement a support and test it (but it can be done relatively easily, if you are willing to help with a money donation to run a Hetzner host, or a long-term hardware access at your location)

## The why?

![Hoster Core Screenshot](https://github.com/yaroslav-gwit/HosterCore/raw/main/screenshots/hoster-core-main.png)
<br>
My entire perspective on virtualization completely changed when I stumbled upon FreeBSD and bhyve.
The potential of combining FreeBSD, bhyve, pf, and ZFS became abundantly clear to me.
However, as I explored existing solutions like `vm-bhyve` and `CBSD`, I couldn't help but feel that they didn't quite match up to my expectations.
It was this realization that inspired me to embark on a journey to create `Hoster` — a platform that seamlessly integrates bhyve, PF, and ZFS into a powerful virtualization solution.
You can effortlessly deploy `Hoster` on any hardware, keeping RAM and CPU usage to a minimum.
Give it a try and let me know your thoughts.
Your input fuels our continuous project improvements.
</br>

## Leveraging modern SD-WAN and VPN technologies for scalable `Hoster` networks

`Hoster` supports a variety of overlay network technologies like ZeroTier, Nebula, WireGuard, IPSec, OpenVPN, etc.
Essentially `Hoster` supports anything FreeBSD supports.
We haven't implemented any tight coupling in terms of networking.
Both, VMs and Jails, are connected to the outside world using the bridge adapters, so as long as your VPN/SD-WAN supports a `bridge` mode you'll be fine.

Check our documentation for the specific instructions on the tech stack you are interested in.

## Are there any plans to develop a WebUI?

Yes, part of the project roadmap includes the development of a WebUI. The WebUI will serve as a user-friendly interface to interact with the system and control multiple hoster nodes simultaneously.
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
| RockyLinux 8        |     🟢 Ready      | VM image is ready to be downloaded directly from our public image server                            |
| Ubuntu 20.04        |     🟢 Ready      | VM image is ready to be downloaded directly from our public image server                            |
| Ubuntu 22.04        |     🟢 Ready      | VM image is ready to be downloaded directly from our public image server                            |
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
