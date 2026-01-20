<h1 align="center">
  ⬡ plugstep
</h1>

<p align="center">
  <strong>Declarative and reproducible Minecraft server deployment.</strong><br>
  Core component to the Mineframe stack.
</p>

<p align="center">
  <a href="#quick-example">Example</a> •
  <a href="#whats-mineframe">About</a> •
  <a href="#can-i-use-this-on-my-server">License</a> •
  <a href="#how-is-this-better">Why Plugstep?</a>
</p>

---

## Quick Example

<table>
<tr>
<td>

**plugstep.toml**

```toml
[server]
vendor = "papermc"
project = "paper"
minecraft_version = "1.21.8"
version = "latest"

[[plugins]]
source = "modrinth"
resource = "luckperms"
version = "v5.5.0-bukkit"
```

</td>
<td align="center" valign="middle">

**→**

</td>
<td>

**Output**

```
INFO Initializing Plugstep ♪(๑ᴖ◡ᴖ๑)♪
|--------------------------------------|
|   Server vendor     papermc          |
|   Server project    paper            |
|   Minecraft version 1.21.8           |
|   Server version    latest           |
|--------------------------------------|
INFO Downloaded server JAR successfully.
INFO Starting plugin download plugins=1
INFO Installed luckperms
INFO Plugins ready. installed=1 checked=0
```

</td>
</tr>
</table>

---

## What's Mineframe?

Mineframe is a work in progress scalable and developer friendly technology stack for developing Minecraft servers. It relies on technology like Docker, Plugstep and Schemastash (TODO) to create a unified and scalable Mineframe network. Craftops was designed for use on the McWar server, but we decided to turn it into an independent project.

## Can I use this on my server?

Yes! Plugstep is licensed under **GPLv3** meaning you can freely use it on your server, whether commercial or not. Please do note that the GPLv3 license is a copyleft license, so all derivative work must also be licensed under it.

## How is this better?

Plugstep is designed to allow a reproducible Minecraft server to be made. This means that multiple developers can develop concurrently, use version control software to combine their work, and deploy it in production without ever manually dragging a single file. It may not be ideal for a small server like a private SMP, but in situations where the ability to spin up an exact replica server in seconds is useful, Plugstep does the hard work for you.

---

<p align="center">
  <strong>Part of Mineframe</strong><br>
  <sub>Opinionated stack for declarative and reproducible Minecraft networks</sub>
</p>
