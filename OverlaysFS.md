# 🗂️ OverlayFS — Linux Overlay Mounting Explained

> **How Docker images, container layers, and union filesystems actually work under the hood.**

[![Linux](https://img.shields.io/badge/Linux-Kernel_3.18+-FCC624?style=flat&logo=linux&logoColor=black)](https://kernel.org)
[![OverlayFS](https://img.shields.io/badge/Filesystem-OverlayFS-0078D4?style=flat)](#what-is-overlayfs)
[![Docker](https://img.shields.io/badge/Used_by-Docker-2496ED?style=flat&logo=docker&logoColor=white)](#real-world-docker-uses-overlayfs)

---

## 📖 Table of Contents

- [What is OverlayFS?](#what-is-overlayfs)
- [The Four Components](#the-four-components)
  - [lowerdir — The Read-Only Base](#lowerdir--the-read-only-base)
  - [upperdir — The Writable Layer](#upperdir--the-writable-layer)
  - [workdir — The Kernel Scratch Space](#workdir--the-kernel-scratch-space)
  - [merged — The Unified View](#merged--the-unified-view)
- [How the Kernel Resolves Files](#how-the-kernel-resolves-files)
- [The Copy-Up Mechanism](#the-copy-up-mechanism)
- [Whiteout Files — How Deletes Work](#whiteout-files--how-deletes-work)
- [Opaque Directories — How Directory Deletes Work](#opaque-directories--how-directory-deletes-work)
- [Multiple Lower Layers](#multiple-lower-layers)
- [Hands-on Demo from Scratch](#hands-on-demo-from-scratch)
- [Mounting OverlayFS — Full Syntax](#mounting-overlayfs--full-syntax)
- [Real-world: Docker Uses OverlayFS](#real-world-docker-uses-overlayfs)
- [Inspecting Overlay Mounts](#inspecting-overlay-mounts)
- [OverlayFS vs Other Union Filesystems](#overlayfs-vs-other-union-filesystems)
- [Rules and Constraints](#rules-and-constraints)
- [Common Errors and Fixes](#common-errors-and-fixes)
- [Resources](#resources)

---

## What is OverlayFS?

**OverlayFS** (Overlay Filesystem) is a **union filesystem** built into the Linux kernel since version 3.18. It lets you **stack multiple directories on top of each other** into a single unified mount point, where:

- Lower directories are **read-only** (base layers)
- One upper directory is **writable** (your changes)
- A merged directory shows the **combined view** of everything

```
What you see (merged/)
        ▲
        │  unified view
┌───────┴───────────────────────────┐
│  upperdir/   (your writes)        │  ← writable
│  workdir/    (kernel internal)    │  ← kernel only
├───────────────────────────────────┤
│  lowerdir3/  (image layer 3)      │  ← read-only
│  lowerdir2/  (image layer 2)      │  ← read-only
│  lowerdir1/  (image layer 1)      │  ← read-only
└───────────────────────────────────┘
```

This is the exact mechanism Docker, Podman, and container runtimes use to implement **image layers** and **container filesystems**.

---

## The Four Components

### `lowerdir` — The Read-Only Base

The lower directory (or directories) is the **foundation** of the overlay. It is **never modified** — the kernel treats it as strictly read-only, regardless of its actual permissions on disk.

**Characteristics:**
- Read-only at the OverlayFS level (even if the directory itself is writable)
- Can be on a completely **different filesystem** than upperdir
- Can be **stacked multiple layers deep** (separated by `:`)
- Up to **500 lower layers** are supported by the kernel
- Can be on **tmpfs, ext4, xfs, btrfs, squashfs**, and more

**What lives here:**
- The original base OS files
- Docker image layers (each `RUN`, `COPY`, `FROM` instruction)
- Any pre-existing state you want to preserve

```
lower/
├── etc/
│   └── hosts          ← original, never touched
├── usr/
│   └── bin/
│       └── python3    ← original binary
└── app/
    └── config.yaml    ← original config
```

---

### `upperdir` — The Writable Layer

The upper directory is where **all writes go**. When you create, modify, or delete files through the `merged` mount, those changes are recorded here.

**Characteristics:**
- Fully writable
- Must be on the **same filesystem** as `workdir`
- Records three types of changes:
  - **New files** — created directly
  - **Modified files** — copy-up from lower, then edited
  - **Deleted files** — whiteout markers placed here

**What lives here after some operations:**

```
upper/
├── config.yaml        ← copy-up from lower (modified)
├── new-file.txt       ← created fresh in merged
└── deleted-file.txt   ← whiteout marker (char device 0,0)
```

When you delete a container, this directory is discarded. The lower layers (image) stay intact and are shared by other containers.

---

### `workdir` — The Kernel Scratch Space

The work directory is an **internal scratch space** used by the kernel during copy-up and atomic write operations. You never interact with it directly.

**Rules (strict):**
- Must be on the **exact same filesystem** as `upperdir` (same mount point / device)
- Must be **completely empty** before mounting
- Must not be used for anything else while the overlay is mounted
- The kernel creates a subdirectory called `work` inside it automatically

**Why it exists:** When the kernel performs a copy-up (copying a file from lowerdir to upperdir before modifying it), it needs a temporary staging area to make the operation atomic. It writes to workdir first, then renames into upperdir — so you never see a half-written file.

```
workdir/
└── work/              ← kernel creates and manages this
    └── ...            ← temporary staging files (usually empty when idle)
```

---

### `merged` — The Unified View

The merged directory is the **mount point** — the directory you actually use. It presents a seamless, unified view of all layers combined.

**What you see in merged:**
- All files from all lower layers
- All files from upper layer (which override lower files of the same name)
- Whiteout markers are invisible (deleted files simply don't appear)

```
merged/           ← what YOU see
├── etc/hosts     ← from lower (unmodified)
├── config.yaml   ← from upper (modified copy)
├── new-file.txt  ← from upper (newly created)
└── usr/bin/...   ← from lower (unmodified)
# deleted-file.txt is invisible (whiteout in upper)
```

---

## How the Kernel Resolves Files

When you read a file from `merged/`, the kernel applies a strict lookup order — **top to bottom**, stopping at the first match:

```
Request: open("merged/config.yaml")
              │
              ▼
   1. Check upperdir/config.yaml ──── found? → serve it ✅
              │ not found
              ▼
   2. Check lowerdir3/config.yaml ─── found? → serve it ✅
              │ not found
              ▼
   3. Check lowerdir2/config.yaml ─── found? → serve it ✅
              │ not found
              ▼
   4. Check lowerdir1/config.yaml ─── found? → serve it ✅
              │ not found
              ▼
         ENOENT (File Not Found) ❌
```

This means **upperdir always wins** over any lower layer. And among lower layers, the **leftmost in the mount command** has the highest priority.

---

## The Copy-Up Mechanism

This is the most important operation in OverlayFS. When you **modify a file that exists only in lowerdir**, the kernel cannot write there (read-only). Instead it performs a **copy-up**:

```
Step 1: You write to merged/config.yaml
              │
              ▼
Step 2: Kernel checks — does upperdir/config.yaml exist?
        → No
              │
              ▼
Step 3: Copy-up begins
        lowerdir/config.yaml ──copy──► workdir/work/tmpXXXX  (atomic staging)
                                                │
                                                ▼ rename (atomic)
                                        upperdir/config.yaml
              │
              ▼
Step 4: Your write is applied to upperdir/config.yaml
        lowerdir/config.yaml is NEVER touched
```

**Key properties of copy-up:**
- The copy is **metadata-preserving** (permissions, ownership, timestamps copied)
- The operation is **atomic** — you never see a partial file (thanks to workdir)
- Only the **modified file** is copied up, not the whole directory tree
- Subsequent writes to the same file skip copy-up (it's already in upper)
- Copy-up happens **on first write**, not on open (lazy copy)

**Performance implication:** The first write to a large file in lowerdir causes a full file copy. Subsequent writes are fast (direct to upper). This is why Docker advises keeping frequently-changed files in upper (container layer) rather than baking them into lower image layers.

---

## Whiteout Files — How Deletes Work

Since lowerdir is read-only, you can't actually delete files from it. Instead, OverlayFS uses **whiteout files** in upperdir to mark files as deleted.

**When you delete `merged/base.txt`:**

```
rm merged/base.txt
      │
      ▼
Kernel creates: upperdir/base.txt
  → type: character device
  → major: 0, minor: 0
  → this is the whiteout marker
      │
      ▼
lowerdir/base.txt still exists on disk
      │
      ▼
merged/base.txt is invisible (kernel hides it)
```

**Inspecting a whiteout:**

```bash
ls -la upper/base.txt
# crw-r--r-- 1 root root 0, 0 Jun 2025 base.txt
#  ↑
#  c = character device
#  0, 0 = major 0, minor 0 = whiteout marker

stat upper/base.txt
# File: upper/base.txt
# Size: 0   Blocks: 0   IO Block: 4096   character special file
# Device type: 0,0
```

The kernel recognizes `(0, 0)` character devices in upperdir as whiteout markers and hides the corresponding entry from `merged/`.

---

## Opaque Directories — How Directory Deletes Work

Whiteouts handle individual files. For **directories**, OverlayFS uses a different trick: **opaque directories**.

When you delete a directory from lowerdir and recreate it (or when a directory in upper should completely shadow a lower directory), the kernel marks the upperdir version with an extended attribute:

```bash
# After: rm -rf merged/somedir && mkdir merged/somedir
getfattr -n trusted.overlay.opaque upper/somedir
# trusted.overlay.opaque = "y"
```

The xattr `trusted.overlay.opaque=y` tells the kernel: **do not look into any lower layer for this directory's contents**. Everything inside comes only from upperdir.

**Without opaque:**
```
merged/somedir/  ← shows files from BOTH upper/somedir and lower/somedir
```

**With opaque:**
```
merged/somedir/  ← shows ONLY files from upper/somedir (lower is masked)
```

---

## Multiple Lower Layers

One of the most powerful features of OverlayFS is stacking **multiple read-only lower layers** using `:` as a separator:

```bash
sudo mount -t overlay overlay \
  -o lowerdir=layer3:layer2:layer1,upperdir=upper,workdir=work \
  merged/
```

**Priority order (left = highest):**

```
merged/         ← unified view
    ▲
    │
layer3          ← checked first among lowers
layer2          ← checked second
layer1          ← checked last (base)
```

**Real example with Docker-like layers:**

```
merged/                   ← running container
    ▲
    │
container_upper/          ← container's writes (upperdir)
    ─────────────────
image_layer3/             ← COPY . /app          (lowerdir, highest)
image_layer2/             ← RUN npm install      (lowerdir)
image_layer1/             ← FROM node:18         (lowerdir, lowest)
```

**Rules for multiple lowers:**
- They can be on **different filesystems**
- A file in `layer3` shadows the same path in `layer2` and `layer1`
- The kernel merges directory **listings** from all layers (union of all entries)
- For files: first match wins (no merging of file content)
- Maximum: **500 lower layers** per mount

---

## Hands-on Demo from Scratch

Here is a complete, step-by-step demo you can run right now:

```bash
# ── Setup ─────────────────────────────────────────────────

mkdir -p /tmp/overlay/{lower,upper,work,merged}

# Populate the lower layer (read-only base)
echo "I am the original base file"   > /tmp/overlay/lower/base.txt
echo "original config content"       > /tmp/overlay/lower/config.txt
echo "lower-only file"               > /tmp/overlay/lower/readonly.txt
mkdir /tmp/overlay/lower/subdir
echo "nested file"                   > /tmp/overlay/lower/subdir/nested.txt


# ── Mount ─────────────────────────────────────────────────

sudo mount -t overlay overlay \
  -o lowerdir=/tmp/overlay/lower,\
upperdir=/tmp/overlay/upper,\
workdir=/tmp/overlay/work \
  /tmp/overlay/merged


# ── Read operations (no copy-up) ──────────────────────────

echo "=== Reading from merged ==="
cat /tmp/overlay/merged/base.txt
# I am the original base file

ls /tmp/overlay/merged/
# base.txt  config.txt  readonly.txt  subdir/

# Upper is still empty — reads don't trigger copy-up
ls /tmp/overlay/upper/
# (empty)


# ── Write: modify a lower file (triggers copy-up) ─────────

echo "=== Modifying config.txt ==="
echo "MODIFIED config" > /tmp/overlay/merged/config.txt

echo "--- merged sees:"
cat /tmp/overlay/merged/config.txt
# MODIFIED config

echo "--- lower is untouched:"
cat /tmp/overlay/lower/config.txt
# original config content

echo "--- upper has the copy:"
cat /tmp/overlay/upper/config.txt
# MODIFIED config


# ── Write: create a new file ──────────────────────────────

echo "=== Creating a new file ==="
echo "brand new file" > /tmp/overlay/merged/newfile.txt

ls /tmp/overlay/upper/
# config.txt  newfile.txt   ← only changes, not the whole fs

ls /tmp/overlay/lower/
# base.txt  config.txt  readonly.txt  subdir/   ← unchanged


# ── Delete: remove a lower file (creates whiteout) ────────

echo "=== Deleting base.txt ==="
rm /tmp/overlay/merged/base.txt

echo "--- merged no longer shows it:"
ls /tmp/overlay/merged/
# config.txt  newfile.txt  readonly.txt  subdir/

echo "--- lower still has it:"
ls /tmp/overlay/lower/
# base.txt  config.txt  readonly.txt  subdir/

echo "--- upper has whiteout marker:"
ls -la /tmp/overlay/upper/base.txt
# crw-r--r-- 1 root root 0, 0 ... base.txt  ← char device 0,0


# ── State summary ─────────────────────────────────────────

echo ""
echo "=== Final state ==="
echo "MERGED contents:"
ls /tmp/overlay/merged/

echo "UPPER contents (your changes only):"
ls /tmp/overlay/upper/

echo "LOWER contents (untouched original):"
ls /tmp/overlay/lower/


# ── Cleanup ───────────────────────────────────────────────

sudo umount /tmp/overlay/merged
rm -rf /tmp/overlay
```

---

## Mounting OverlayFS — Full Syntax

```bash
sudo mount -t overlay overlay \
  -o lowerdir=<lower>[:<lower2>:<lower3>...],\
     upperdir=<upper>,\
     workdir=<work> \
  <merged>
```

**Read-only overlay (no upper/work needed):**
```bash
# When you only need a unified read-only view
sudo mount -t overlay overlay \
  -o lowerdir=layer2:layer1 \
  merged/
```

**In `/etc/fstab`:**
```
overlay /merged overlay lowerdir=/lower,upperdir=/upper,workdir=/work 0 0
```

**Programmatically (in C or Go):**
```c
mount("overlay", "/merged", "overlay", 0,
      "lowerdir=/lower,upperdir=/upper,workdir=/work");
```

---

## Real-world: Docker Uses OverlayFS

Docker's `overlay2` storage driver is a direct implementation of OverlayFS. Every image and container maps to these directories.

**Where Docker stores overlay data:**
```
/var/lib/docker/overlay2/
├── <layer-sha256>/          ← each image layer = one lowerdir
│   └── diff/                ← the actual files for this layer
├── <layer-sha256>/
│   └── diff/
└── <container-sha256>/      ← running container
    ├── diff/                ← upperdir (container writes)
    ├── work/                ← workdir
    └── merged/              ← what the container sees
```

**Inspect a running container's overlay mount:**
```bash
# Find the container's PID
docker inspect <container> --format '{{.State.Pid}}'

# See its mount info
cat /proc/<pid>/mounts | grep overlay

# Or inspect directly
docker inspect <container> | grep -A 10 "GraphDriver"
# "Data": {
#   "LowerDir":  "/var/lib/docker/overlay2/abc.../diff:...",
#   "MergedDir": "/var/lib/docker/overlay2/xyz.../merged",
#   "UpperDir":  "/var/lib/docker/overlay2/xyz.../diff",
#   "WorkDir":   "/var/lib/docker/overlay2/xyz.../work"
# }
```

**What happens on `docker commit`:**
```
Container upperdir  ──snapshot──►  new image layer (lowerdir)
                                   (upperdir is frozen, becomes a lower)
```

**What happens on `docker run` (second container from same image):**
```
image layers (shared lowerdirs, unchanged)
    ▲
    │
container1/upper/   ← container 1's writes (isolated)

image layers (same shared lowerdirs!)
    ▲
    │
container2/upper/   ← container 2's writes (isolated)
```

Both containers share all image layers. Only their `upperdir` is unique. This is why Docker images are **space-efficient** — 100 containers from the same image share the image layers on disk.

---

## Inspecting Overlay Mounts

```bash
# See all current overlay mounts
mount | grep overlay

# Detailed mount info with options
cat /proc/mounts | grep overlay

# See overlay info for a specific directory
findmnt /tmp/overlay/merged

# Check if your kernel supports overlayfs
grep -i overlay /proc/filesystems
# nodev   overlay   ← good, it's supported

# Kernel module status
lsmod | grep overlay
# overlay   ...

# Load if not loaded
sudo modprobe overlay
```

---

## OverlayFS vs Other Union Filesystems

| Feature | OverlayFS | AUFS | UnionFS | BindFS |
|---------|-----------|------|---------|--------|
| In mainline kernel | ✅ since 3.18 | ❌ patch | ❌ FUSE | ✅ |
| Multiple lower layers | ✅ up to 500 | ✅ | ✅ | ❌ one |
| Read-only lowers | ✅ | ✅ | ✅ | ❌ |
| Copy-up on write | ✅ | ✅ | ✅ | N/A |
| Whiteout files | ✅ char 0,0 | ✅ `.wh.` files | ✅ | N/A |
| Performance | ✅ fast | medium | slow | fast |
| Used by Docker | ✅ overlay2 | legacy | no | no |
| upper+work same FS | required | no | no | N/A |
| Squashfs as lower | ✅ | ✅ | limited | ❌ |

---

## Rules and Constraints

These are the hard rules enforced by the kernel — violating them gives a mount error:

```
✅  lowerdir   can be on a different filesystem than upper
✅  lowerdir   can be on squashfs, tmpfs, ext4, xfs, btrfs, etc.
✅  multiple   lowerdirs separated by : (up to 500)
✅  read-only  overlay possible with just lowerdir (no upper/work)

❌  upperdir and workdir MUST be on the same filesystem (same device)
❌  workdir   must be empty before mounting
❌  workdir   must not be inside upperdir
❌  lowerdir  must not be inside upperdir or workdir
❌  overlayfs on overlayfs not supported in all kernels
```

**Common constraint errors:**
```
# Wrong: upper and work on different filesystems
mount: wrong fs type ... "invalid argument"
→ Fix: ensure upper/ and work/ are on the same partition

# Wrong: workdir not empty
mount: ... "directory not empty"
→ Fix: rm -rf work/* before mounting

# Wrong: no root permissions
mount: only root can use "--types" here
→ Fix: use sudo, or use user namespaces (unshare -U)
```

---

## Common Errors and Fixes

| Error | Cause | Fix |
|-------|-------|-----|
| `invalid argument` on mount | upper and work on different filesystems | Move both to same partition |
| `directory not empty` | workdir has leftover files | `rm -rf work/*` |
| `permission denied` | Not root | Use `sudo` |
| Copy-up very slow | Large file in lowerdir modified | Restructure layers, put large changing files in upper |
| `too many levels of symbolic links` | Overlay stacked too deep | Flatten layers |
| Changes lost after umount | Wrote to merged but forgot to keep upper | Backup upperdir before umount |

---

## Resources

- 📘 [Linux Kernel OverlayFS Documentation](https://www.kernel.org/doc/html/latest/filesystems/overlayfs.html)
- 📘 [man 8 mount — overlay type](https://man7.org/linux/man-pages/man8/mount.8.html)
- 🔗 [Docker overlay2 storage driver](https://docs.docker.com/storage/storagedriver/overlayfs-driver/)
- 🔗 [Containers from Scratch — Liz Rice (uses overlayfs)](https://github.com/lizrice/containers-from-scratch)
- 🔗 [OverlayFS merged into kernel — LWN](https://lwn.net/Articles/610765/)
- 📘 [Understanding Docker Storage Drivers](https://docs.docker.com/storage/storagedriver/)

---

## Quick Reference Cheatsheet

```bash
# Basic overlay mount
sudo mount -t overlay overlay \
  -o lowerdir=/lower,upperdir=/upper,workdir=/work \
  /merged

# Multiple lower layers (left = highest priority)
sudo mount -t overlay overlay \
  -o lowerdir=/layer3:/layer2:/layer1,upperdir=/upper,workdir=/work \
  /merged

# Read-only overlay (no writes needed)
sudo mount -t overlay overlay \
  -o lowerdir=/layer2:/layer1 \
  /merged

# Unmount
sudo umount /merged

# Inspect current overlays
mount | grep overlay
cat /proc/mounts | grep overlay

# Check whiteout file
ls -la upper/deleted-file     # crw-r--r-- 0, 0 = whiteout

# Check opaque directory
getfattr -n trusted.overlay.opaque upper/somedir
```

---

<div align="center">

**Star ⭐ this repo if it helped you understand OverlayFS and container internals!**

</div>
