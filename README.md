A weekend hack: listen to Coinbase’s live BTC order-book and print the nanosecond whenever the best bid sits on top of the ask (a cross).

Still in developement. Expect messy logs and the odd crash.

Right now it’s just a single Go file and a dream.

Run it:

go run .

You’ll see something like:

cross 1701432000000000000 42150.00 42149.99

Leave the terminal open and go touch grass. If the market hiccups you’ll catch it.
