This little Go program listens to Coinbase’s live order-book feed and prints the exact nanosecond whenever the best bid price is higher or equal to the best ask (a cross).

How to run:

go run .

You’ll see something like:

cross 1701432000000000000 42150.00 42149.99
