# lil-gator-game-mp-server

A lightweight WebSocket server written in Go that powers the [Gator Gang](https://github.com/KooperL/lil-gator-game-mp) co-presence multiplayer mod for Lil Gator Game.
This server handles session management and relays player movement data between connected clients.

Clients connect via:

```ws://<configurable-host>?sessionKey=xxx&clientVersion=x.x.x```

Where `<configurable-host>` is defined in the `lggmp_config.ini` file. As a server hoster, you may present the server directly via IP and port, or through a combination of subdomains, domains and paths; though the full URL (minus the protocol) needs to be communicated to clients. **DO NOT** add query params to a server without making custom modifications to the client's `.dll`.
