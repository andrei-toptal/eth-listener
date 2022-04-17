# eth-listener
A simple console app listening to your ETH transactions

## Configuration
All configuration is done by editing `config.yaml` file.
1. Specify your ETH node URL. If you don't have one, use 3rd-party providers such as [alchemy](https://alchemy.com/?r=62491cd8ac883927) to create one.
2. Enter your ETH account addresses you want to watch (there may be many addresses).
3. Enter your token contract addresses you want to watch (there may be many contract addresses).

NOTE: if you don't know your contract address for your chain, just google it, e.g. "link contract address on kovan".

## Using
After you edited `config.yaml`, just start the app by `go run .`
You will see the output like this:
```
2022/04/17 13:05:00 Working chain ID: 1
2022/04/17 13:05:00 Fetching tokens...
2022/04/17 13:05:02 Fetching balances...
2022/04/17 13:05:02 ADDRESS: 0xC730155a5F702B6bF3CaeCD25ebff9fB1b2a2B85
2022/04/17 13:05:02 - ETH balance: 0.001106166871
2022/04/17 13:05:02 - LINK balance: 0
2022/04/17 13:05:02 - USDC balance: 0
2022/04/17 13:05:02 - USDT balance: 0
2022/04/17 13:05:02 Watching for transactions...
```
When a new transaction is detected, you will see more log entires such as "You received..." or "You sent...".
To terminate the app, just hit `Ctrl+C`.
