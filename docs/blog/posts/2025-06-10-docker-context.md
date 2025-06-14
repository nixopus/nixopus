---
title: How Docker Contexts Transformed My Multi-Environment Workflow
description: Nixopus internally relies on Docker for running and managing its services, so the script had to ensure full and secure access to Docker on the host machine.
date: 2025-06-10
author: Nixopus Team
---

One fine evening, I was working on my project [Nixopus](https://github.com/raghavyuva/nixopus). Nixopus is a self-hostable application that puts an end to the chaos of traditional VPS management.. As part of the setup process, I had written a script to automate its deployment across different environments.

## The Context of the Problem

Nixopus internally relies on Docker for running and managing its services, so the script had to ensure full and secure access to Docker on the host machine. Here's what it needed to handle across any environment:


![Nixopus Docker](https://dev-to-uploads.s3.amazonaws.com/uploads/articles/7d16xn98z36h7hx9hivr.png)

1. **Verify Docker installation** — Check whether Docker was already installed on the server.
2. **Install Docker if missing** — Make sure to install docker and docker dependencies that are compatible.
3. **Version compatibility check** — If Docker is already present, compare its version with what Nixopus requires to ensure smooth operation.
4. **Generate TLS certificates** — Use [`openssl`](https://www.openssl.org/) to create certificates that allow secure connections to Docker over [TLS](https://en.wikipedia.org/wiki/Transport_Layer_Security).
5. **Configure Docker daemon** — Modify the `daemon.json` file to:
   - Use the generated TLS certificates,
   - Enable secure communication,
   - Bind Docker to the required network interfaces.
6. **Validate Docker access** — Reload or restart Docker using `systemd` (or another service manager) and ensure secure remote access works as expected.

## The Chaotism Begins

I *thought* I was doing everything right — but then, one by one, problems began piling up like a stack. Let me walk you through the chaos:


![Image description](https://dev-to-uploads.s3.amazonaws.com/uploads/articles/nvjd1s6j4moalm32kfzc.png)

1. **Conflicting [`daemon.json` setups](https://docs.docker.com/engine/daemon/)**
   What if the user already had a `daemon.json` configured for their own purposes? Should we just go ahead and overwrite it with our custom setup?  
   At first, I naïvely said “yes.” And that’s when it hit me — we were doomed. Overwriting user configurations without merging or backing them up? A recipe for disaster.

2. **No isolation between environments**  
   What if I wanted to install Nixopus on the same machine for both `production` and `staging`? Docker doesn't support running two separate daemons bound to different environments by default. Suddenly, our clean setup felt too rigid. I needed some form of isolation that wouldn't require spinning up full-blown [VMs](https://www.vmware.com/topics/virtual-machine) or resorting to hacky workarounds.

3. **Credential conflicts and confusion**  
   Each environment had its own TLS certs and binding requirements. But trying to manage them manually — keeping track of which certs belonged to which context, ensuring the right IP was exposed — quickly became error-prone. One small mistake and I’d be locked out of Docker, or worse, end up exposing a production port unintentionally.

This is where things spiraled. Each new problem broke my illusion of control and made it clear: we needed a better, more structured way to manage Docker across environments — without losing sanity or safety.

And that’s when Docker **contexts** entered the scene - like a life saviour

## Efficient Chaos Management with Docker Context


![Docker Context Example](https://dev-to-uploads.s3.amazonaws.com/uploads/articles/f2dj1y14ytuad251bkwi.png)

[Docker Context](https://docs.docker.com/engine/manage-resources/contexts/) lets me solve many of the problems I used to face in multi-environment Docker workflows. It brings **isolation**, **simplicity**, and **clarity** to the way I manage local, staging, and production setups.

With Docker Context, I can:
- Isolate environments cleanly
- Use different ports for different daemons
- Configure environment-specific [TLS certificates](https://www.digicert.com/tls-ssl/tls-ssl-certificates)
- Avoid the chaos of switching between shell exports, flags, or config files

And the best part? It’s _really_ easy to use. No messy commands. No juggling environment variables. Just simple CLI instructions:

```
docker context create production 
docker context use production 
docker context ls
```

With these, I can switch between environments effortlessly, knowing that each one is securely configured and neatly separated from the rest

## Have We Solved Our Docker Daemon Problem Yet?


![Docker Deamon](https://dev-to-uploads.s3.amazonaws.com/uploads/articles/yqa0zj2vow0213dpy7p9.png)

While Docker Contexts make it easy to switch between different Docker environments from the client side, they don’t solve the problem of running multiple Docker daemons with separate `daemon.json` configurations. Docker still expects a single, global `daemon.json` file — there's no built-in way to maintain separate configs for different environments on the same machine

### So how did i work around that?

Here’s the trick: instead of trying to juggle multiple `daemon.json` files, i  used **systemd service overrides**. This allows me to run **multiple Docker daemons**, each configured with its own port, TLS settings, and data directory — without touching the default config.

With Docker Contexts handling the client-side switching and [systemd overrides](https://www.freedesktop.org/software/systemd/man/latest/systemd.service.html) managing isolated daemon instances, I finally got true multi-environment control — clean, secure, and conflict-free.


## What is Nixopus?

![Nixopus - Streamline your server workflow](https://dev-to-uploads.s3.amazonaws.com/uploads/articles/rxjzvcy6azcp8cowa6de.png)

[Nixopus is a self-hostable application](https://dev.to/raghavyuva/introducing-nixopus-all-in-one-open-source-vps-management-solution-4keg) that puts an end to the chaos of traditional [VPS](https://en.wikipedia.org/wiki/Virtual_private_server) management.

Whether you're running multiple environments, deploying Docker containers, managing static sites, or just tinkering with servers for fun — Nixopus gives you a clean, powerful web interface to do it all.

No more memorizing commands, editing obscure config files, or SSH-ing into every box manually. With Nixopus, you can focus on what really matters: building your side project, running your business, or learning infrastructure at your own pace — without the stress.It’s open-source, community-driven, and designed for developers who value **clarity**, **control**, and **simplicity**.
## Did You Know About Docker Contexts and This Hack?

Were you already using Docker Contexts or aware of this systemd override trick?  
Or how do *you* usually manage these kinds of multi-environment challenges?

Maybe there’s an even better solution we haven’t thought of yet — and we’d love to hear about it.

At Nixopus, we believe in **community-driven development**. We're building a powerful, open-source VPS workflow manager — and solving real-world infrastructure pain points together.

If you're passionate about DevOps, automation, or self-hosting, [join our community](https://discord.gg/skdcq39Wpv) and help shape the future of server management.
