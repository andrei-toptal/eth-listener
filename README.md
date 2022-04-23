# eth-listener
A simple console (server) app that is listening to your ETH transactions.
Whenever it detects a transfer for one of the specified ETH accounts, it logs details to the console and also sends a notification to your Telegram bot.

## Configuration
All configuration is done by editing `config.yaml` file.
You must specify at least:
1. ETH node URL. If you don't have one, use 3rd-party providers such as [alchemy](https://alchemy.com/?r=62491cd8ac883927) for free.
2. Add ETH account addresses you want to watch, there may be many addresses. Each address can have a human-readable alias.

Additionally, if you wish to receive notifications to your TG bot:
3. Confgiure your Telegram bot by specifying bot's token and your Telegram username.

## Usage
After you edited `config.yaml`, just start the app by `go run .`
When a new transaction is detected, you will see more log entires such as "received" or "sent":
```
2022/04/23 09:28:10 Metamask sent 0.1 LINK to 0x313573780DB563D6574424A08740f24787a0D6Ba, new balance: 15.2734 LINK
```
To terminate the app, just hit `Ctrl+C`.

## Telegram integration
Telegram bot supports two commans: `/subscribe` and `/unsubscribe`.
The first command will enable bot's notifications and the second command will stop notifications.
The `username` specified in `config.yaml` will restrict other users to see your notifications and/or subscribe/unsubscribe.
