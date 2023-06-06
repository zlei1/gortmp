# gortmp

```
连接

        +--------------+                +-------------+
        | Client       |        |       |    Server   |
        +------+-------+        |       +------+------+
        |                Handshaking done             |
        |                       |                     |
        |                       |                     |
        |----------- Command Message(connect) ------->|
        |                                             |
        |<------- Window Acknowledgement Size --------|
        |                                             |
        |<----------- Set Peer Bandwidth -------------|
        |                                             |
        |-------- Window Acknowledgement Size ------->|
        |                                             |
        |<------ User Control Message(StreamBegin) ---|
        |                                             |
        |<------------ Command Message ---------------|
        |       (_result- connect response)           |
```

```
传输流

            +--------------------+        +-----------+
            | Publisher Client |     |       | Server |
            +----------+---------+   |    +-----+-----+
                |            Handshaking Done         |
                |                    |                |
                |                    |                |
        ---+----|-----  Command Message(connect)----->|
                |                    |                |
              | |<----- Window Acknowledge Size ------|
     Connect  | |                                     |
              | |<-------Set Peer BandWidth ----------|
              | |                                     |
              | |------ Window Acknowledge Size ----->|
              | |                                     |
              | |<------User Control(StreamBegin)-----|
              | |                                     |
       ---+---- |<---------Command Message -----------|
                |     (_result- connect response)     |
                |                                     |
       ---+---- |--- Command Message(createStream)--->|
       Create | |                                     |
       Stream | |                                     |
       ---+---- |<------- Command Message ------------|
                |   (_result- createStream response)  |
                |                                     |
       ---+---- |---- Command Message(publish) ------>|
              | |                                     |
              | |<------User Control(StreamBegin)-----|
              | |                                     |
              | |-----Data Message (Metadata)-------->|
              | |                                     |
    Publishing| |------------ Audio Data ------------>|
      Content | |                                     |
              | |------------ SetChunkSize ---------->|
              | |                                     |
              | |<----------Command Message ----------|
              | |      (_result- publish result)      |
              | |                                     |
              | |------------- Video Data ----------->|
              | |                                     |
              | |                                     |
              | |    Until the stream is complete     |
              | |                                     |
             Message flow in publishing a video stream| 
```

## 启动
```
go build -o gortmp ./cmd/
./gortmp
```

## 推流
```
ffmpeg -re -i demo.flv -c copy -f flv rtmp://localhost:1935/live/live_stream
```

## 拉流地址（可使用 VLC 播放器播放）
```
rtmp://localhost:1935/live/live_stream
```