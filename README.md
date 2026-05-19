# 🍵 GOChat

Un clon de chat P2P hiper-minimalista inspirado en la arquitectura **IRC**, construido en **Go** utilizando el ecosistema moderno de TUIs de Charm (**Bubble Tea** y **Lip Gloss**).

## 💎 Características

* **Arquitectura IRC-Like:** Conexiones directas por sockets TCP (Host/Cliente) con intercambio síncrono de Nicks.
* **TUI Moderna:** Interfaz interactiva y limpia usando `bubbletea` (Elm Architecture) y `viewport` para el scroll.
* **Estilos Premium:** Diseño visual con bordes redondeados y paleta de colores TrueColor gracias a `lipgloss`.
* **Buffer DB (Historial):** El Host mantiene una base de datos local ultra-ligera en JSON (`~/.gochat_history.json`). Al conectarse un cliente nuevo, se le inyecta todo el historial automáticamente.
* **Notificaciones Limpias:** Avisos de desconexión en tiempo real (`[!] Usuario ha salido`) sin ensuciar la base de datos.
* **Control de Versiones Amigable:** Desarrollado pensando en la compatibilidad nativa de Git y Jujutsu (`jj`).

---

## 🛠️ Requisitos Previos

Asegúrate de tener instalado **Go 1.18 o superior** y las dependencias del proyecto:

```bash
git clone https://github.com/Qmaker-programmer/gochat.git
cd gochat
go mod tidy
```
# ¿Como usarlo?
**Modo desarrollador (Ejècucion rapida)**
```bash
make run
```
**Comandos de chat**
Dentro de la interfaz de la TUI puedes usar los siguientes comandos nativos:
* `/clear`
* `/exit`
* `CTRL+C`

# Compilacion y distribuciòn
**El proyecto incluye un Makefile automatizado para realizar compilación cruzada estática optimizada (-ldflags="-s -w"):**
* Compilar para tu sistema actual:
```bash
make build
```
**(El ejecutable quedará en bin/GOChat si es windows .exe)**
* Compilación Cruzada Total (6 binarios: Linux, Windows, macOS en amd64 y arm64):
```bash
make build-all
```
**(Ideal para generar el binario que le pasarás a otros usuarios).**
# ⚙️  Persistencia
**Las opciones de tu última sesión (Nick, IP y Puerto) se guardan automáticamente en tu directorio Home (~/.shchat.json) para que no tengas que reescribirlas cada vez que abres la app.**
# 💾 Salvando el progreso en Jujutsu / Git

Una vez que hayas completado una feature, es hora de registrar este hito en tu control de versiones. 
```bash
git add .
git commit -m "
Features:
tus features...
"
```
**opcionalmente si quieres contribuir al proyecto, tienes que crear un fork, despues un push a tu fork den github, y un PR(Pull request)**
