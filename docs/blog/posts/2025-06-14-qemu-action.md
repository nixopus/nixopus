---
title: The Gem of a Github Action you never used
description: It's time I show you one of the coolest GitHub Actions we're using at Nixopus.If you're like me, you're probably more curious about the “why” than just hearing the solution — QEMU
date: 2025-06-14
author: Nixopus Team
---

It's time I show you one of the coolest GitHub Actions we're using at [Nixopus](https://github.com/raghavyuva/nixopus).If you're like me, you're probably more curious about the **“why”** than just hearing the solution — QEMU.

So before diving into how we use QEMU, let me walk you through the *why* — so the context is clear and the solution makes perfect sense.

## Too Many Moving Parts: Why We Needed Emulation?

Nixopus is a platform that streamlines your entire VPS/server workflow. But deploying it wasn’t always smooth — we hit a few bottlenecks that made us rethink our existing approach.

![The need of Emulation](https://dev-to-uploads.s3.amazonaws.com/uploads/articles/uvovsktgnxgf3bfax5qu.png)

Here’s the situation, We offer a **self-hosting one-liner installation script**. This script handles:
- Docker setup  
- SSH configuration  
- Proxy management  
- Bringing up Nixopus services (API, database, etc.)

Naturally, this raised a few critical questions:
1. How can we test this installation script *every time* we make changes?
2. How can we be sure it works across all major Linux distributions, regardless of their initial state?
3.  How do we verify the script works with a **matrix of different parameters**?

## The Obvious Solution You’d Think Of

If you're anything like me, your first instinct is to think of a **quick fix**, then improve upon it. So I started thinking in these directions:

### 1. "Let’s Write Unit Tests for Each Function?"

That was my first idea — the classic starting point. But as you’ll soon see, it doesn’t really address the kind of problems we’re facing. Sure, unit tests are great. It’s tempting to think “Let’s just write a test for each function and run it on every change via CI/CD. Easy, right?”

*Sounds good in theory.*  But *falls apart in practice.*

Why? Because we're not just dealing with simple logic here — we’re dealing with [the full complexity of what our installer sets up](#-why-did-we-need-qemu). Now ask yourself : How do you unit test that without turning your test suite into a full-blown Docker simulator Mocking all of that? At that point, you’re not testing your install script... you’re testing your mock environment.

![The Secret Funding](https://dev-to-uploads.s3.amazonaws.com/uploads/articles/rlqs2a2s6jxailn033h0.png)

And that’s when I knew — unit tests alone weren’t going to cut it

### 2. "How about manually testing the installer before release?"

I figured — if nothing else works, I’ll just set up a GitHub Action that SSHes into my VPS and runs the installer.  I even explored [`appleboy/ssh-action`](https://github.com/appleboy/ssh-action) to automate this through GitHub Actions.

Sounds promising, right? You could:
- Run tests on every push
- Validate the installer across multiple config permutations using matrix builds
- Catch failures early

But here’s where it started to fall apart:
1. You need a **dedicated VPS just for testing**  
2. How do you **run parallel executions** of the installer, especially when `systemd` is involved?  
3. Want to test different distros? Well… you'd need a VPS for *each* one.

![Complexity of Multiple Distribution Testing](https://dev-to-uploads.s3.amazonaws.com/uploads/articles/h2bzvw81p3qhis7vlxsp.png)

### 3. "Virtualization is the easy-peasy solution, right?"

At this point, the idea of spinning up virtual machines felt like a natural next step.  We were close — testing on a real VPS worked, but lacked multi-distro flexibility.

That’s when tools like [HashiCorp Vagrant](https://developer.hashicorp.com/vagrant) caught my eye.  It provides a full-featured way to manage virtual machine lifecycles.  

So in theory:
- We dedicate one VPS for testing
- Spin up VMs for each target distro
- Run the installation script in each

Sounds solid. But then reality kicked in:  
**"Wait... aren’t we over-engineering this just to test an install script?"**

Let’s be honest:
1. Do we really want to maintain VM lifecycle logic inside our CI?  
2. Do we even have enough system resources on a single VPS to run multiple distros *in parallel*?

#### Now you might be thinking — "Why not just use containers?"

Yup. Same thought crossed my mind. But that idea died the moment I remembered  
We’re dealing with `systemd`, low-level networking, service bootstrapping — stuff containers just aren’t great at simulating. **And that’s how we ended up looking at full system emulation...**

## The Life Saviour: [QEMU GitHub Action](https://github.com/docker/setup-qemu-action) from Docker

```yaml
name: ci

on:
  push:

jobs:
  qemu:
    runs-on: ubuntu-latest
    steps:
      - name: Set up QEMU
        uses: docker/setup-qemu-action@v3
```

[QEMU](https://www.qemu.org/) is a generic and open-source machine emulator and virtualizer.

After walking through all those pain points — distro variations, systemd compatibility, real OS behavior, parameter matrix testing — QEMU turned out to be the **cleanest and most reliable solution**. 

And the best part?  
It worked out-of-the-box _without any fuss_

## What Did We Learn From All This?

![Nixopus Logo](https://dev-to-uploads.s3.amazonaws.com/uploads/articles/u1y3s1475irghih86kwm.png)

You might be thinking **“Why not just look at how other projects test their self-hosting installation scripts?”** You're right — that would’ve been the easier route.  But here’s the thing: we’re not just blindly following others.  We’re solving a real-world, reproducible problem _with intent_.

Others might have elegant test setups — or they might be winging it behind the scenes.  
What matters is the **process** we went through:

- Hitting walls
- Considering workarounds
- Evaluating complexity vs. scalability
- Finally arriving at a solution that fits **our needs**

And funny enough — all of this, just to test a one-liner installation script 😂

But hey, that’s engineering. 

If you know any better approach than this let's talk more and shape it out. if you are someone who is interested in such engineering take ups here we are [Join our Community](https://discord.gg/skdcq39Wpv) 