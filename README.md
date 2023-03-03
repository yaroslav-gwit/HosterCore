# General Information
Introducing `Hoster` 🚀 - VM management framework that will make your life easier. Whether you're an experienced sysadmin or just starting out, `Hoster` has got you covered. With VM network isolation at the bridge level, ZFS dataset encryption that applies to all underlying VMs, and instant VM deployments, `Hoster` is designed to help you get your work done quickly and efficiently. And that's not all - ZFS send/receive also gives Hoster the ability to offer very reliable, asynchronous storage replication between 2 or more hosts, ensuring data safety and availability 🛡️</br>

Built using modern, rock solid and battle tested technologies like Golang, FreeBSD, bhyve, ZFS, and PF, `Hoster` is a highly opinionated framework that puts an emphasis on ease of use and speed of deployment. Say goodbye to the headaches of traditional VM management and hello to the world of simplicity and reliability. Whether you're managing a small home lab or a large-scale production, `Hoster` can easily accommodate your environment 🧑🏼‍💻

# The why?
![Hoster Core Screenshot](https://github.com/yaroslav-gwit/HosterRed-HyperVisor/blob/main/screenshots/HosterRedScreenshotMain.png)
<br>
Have you ever found yourself frustrated with your current hosting solution? Maybe you're a `Proxmox/XCP-ng/XenServer` user like I was for a long time, but when I started renting Hetzner hardware servers, I ran into some serious roadblocks. High RAM usage on smaller servers, no integrated NAT management, and a nightmare of working with multiple public IPs were just a few of the issues I faced. But then I started to look around for an alternative solution to Linux based virtualization, discovered FreeBSD and bhyve, and everything changed. With PF to control the traffic and native ZFS encryption, I could do so much more, but existing solutions like `vmbhyve` and `CBSD` (which heavily inspired my development) just could'nt cut it for me. That's when I decided to create Hoster: a wholistic system that combines bhyve, PF, and ZFS into a powerful hosting platform that can be deployed on any hardware with minimal RAM overhead.</br></br>
`Hoster` was initially written using Python3 (in 2021), but as the project's codebase grew it became clear that a compiled, statically typed language was necessary to improve the runtime speed (`vm list` in Python3 takes ~500ms to execute, while in Go it's 50-150ms, depending on the number of VMs + your hardware speed) and new node onboarding experience. That's when I made the decision to rewrite everything in Go, resulting in execution speeds that are up to 20 times faster! `Hoster` is now used by several individuals (including myself) as their hosting platform of choice. Give it a try and let me know your thoughts 😉
</br>

# Using Nebula to scale `Hoster` networks
Ever though about connecting all of your `hoster` nodes using some kind of VPN/networking overlay technology to achieve automatic VM network routing across multiple locations/cities/data centers/continents? Nebula has got you covered. One of the key benefits of Nebula is its ease of use and deployment. It can be installed on any Linux, macOS, or Windows machine in a matter of minutes, and it requires minimal configuration. Once installed, it automatically discovers other Nebula nodes on the network and establishes secure, point-to-point connections with them. This means that you don't need to worry about complex network topologies, routing protocols, or VPN configurations - Nebula takes care of everything for you. It's really magical.

Another advantage of Nebula is its scalability. It can efficiently handle thousands of nodes, making it ideal for large-scale deployments. I've built an automated REST API server, that allows you to join your `Hoster` nodes into one big network in a matter of seconds, with point-to-point connections (where possible), failover channels and automatic routing.

# Any plans to make a WebUI?
Yes, I am planning to develop a central REST API that will be able to control 100s of `hoster` nodes at the same time, and a WebUI using VueJS. At the moment it's not a priority due to time constraints on my part, but given enough interest from the community and investors - this task could be achieved in just a matter of few months.

### VM Status (state) icons
🟢 - VM is running
<br>🔴 - VM is stopped
<br>💾 - VM is a backup from another node
<br>🔒 - VM is located on the encrypted ZFS Dataset
<br>🔁 - Production VM icon: VM will be included in the autostart, automatic snapshots/replication, etc

## OS Support
### List of supported OSes
- [x] Debian 11
- [x] AlmaLinux 8
- [x] RockyLinux 8
- [x] Ubuntu 20.04
- [x] Windows 10 (You'll have to provide your own image, instructions on how to build one will be released in the Wiki section soon)

### OSes on the roadmap
- [ ] FreeBSD 13 ZFS
- [ ] FreeBSD 13 UFS
- [ ] Ubuntu 22.04
- [ ] Ubuntu 20.04 LVM Hardened
- [ ] Fedora (latest)
- [ ] CentOS 7
- [ ] OpenBSD
- [ ] OpenSUSE Leap
- [ ] OpenSUSE Tumbleweed
- [ ] Windows 11
- [ ] Windows Server (latest)

### OSes not on the roadmap currently
- [ ] ~~MacOS (any release)~~

# Quick start section
## Installation
Login as root and install `bash`, `curl` and `tmux`
```
sudo su -
pkg update && pkg install -y bash curl tmux
```

This step is optional but highly recommended. Essentially, if you ignore to set any of these values they will be generated automatically. Specifically look at the network port and ZFS encryption password:
```
export DEF_NETWORK_NAME=internal
export DEF_NETWORK_BR_ADDR=10.0.103.254
export DEF_NETWORK_SUBNET=10.0.103.0/24
export DEF_NETWORK_RANGE_START=10.0.103.10
export DEF_NETWORK_RANGE_END=10.0.103.200
export DEF_PUBLIC_INTERFACE=bge0
export DEF_ZFS_ENCRYPTION_PASSWORD="SuperSecretRandom_password"
```

Run the installation script:
```
curl -S https://raw.githubusercontent.com/yaroslav-gwit/HosterRed-HyperVisor/python-branch-main/deploy.sh | bash
```

At the end of the installation you will receive a following message:
```
╭────────────────────────────────────────────────────────────────────────────╮
│                                                                            │
│  The installation is now finished.                                         │
│  Your ZFS encryption password: it's right below this box                   │
│                                                                            │
│  Please save your password! If you lose it, your VMs on the encrypted      │
│  dataset will be lost!                                                     │
│                                                                            │
│  Reboot the system now to apply changes.                                   │
│                                                                            │
│  After the reboot mount the encrypted ZFS dataset and initialize Hoster    │
│  (these 2 steps are required after each reboot):                           │
│                                                                            │
│  zfs mount -a -l                                                           │
│  hoster init                                                               │
│                                                                            │
╰────────────────────────────────────────────────────────────────────────────╯
SuperSecretRandom_password
```
At this point take a minute and save the ZFS encryption password, otherwise you'll lose access to the encrypted dataset!<br>
Now reboot the system, and once the system is back online run `hoster init` to load any missing kernel modules or services:
```
hoster init
```
> `hoster init` has to be executed after every reboot

<br>

Mount your encrypted ZFS dataset:
```
zfs mount -a -l
```

Download your first Linux based image to start the virtualization journey with `Hoster`:
```
hoster image download -t debian11
```

And now you can finally deploy your VM:
```
hoster vm deploy -n newAwesomeVmName -c 1 -r 1G --start-now
# or with long flags
hoster vm deploy --name newAwesomeVmName --cpu-cores 2 --ram 2G --start-now
```

## Backups
### Scheduled automatic snapshots for all production VMs
```
#== AUTOMATIC SNAPSHOTS ==#
33 * *  * *      root  hoster vm snapshot-all  --stype  hourly   --keep 3
5  4 *  * *      root  hoster vm snapshot-all  --stype  daily    --keep 5
5  5 *  * 3      root  hoster vm snapshot-all  --stype  weekly   --keep 3
5  3 15 * *      root  hoster vm snapshot-all  --stype  monthly  --keep 6
```

### Scheduled automatic replication for all production VMs
```
#== AUTOMATIC REPLICATION ==#
50 * *  * *      root  hoster vm replicate-all --endpoint 192.168.1.11
```
