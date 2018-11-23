Mollstam: A Minecraft Monitoring Discord Bot
============================================
Mollstam integrates Minecraft server information with Discord. It can update a channel's name with the current player count and the topic with a list of online players. It can also notify a user when the server is unreachable.

Installation
------------
Simply `go get` this repository.

Configuration
-------------
Mollstam reads `config.json` in the current directory by default. The config file location can by specified with the `-c` flag e.g. `$ mollstam -c path/to/file.json`. The config specification is in `config.go`. An example configuration is provided below.

```JSON
{
    "address": "127.0.0.1:25565",
    "polling_rate": "30s",
    "timeout": "5s",
    "channel_id": "",
    "channel_name": "minecraft",
    "discord_token": "",
    "notify_user_id": "",
    "notify_message": "The server appears to be offline."
}
```
