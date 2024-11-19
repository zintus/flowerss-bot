# flowerss bot

[![Build Status](https://github.com/zintus/flowerss-bot/workflows/Release/badge.svg)](https://github.com/zintus/flowerss-bot/actions?query=workflow%3ARelease)
[![Test Status](https://github.com/zintus/flowerss-bot/workflows/Test/badge.svg)](https://github.com/zintus/flowerss-bot/actions?query=workflow%3ATest)
![Build Docker Image](https://github.com/zintus/flowerss-bot/workflows/Build%20Docker%20Image/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/zintus/flowerss-bot)](https://goreportcard.com/report/github.com/zintus/flowerss-bot)
![GitHub](https://img.shields.io/github/license/zintus/flowerss-bot.svg)

[Installation and Usage Documentation](https://flowerss-bot.now.sh/)

<img src="https://github.com/rssflow/img/raw/master/images/rssflow_demo.gif" width = "300"/>

## Features

- All common RSS Bot functionalities
- Support for Telegram in-app instant view
- Support for RSS message subscription in Groups and Channels
- Rich subscription settings

## Installation and Usage

For detailed installation and usage instructions, please refer to the project's [documentation](https://flowerss-bot.now.sh/).

Available commands:

```
/sub [url] Subscribe to RSS feed (url is optional)
/unsub [url] Unsubscribe from RSS feed (url is optional)
/list View current subscriptions
/set Configure subscription settings
/check Check current subscriptions
/setfeedtag [sub id] [tag1] [tag2] Set subscription tags (max 3 tags, space-separated)
/setinterval [interval] [sub id] Set refresh interval (multiple sub ids allowed, space-separated)
/activeall Activate all subscriptions
/pauseall Pause all subscriptions
/import Import OPML file
/export Export OPML file
/unsuball Unsubscribe from all feeds
/help Help
```

For detailed usage instructions, please refer to the project's [documentation](https://flowerss-bot.now.sh/#/usage).
