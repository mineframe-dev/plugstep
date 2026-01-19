<img width="100%" height="auto" alt="Slide 16_9 - 1(6)" src="https://github.com/user-attachments/assets/0aefb022-8e47-4cf8-97ed-0dd2c01dfea4" />

# plugstep

Declarative and reproducible Minecraft server deployment.
Core component to the Mineframe stack.

---

[server]
vendor = "papermc"
project = "paper"
minecraft_version = "1.21.8"
version = "latest"

[[plugins]]
source = "modrinth"
resource = "luckperms"
version = "v5.5.0-bukkit"

TURNS INTO

2025/09/18 22:43:11 INFO Initializing Plugstep ♪(๑ᴖ◡ᴖ๑)♪
|-----------------------------------------------|
| Server config: |
| |
| Server vendor papermc |
| Server project paper |
| Server Minecraft version 1.21.8 |
| Server version latest |
|-----------------------------------------------|
2025/09/18 22:43:12 INFO Downloaded server JAR successfully.
2025/09/18 22:43:12 INFO Starting plugin download plugins=1
2025/09/18 22:43:12 INFO Installed luckperms
2025/09/18 22:43:12 INFO Plugins ready. installed=1 checked=0

---

What’s Mineframe?
Mineframe is a work in progress scalable and developer friendly technology stack for developing Minecraft servers. It relies on technology like Docker, Plugstep and Schemastash (TODO) to create a unified and scalable Mineframe network. Craftops was designed for use on the McWar server, but we decided to turn it into an independent project.

Can I use this on my server?
Yes! Plugstep is licensed under GPLv3 meaning you can freely use it on your server, whether commercial or not. Please do note that the GPLv3 license is a copyleft license, so all derivative work must also be licensed under it.

How is this better?
Plugstep is designed to allow a reproducible Minecraft server to be made. This means that multiple developers can develop concurrently, use version control software to combine their work, and deploy it in production without ever manually dragging a single file. It may not be ideal for a small server like a private SMP, but in situations where the ability to spin up an exact replica server in seconds is useful, Plugstep does the hard work for you.

Part of
Mineframe
Opinionated stack for declarative and reproducible Minecraft networks
