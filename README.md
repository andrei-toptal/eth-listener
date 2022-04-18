# eth-listener
A simple console (server) app that is listening to your ETH transactions.
Whenever it detects a transfer for one of your ETH accounts, it logs details to the console and also sends a notification to your Telegram bot.

## Configuration
All configuration is done by editing `config.yaml` file.
1. Specify your ETH node URL. If you don't have one, use 3rd-party providers such as [alchemy](https://alchemy.com/?r=62491cd8ac883927) to create one.
2. Enter your ETH account addresses you want to watch (there may be many addresses).
3. Enter your token contract addresses you want to watch (there may be many contract addresses).
4. Confgiure your Telegram bot by specifying bot's token and your Telegram username.

## Usage
After you edited `config.yaml`, just start the app by `go run .`
You will see the output like this:
```
2022/04/17 13:05:00 Working chain ID: 1
2022/04/17 13:05:00 Fetching tokens...
2022/04/17 13:05:02 Fetching balances...
2022/04/17 13:05:02 ADDRESS: 0xC730155a5F702B6bF3CaeCD25ebff9fB1b2a2B85
2022/04/17 13:05:02 - ETH balance: 100.00
2022/04/17 13:05:02 - LINK balance: 200.50
2022/04/17 13:05:02 - USDC balance: 3450.00
2022/04/17 13:05:02 - USDT balance: 0
2022/04/17 13:05:02 Watching for transactions...
```
When a new transaction is detected, you will see more log entires such as "You received..." or "You sent...".
To terminate the app, just hit `Ctrl+C`.

## Telegram integration
Telegram bot supports two commans: `/subscribe` and `/unsubscribe`.
The first command will enable bot's notifications and the second command will stop notifications.
The `username` specified in `config.yaml` will restrict other users to see your notifications and/or subscribe/unsubscribe.
