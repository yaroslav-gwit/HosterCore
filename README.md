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
Have you considered utilizing Nebula, a powerful VPN and networking overlay technology, to achieve seamless VM network routing across diverse locations such as cities, data centers, and even continents? Nebula presents an excellent solution for this purpose. It boasts numerous benefits, foremost among them being its simplicity and ease of deployment. In a matter of minutes, Nebula can be installed on Linux, macOS, or Windows machines with minimal configuration requirements. Upon installation, Nebula automatically discovers other nodes within the network, establishing secure, point-to-point connections effortlessly. Consequently, complex network topologies, routing protocols, and VPN configurations become a thing of the past as Nebula efficiently handles them on your behalf. It truly works like magic.

Another compelling feature of Nebula lies in its scalability. With the ability to handle thousands of nodes, Nebula proves to be an ideal choice for large-scale deployments. To further enhance this capability, I have developed an automated REST API server that streamlines the process of joining your Hoster nodes into a unified network. In just a matter of seconds, you can establish point-to-point connections wherever possible, create failover channels, and enable automatic routing. This ensures seamless integration and efficient management of your Hoster infrastructure on a grand scale.

By leveraging Nebula's capabilities, you can achieve robust and scalable networking for your Hoster nodes, making it easier than ever to manage and control your infrastructure across diverse locations.

## Are there any plans to develop a WebUI?
Yes, part of the project roadmap includes the development of a WebUI using VueJS. The WebUI will serve as a user-friendly interface to interact with the system and control multiple hoster nodes simultaneously. While currently not the highest priority due to time constraints, I am open to exploring this feature further with increased community engagement and potential investment. With sufficient interest and support, the development of the WebUI could be accelerated and completed within a few months.

#### VM Status (state) icons
🟢 - VM is running
<br>🔴 - VM is stopped
<br>💾 - VM is a backup from another node
<br>🔒 - VM is located on the encrypted ZFS Dataset
<br>🔁 - Production VM icon: VM will be included in the autostart, automatic snapshots/replication, etc

### OS Support
#### List of supported OSes
- [x] Debian 11
- [x] AlmaLinux 8
- [x] RockyLinux 8
- [x] Ubuntu 20.04
- [x] Ubuntu 22.04
- [x] Windows 10 (You'll have to provide your own image, instructions on how to build one will be released in the Wiki section soon)
- [x] Windows Server 19 (same as Windows 10)

#### OSes on the roadmap
- [ ] FreeBSD 13 ZFS
- [ ] FreeBSD 13 UFS
- [ ] Fedora (latest)
- [ ] OpenBSD
- [ ] OpenSUSE Leap
- [ ] OpenSUSE Tumbleweed
- [ ] Windows 11

#### OSes not on the roadmap currently
- [ ] ~~MacOS (any release)~~

## Start using Hoster

Whether you need a quick start guidance, or a deeper dive into `Hoster`'s documentation, you can definitely do so by visiting this link:
[Hoster Core Docs](https://docs.hoster-core.gateway-it.com/)

## Stargazers over time

[![Stargazers over time](https://starchart.cc/yaroslav-gwit/HosterCore.svg)](https://starchart.cc/yaroslav-gwit/HosterCore)
