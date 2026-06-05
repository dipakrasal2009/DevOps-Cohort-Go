# 🐧 Linux Namespaces & `unshare` Command

> **The kernel feature that powers every container you've ever run.**

[![Linux](https://img.shields.io/badge/Linux-Kernel-FCC624?style=flat&logo=linux&logoColor=black)](https://kernel.org)
[![Namespaces](https://img.shields.io/badge/Feature-Namespaces-0078D4?style=flat)](#namespace-types)
[![Containers](https://img.shields.io/badge/Related-Containers-2496ED?style=flat&logo=docker&logoColor=white)](#building-a-minimal-container)
[![License](https://img.shields.io/badge/License-MIT-green?style=flat)](LICENSE)

---

## 📖 Table of Contents

- [What is `unshare`?](#what-is-unshare)
- [Synopsis](#synopsis)
- [Namespace Types](#namespace-types)
- [All Flags Reference](#all-flags-reference)
- [Examples by Namespace](#examples-by-namespace)
  - [UTS — Hostname Isolation](#1-uts-namespace--hostname-isolation)
  - [PID — Process Isolation](#2-pid-namespace--process-isolation)
  - [Network — Network Isolation](#3-network-namespace--network-isolation)
  - [Mount — Filesystem Isolation](#4-mount-namespace--filesystem-isolation)
  - [User — User ID Isolation](#5-user-namespace--user-id-isolation)
  - [IPC, Cgroup, Time](#6-ipc-cgroup-and-time-namespaces)
- [Full Isolation Example](#full-isolation-example)
- [Building a Minimal Container](#building-a-minimal-container)
- [How It Works Internally](#how-it-works-internally)
- [Inspecting Namespaces via /proc](#inspecting-namespaces-via-proc)
- [Relation to Docker & Containers](#relation-to-docker--containers)
- [Resources](#resources)

---

## What is `unshare`?

`unshare` is a Linux userspace utility that runs a program in **new, isolated namespaces** that are separate from the calling (parent) process. It is a direct wrapper around the `unshare(2)` Linux system call.

```
parent process  ──[unshare syscall]──►  child process
   (host namespaces)                   (new namespaces)
```

A **namespace** is a kernel abstraction that wraps a global system resource and gives a process its own private view of it. Changes inside the namespace don't affect the host — and vice versa.

> 💡 **Key insight:** Containers (Docker, Podman, LXC) are just processes running inside multiple namespaces with cgroup resource limits applied on top.

---

## Synopsis

```bash
unshare [options] [program [arguments]]
```

If no program is specified, `$SHELL` is launched.

---

## Namespace Types

| Flag | Long Form | Namespace | What It Isolates |
|------|-----------|-----------|-----------------|
| `-m` | `--mount` | **Mount** | Filesystem mount points |
| `-u` | `--uts` | **UTS** | Hostname and domain name |
| `-i` | `--ipc` | **IPC** | System V IPC, POSIX message queues |
| `-n` | `--net` | **Network** | Network devices, ports, routing tables |
| `-p` | `--pid` | **PID** | Process ID numbers |
| `-U` | `--user` | **User** | User IDs, Group IDs |
| `-C` | `--cgroup` | **Cgroup** | Control group root directory |
| `-T` | `--time` | **Time** | Boot and monotonic clock offsets (Linux 5.6+) |

---

## All Flags Reference

| Option | Description |
|--------|-------------|
| `--fork` | Fork a child process (required with `--pid`) |
| `--kill-child[=signame]` | Send signal to child when `unshare` exits |
| `--mount-proc[=mountpoint]` | Mount fresh `/proc` (use with `--pid`) |
| `--map-root-user` / `-r` | Map current user to UID 0 in user namespace |
| `--map-user=uid` | Map a specific UID |
| `--map-group=gid` | Map a specific GID |
| `--map-auto` | Automatically map UIDs/GIDs from `/etc/subuid` |
| `--propagation private\|shared\|slave\|unchanged` | Set mount propagation mode |
| `--setgroups=allow\|deny` | Allow or deny `setgroups()` in user namespace |
| `--keep-caps` | Retain capabilities when mapping to non-root |
| `-r` | Shorthand for `--map-root-user` |

---

## Examples by Namespace

### 1. UTS Namespace — Hostname Isolation

```bash
# Enter a new UTS namespace
sudo unshare --uts bash

# Set a custom hostname — only affects this shell
hostname mycontainer
hostname
# → mycontainer

exit

# Host hostname is completely unchanged
hostname
# → your-real-hostname
```

---

### 2. PID Namespace — Process Isolation

```bash
# Always pair --pid with --fork and --mount-proc
sudo unshare --pid --fork --mount-proc bash

# Only sees its own processes
ps aux
# PID 1 = bash  (this is init inside the namespace!)
# PID 2 = ps

exit
```

> ⚠️ **Important:** Without `--fork`, the shell itself enters the new PID namespace but PID numbering won't reset. Without `--mount-proc`, `ps` still reads the host `/proc`.

---

### 3. Network Namespace — Network Isolation

```bash
# New network namespace — clean slate
sudo unshare --net bash

ip link show
# 1: lo: <LOOPBACK> ...  (only loopback, no eth0/wlan0)

ip route
# (empty — no routes configured)

# You can now set up virtual ethernet (veth) pairs
# to connect this namespace to the outside world
```

---

### 4. Mount Namespace — Filesystem Isolation

```bash
# Private mount namespace
sudo unshare --mount bash

# This mount is invisible on the host
mount -t tmpfs tmpfs /mnt
echo "secret" > /mnt/data.txt

exit

# /mnt is clean on the host — mount never happened there
ls /mnt
# (empty)
```

**Private propagation (recommended):**
```bash
sudo unshare --mount --propagation private bash
```

---

### 5. User Namespace — User ID Isolation

```bash
# No sudo needed for user namespaces!
unshare --user --map-root-user bash

whoami
# → root

id
# → uid=0(root) gid=0(root)

# On the host, kernel maps: UID 0 (inside) = UID 1000 (outside)
cat /proc/self/uid_map
#          0       1000          1
```

**Use `--map-auto` with `/etc/subuid` (Podman-style):**
```bash
unshare --user --map-auto bash
```

---

### 6. IPC, Cgroup, and Time Namespaces

```bash
# IPC isolation
sudo unshare --ipc bash
ipcs   # clean IPC slate

# Cgroup namespace
sudo unshare --cgroup bash

# Time namespace (Linux kernel 5.6+)
sudo unshare --time bash
```

---

## Full Isolation Example

Combine all namespaces for maximum isolation, similar to a real container:

```bash
sudo unshare \
  --pid \
  --mount \
  --net \
  --uts \
  --ipc \
  --fork \
  --mount-proc \
  bash
```

With a hostname set:
```bash
sudo unshare --pid --mount --net --uts --ipc --fork --mount-proc bash -c \
  "hostname mycontainer && exec bash"
```

---

## Building a Minimal Container

You can build a real, working container from scratch using only `unshare` + `chroot`:

```bash
# Step 1: Download a minimal Alpine Linux rootfs
mkdir /tmp/rootfs
curl -L https://dl-cdn.alpinelinux.org/alpine/latest-stable/releases/x86_64/alpine-minirootfs-latest-x86_64.tar.gz \
  | tar -xzC /tmp/rootfs

# Step 2: Enter full namespace isolation
sudo unshare \
  --pid --mount --net --uts --ipc --user \
  --fork --mount-proc \
  --map-root-user \
  bash

# Step 3: Mount the rootfs and chroot into it
mount --bind /tmp/rootfs /tmp/rootfs
cd /tmp/rootfs

# Mount essential pseudo-filesystems
mount -t proc  proc  proc/
mount -t sysfs sysfs sys/
mount -t devtmpfs devtmpfs dev/

# Step 4: Enter the container!
chroot . /bin/sh

# ✅ You are now in a minimal Alpine Linux container
cat /etc/os-release
# NAME="Alpine Linux"
```

---

## How It Works Internally

```
User Space
──────────────────────────────────────────────────────────
  $ sudo unshare --pid --net bash

  unshare utility
       │
       │  calls unshare(2) syscall
       │  flags: CLONE_NEWPID | CLONE_NEWNET
       │
       │  calls fork()
       ▼
  Child Process
       │  attached to NEW pid namespace  (PID 1 = bash)
       │  attached to NEW net namespace  (lo only)
       │  still shares: mount, uts, ipc, user namespaces

Kernel Space
──────────────────────────────────────────────────────────
  Namespace objects (reference-counted kernel structs):

  [pid_ns: 0xffff...]  ◄── new, owned by child
  [net_ns: 0xffff...]  ◄── new, owned by child
  [mnt_ns: 0xffff...]  ◄── shared with parent
```

The `unshare(2)` system call sets `CLONE_NEW*` flags on the calling process, detaching it from shared namespaces and creating fresh ones. Namespaces persist as long as at least one process or file descriptor references them.

---

## Inspecting Namespaces via `/proc`

```bash
# See your current namespace memberships
ls -la /proc/self/ns/
# lrwxrwxrwx cgroup -> cgroup:[4026531835]
# lrwxrwxrwx ipc    -> ipc:[4026531839]
# lrwxrwxrwx mnt    -> mnt:[4026531840]
# lrwxrwxrwx net    -> net:[4026531992]
# lrwxrwxrwx pid    -> pid:[4026531836]
# lrwxrwxrwx time   -> time:[4026531834]
# lrwxrwxrwx user   -> user:[4026531837]
# lrwxrwxrwx uts    -> uts:[4026531838]

# Compare two processes — same inode = same namespace
ls -lai /proc/1/ns/net /proc/$$/ns/net

# Enter an existing process's namespace (requires root)
nsenter --target <PID> --pid --net bash
```

---

## Relation to Docker & Containers

| Feature | `unshare` | Docker |
|---------|-----------|--------|
| PID isolation | `--pid --fork --mount-proc` | ✅ automatic |
| Network isolation | `--net` | ✅ automatic + veth setup |
| Filesystem isolation | `--mount` + `chroot` | ✅ overlayfs layers |
| Hostname isolation | `--uts` | ✅ automatic |
| User isolation | `--user --map-root-user` | ✅ optional (rootless mode) |
| Resource limits | ❌ (use `systemd-run` or `cgcreate`) | ✅ cgroups v2 |
| Image management | ❌ manual | ✅ registry + layers |

Docker, Podman, and containerd are fundamentally orchestration layers on top of exactly these primitives.

---

## Resources

- 📘 [man 1 unshare](https://man7.org/linux/man-pages/man1/unshare.1.html) — Official manpage
- 📘 [man 2 unshare](https://man7.org/linux/man-pages/man2/unshare.2.html) — System call reference
- 📘 [man 7 namespaces](https://man7.org/linux/man-pages/man7/namespaces.7.html) — Namespaces overview
- 📘 [man 7 user_namespaces](https://man7.org/linux/man-pages/man7/user_namespaces.7.html) — User namespace details
- 🔗 [Linux Kernel Namespaces — LWN series](https://lwn.net/Articles/531114/)
- 🔗 [Containers from Scratch — Liz Rice (YouTube)](https://www.youtube.com/watch?v=8fi7uSYlOdc)
- 🔗 [nsenter manpage](https://man7.org/linux/man-pages/man1/nsenter.1.html)

---

## Quick Cheatsheet

```bash
# Hostname isolation
sudo unshare -u bash

# PID isolation (always add --fork --mount-proc)
sudo unshare -p --fork --mount-proc bash

# Network isolation
sudo unshare -n bash

# Mount isolation
sudo unshare -m bash

# Become root without sudo (user namespace)
unshare -Ur bash

# Full container-grade isolation
sudo unshare -m -u -i -n -p -U --fork --mount-proc bash
```

---

<div align="center">

Made with ❤️ for Linux system programming enthusiasts.  
**Star ⭐ this repo if it helped you understand containers better!**

</div>
