# WebSocket Stream Benchmark

## Why this project

What I am interested in is if the WebSocket server sends message to the client frequently, does it affect the client's requests? Should I use multiple WebSocket connections to split the data traffic?

So I used Golang to do the benchmarking for this scenario. Finally, let's share this result with you.

## Benchmark Goals

When the server keeps sending data to the client at a high frequency over a WebSocket connection, how long does it take for the client to receive a reply if it sends a low frequency message to the server over the same WebSocket connection?

## How to use this project

Clone this project and build

```sh
git clone https://github.com/weirenxue/websocket-stream-benchmarck.git # clone

make # build
```

You will find two executable files `client` and `server` in the `./bin` folder. Start by running `./bin/server`, then open another terminal window and run `./bin/client` and wait until `client` finishes executing and outputs the result.

You may want to change the configuration in `config.toml` to see the test results for different scenarios.

## Parameters in config.toml

- `server`
  - `host`: WebSocket server host.
  - `port`: WebSocket server port.
  - `dummy_message_size`: The size for each dummy message.
  - `dummy_message_duration`: The time interval to deliver each dummy message.
- `client`
  - `request_duration`: The time interval to send each request.
  - `total_request`: The number of requests that should be sent.

## Design concept

- Server: As soon as the client connects to the server, the server sends a message of size `dummy_message_size` to the client every `dummy_message_duration`. The server listens to see if the client is sending a message, and if so, it will receive and reply directly to the original message.
- Client: The client generates a `uuid` as a message body to the server, and the server must send back the same message. Based on the uniqueness of `uuid`, we can know the time it takes for a request to go from being sent to being answered, and finally divide the time required for each request by the number of requests to get the average round-trip time.

**Finally, each case is tested five times, and the worst case is selected and recorded in the experimental results.**

## Test environment and results

- OS: macOS 12.2.1
- CPU: i7 2.7GHz 4 Cores
- RAM: 16 GB LPDDR3

<!-- markdownlint-disable MD033 -->

dummy config→<br>request config↓|dummy_message_size="10KB"<br>dummy_message_duration="1h"<br>(no dummy message due to long duration)|dummy_message_size="10KB"<br>dummy_message_duration="500us"|dummy_message_size="1MB"<br>dummy_message_duration="500us"
:---|:---|:---|:---
total_request=1000<br>request_duration="1ms"|total duration: 160.714504ms<br>average duration: **160.714µs**<br>total received dummy message: 0B|total duration: 83.472281ms<br>average duration: **83.472µs**<br>total received dummy message: 19710KB<br>|total duration: 5.364239748s<br>average duration: **5.364239ms**<br>total received dummy message: 787MB<br>
request_duration="10ms"<br>total_request=1000|total duration: 300.998294ms<br>average duration: **300.998µs**<br>total received dummy message: 0B<br>|total duration: 110.638204ms<br>average duration: **110.638µs**<br>total received dummy message: 154350KB<br>|total duration: 5.675233495s<br>average duration: **5.675233ms**<br>total received dummy message: 7520MB<br>

<!-- markdownlint-restore MD033 -->

## Conclusion

We may be afraid that if all the data traffic is on the same WebSocket connection, the bandwidth will be split and some client-side commands will not be processed in time. But according to the results of this project, we can safely let all data traffic pass over one WebSocket connection. This is a good thing, because multiple connections would put some burden on the server.

In addition, the WebSocket connection count issue can be seen in commercial products. For example,in the [official pricing for Firebase's Realtime Database service](https://firebase.google.com/pricing#realtime-database) includes a limit on the number of connections that can be used at the same time, so if a user has multiple connections on a browser window it will be a costly burden. Of course, the Firebase SDK is considerate enough to allow all requests to be made on the same WebSocket, ensuring that there is only one WebSocket connection in a browser window.
