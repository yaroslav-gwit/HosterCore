# Hoster - virtualization made easy

## General Information

![Hoster Core Logo](https://github.com/yaroslav-gwit/HosterCore/raw/main/screenshots/hoster-core-cropped.png)
Introducing `Hoster` 🚀 - VM management framework that will make your life easier. Whether you're an experienced sysadmin or just starting out, `Hoster` has got you covered. With the firewall that affects every VM individually (without the need for VLANs), ZFS dataset level encryption (so your data is safe in co-location), and instant VM deployments (a new VM can be deployed in less than 1 second), `Hoster` is designed to help you get your work done quickly and efficiently. And that's not all - built-in and easy to use replication (based on ZFS send/receive) also gives Hoster the ability to offer very reliable, asynchronous VM replication between 2 or more hosts, ensuring data safety and high availability 🛡️</br>

Built using modern, rock solid and battle tested technologies like Golang, FreeBSD, bhyve, ZFS, and PF, `Hoster` is a highly opinionated framework that puts an emphasis on ease of use and speed of VM deployment. Whether you're managing a small home lab or a large-scale production, `Hoster` can easily accommodate your environment 🧑🏼‍💻

## The why?

![Hoster Core Screenshot](https://github.com/yaroslav-gwit/HosterCore/raw/main/screenshots/hoster-core-main.png)
<br>
My entire perspective on virtualization completely changed when I stumbled upon FreeBSD and bhyve. The potential of combining FreeBSD, bhyve, pf, and ZFS became abundantly clear to me. However, as I explored existing solutions like vmbhyve and CBSD, I couldn't help but feel that they didn't quite match up to my expectations. It was this realization that inspired me to embark on a journey to create Hoster — a platform that seamlessly integrates bhyve, PF, and ZFS into a powerful virtualization solution. You can effortlessly deploy Hoster on any hardware, keeping RAM and CPU usage to a minimum. Give it a try and let me know your thoughts. Your input fuels the continuous development and improvement of Hoster.
</br>

## Leveraging Nebula for Scalable Hoster Networks

Have you considered utilizing Nebula, a powerful VPN and networking overlay technology, to achieve seamless VM network routing across diverse locations such as cities, data centers, and even continents? Nebula presents an excellent solution for this purpose. It boasts numerous benefits, foremost among them being its simplicity and ease of deployment. Upon installation, Nebula automatically discovers other nodes within the network, establishing secure, point-to-point connections effortlessly. Consequently, complex network topologies, routing protocols, and VPN configurations become a thing of the past as Nebula efficiently handles them on your behalf. It truly works like magic.

Another compelling feature of Nebula lies in its scalability. With the ability to handle thousands of nodes, Nebula proves to be an ideal choice for large-scale deployments. To further enhance this capability, I have developed a separate REST API server that streamlines the process of joining your `Hoster` nodes into a unified network. In just a matter of seconds, you can establish point-to-point connections wherever possible, create failover channels, and enable automatic internal VM networks routing.

By coupling Nebula's and PF's capabilities, you can achieve robust and scalable networking for your `Hoster` nodes, making it easier than ever to manage and control your infrastructure across diverse locations.

## Are there any plans to develop a WebUI?

Yes, part of the project roadmap includes the development of a WebUI. The WebUI will serve as a user-friendly interface to interact with the system and control multiple hoster nodes simultaneously. While currently not the highest priority due to time constraints, I am open to exploring this feature further with increased community engagement and potential investment.

Our payed customers already have access to an early version of the WebUI, that looks like this:
![Hoster Core WebUI 1](https://github.com/yaroslav-gwit/HosterCore/raw/main/screenshots/hoster-web-ui-1.png)
![Hoster Core WebUI 2](https://github.com/yaroslav-gwit/HosterCore/raw/main/screenshots/hoster-web-ui-2.png)
![Hoster Core WebUI 3](https://github.com/yaroslav-gwit/HosterCore/raw/main/screenshots/hoster-web-ui-3.png)
![Hoster Core WebUI 4](https://github.com/yaroslav-gwit/HosterCore/raw/main/screenshots/hoster-web-ui-4.png)
![Hoster Core WebUI 5](https://github.com/yaroslav-gwit/HosterCore/raw/main/screenshots/hoster-web-ui-5.png)
![Hoster Core WebUI 6](https://github.com/yaroslav-gwit/HosterCore/raw/main/screenshots/hoster-web-ui-6.png)

The main idea behind our WebUI is to keep things simple. We are not aiming to be yet another XenSever/Proxmox feature clone: the WebUI will do basic things like managing and deploying new VMs, displaying monitoring information for the VMs and Hosts, managing VM snapshots, connecting to VNC, etc. Everything else in terms of configuration and `Hoster` management still happens on the CLI.

#### VM Status (state) icons

| Icon  | Meaning                                    |
| :--:  | :--                                        |
| 🟢    | VM is running                              |
| 🔴    | VM is stopped                              |
| 🔁    | Production VM**                            |
| 🔒    | VM is located on the encrypted ZFS Dataset |
| 💾    | VM is a backup from another node           |
 

** Only production VMs will be included in the `start-all`, `snapshot-all`, `replicate-all`, etc

### OS Support

#### List of supported OSes

|  OS                 | State             | Notes                                                                                |
| :--                 | :--:              | :--                                                                                  |
| Debian 11           | 🟢 Ready          | VM image is ready to be downloaded directly from our public image server             |
| Debian 12           | 🟢 Ready          | VM image is ready to be downloaded directly from our public image server             |
| AlmaLinux 8         | 🟢 Ready          | VM image is ready to be downloaded directly from our public image server             |
| RockyLinux 8        | 🟢 Ready          | VM image is ready to be downloaded directly from our public image server             |
| Ubuntu 20.04        | 🟢 Ready          | VM image is ready to be downloaded directly from our public image server             |
| Ubuntu 22.04        | 🟢 Ready          | VM image is ready to be downloaded directly from our public image server             |
| Windows 10          | 🟡 Compatible     | VM image will have to be built manually by the end user due to licensing issues      |
| Windows 11          | 🔴 NOT Compatible | Waiting for the TMP module to be implemented within Bhyve                            |
| Windows Server 19   | 🟡 Compatible     | VM image will have to be built manually by the end user due to licensing issues      |
| Windows Server 22   | 🟡 Compatible     | VM image will have to be built manually by the end user due to licensing issues      |
| FreeBSD 13 ZFS      | 🔴 Not ready yet  | VM image will be released on our public server at some point, but it's not ready yet |
| FreeBSD 13 UFS      | 🔴 Not ready yet  | VM image will be released on our public server at some point, but it's not ready yet |
| Fedora (latest)     | 🔴 Not ready yet  | VM image will be released on our public server at some point, but it's not ready yet |
| OpenBSD             | 🔴 Not ready yet  | VM image will be released on our public server at some point, but it's not ready yet |
| OpenSUSE Leap       | 🔴 Not ready yet  | VM image will be released on our public server at some point, but it's not ready yet |
| OpenSUSE Tumbleweed | 🔴 Not ready yet  | VM image will be released on our public server at some point, but it's not ready yet |


## Start using Hoster

Whether you need a quick start guidance, or a deeper dive into `Hoster`'s documentation, you can definitely do so by visiting this link:
[Hoster Core Docs](https://docs.hoster-core.gateway-it.com/)

## Stargazers over time

[![Stargazers over time](https://starchart.cc/yaroslav-gwit/HosterCore.svg)](https://starchart.cc/yaroslav-gwit/HosterCore)
