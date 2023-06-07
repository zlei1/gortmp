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
- RTMP协议是基于TCP的，在TCP连接建立以后，进行RTMP协议层次的握手
- 客户端发送connect命令到服务器，请求与服务端的application进行连接
- 服务端收到connect命令后，服务器会发送协议消息“Window Acknowledgement size”消息到客户端。服务端同时连接到connect中请求的application
- 服务端发送协议消息“Set Peer BandWidth”到客户端
- 客户端在处理完服务端发来的“Set Peer BandWidth”消息后，向服务端发送“Window Acknowledgment”消息
- 服务端向客户端发送一条用户控制消息（Stream Begin）
- 如果连接成功，服务端向客户端发送_result消息，否则发送_error消息

```
传输流

            +--------------------+        +-----------+
            | Publisher Client |     |       | Server |
            +----------+---------+   |    +-----+-----+
                |    Handshaking and Application      |
                |            connect done             |
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
   Publishing | |------------ Audio Data ------------>|
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
- 客户端发送createStream请求命令消息，请求服务端创建一条流
- 服务端发送createStream响应命令消息，返回NetConnection的流 ID
- 客户端发送Publish请求命令消息，请求发布内容信息
- 服务端发送带有StreamBegin事件的用户控制消息，通知客户端指定流已经准备就绪可以用来通信
- 客户端发送元数据消息
- 客户端发送音频数据
- 客户端设置块大小
- 服务端响应Publish状态
- 客户端发送视频数据

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

## 代码主要来源

- [joy5](https://github.com/nareix/joy5) 
