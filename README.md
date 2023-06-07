# Livekit-CLI

Livekit-CLI is a tool designed to perform various types of video load testing. This utility provides several functionalities such as token generation, room connection, and more. This document focuses on the subcommands for load testing functionalities.

Before using the utility, you need to build it. In order to do this, you need to have Golang installed on your system. If you already have Golang installed, you can skip the Golang installation step and proceed to the build step.

## Install Go
### Install Go on MacOS via Homebrew

1. Open a terminal.
2. Check if Homebrew is installed by typing: 
```bash
brew --version
```
If Homebrew is not installed, you can install it by running:
```bash
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
```
3. Once Homebrew is installed, you can install Go by typing: 
```bash
brew install go
```
4. To verify your installation, run: 
```bash
go version
```
This command should display the version of Go that was installed.

### Install Go on Debian Linux via apt

1. Open a terminal.
2. Update your package list by running: 
```bash
sudo apt update
```
3. Install Go by typing: 
```bash
sudo apt install golang-go
```
4. To verify your installation, run: 
```bash
go version
```
This command should display the version of Go that was installed.

### Install Go Manually from the Official Website (For MacOS or Debian)

1. Visit the Go Downloads page at [https://golang.org/dl/](https://golang.org/dl/).
2. Download the distribution appropriate for your system. For MacOS, this will be a .pkg file; for Debian Linux, this will be a .tar.gz file.
3. For MacOS:
   - Open the downloaded .pkg file and follow the prompts to install Go.
   - Verify your installation by opening a terminal and running: 
   ```bash
   go version
   ```
4. For Debian Linux:
   - Open a terminal and navigate to the directory where the .tar.gz file was downloaded.
   - Extract the archive using the command 
   ```bash
   tar -C /usr/local -xzf go$VERSION.$OS-$ARCH.tar.gz
   ```
   - Add /usr/local/go/bin to the PATH environment variable by adding the following line to your /etc/profile (for a system-wide installation) or $HOME/.profile:
   ```bash
   export PATH=$PATH:/usr/local/go/bin
   ```
   - Verify your installation by opening a new terminal session and running: 
   ```bash
   go version
   ```

**Note:** Make sure you have the necessary privileges to install software on your machine.

## Build livekit-cli
This repo uses [Git LFS](https://git-lfs.github.com/) for embedded video resources. Please ensure git-lfs is installed on your machine.
```bash
git clone path
make install
```

**Note:**, if you're building under MacOS Apple Silicon, be careful as you might encounter an error. This error may be related to the command being run under Intel. To ensure that the utility was built, check the bin/ directory. If there's a file in it, then the build was successful.

You can also manually build without `make`.
```bash
go build -o bin/livekit-cli ./cmd/livekit-cli
```

## Subcommands

- `room-name`: Specifies the prefix of the room from which the room names will be created as room-name + publisher number.
- `start-publisher` and `end-publisher`: These define the number of publishers who will be streaming to different rooms. The room name is determined by the  `room-name` option, and the rooms will then have names like `room-name` + publisher number. It's also possible to specify publishers distributed by indicating the same room-name and start-publisher and `--end-publisher`, considering the previous numbers of publishers. This allows for testing a large number of publishers from different machines on a single server.
- `start-room-number` and `end-room-number`: These enable the capability to connect to remote rooms. The room name is determined by the `room-name` option. It's also possible to distribute connections from different machines by specifying `start-room-number` and `end-room-number`.
- `subscribers`: Specifies the number of viewers for each publisher. This command depends on the `video-publishers` command.
- `video-resolution`: Specifies the resolution of the video streamed by the publisher. In this option, the resolution is separated by space and indicated for each publisher specified in the `video-publishers` command.
- `no-simulcast`: Indicates that the publisher streams without Simulcasting and in high resolution.
- `duration`: Specifies the test duration.
- `video-codec`: Specifies the video codec used by the video publisher.
- `high`, `medium`, `low`: If the `no-simulcast` option is not selected, it specifies the resolution at which the subscriber will consume the video. These parameters depend on the `subscribers` parameter. With the `high` option, we specify how many subscribers will consume the video in high resolution, etc.
- `data-publishers`: Specifies the number of publishers for the data channel.
- `data-packet-bytes`, `data-bitrate`: These parameters specify the size of the data packet and how many of these packets will be sent per second.
- `with-audio`: Indicates that the publisher will stream with audio.
- `same-room`: Indicates that the all publishers and subscribers will be in the same room.

Currently, the following resolution formats are supported: 1440p, 1080p, 720p, 360p. We support the following resolution table with bitrate for these formats:

- 1440p
    - High - Width: 2560, Height: 1440, Bitrate: 7300
    - Medium - Width: 2048, Height: 1152, Bitrate 4800
    - Low - Width: 1024, Height: 576, Bitrate 1200
- 1080p
    - High - Width: 1920, Height: 1080, Bitrate: 4100
    - Medium - Width: 800, Height: 450, Bitrate: 720
    - Low - Width: 640, Height: 360, Bitrate 460
- 720p
    - High - Width: 1280, Height: 720, Bitrate: 1800
    - Medium - Width: 800, Height: 450, Bitrate: 720
    - Low - Width: 640, Height: 360, Bitrate 460
- 360p 
    - High - Width: 640, Height: 360, Bitrate 460
    - Medium - Width: 640, Height: 360, Bitrate 460
    - Low - Width: 640, Height: 360, Bitrate 460

By default, 1080p is supported.

The video codec currently supported for video streaming is h264.

Connection parameters to the Livekit server should also be provided, preferably via environment variables:

```bash
export LIVEKIT_URL=
export LIVEKIT_API_KEY=
export LIVEKIT_API_SECRET=
```

### Launch Examples:

#### 1. Launch with a single publisher in 1080p resolution and two subscribers with a 1-minute stream interval without simulcasting in room with prefix `VM1`:
```shell
./livekit-cli load-test --duration 1m --video-codec h264 --video-resolution "1080p" --no-simulcast --room-name VM1   --start-publisher 1  --end-publisher 1 --subscribers 2

Statistics for room VM1_1

Sub 0 in VM1_1 | Track             | Kind  | Pkts  | Bitrate | Latency    | Dropped
               | TR_VC5kYBccKKTiyr | video | 26593 | 4.1mbps | 7.284755ms | 0 (0%)

Sub 1 in VM1_1 | Track             | Kind  | Pkts  | Bitrate | Latency    | Dropped
               | TR_VC5kYBccKKTiyr | video | 26593 | 4.1mbps | 7.288989ms | 0 (0%)

Summary for room VM1_1

Summary | Tester         | Kind  | Tracks | Bitrate               | Latency    | Total Dropped | Error
        | Sub 0 in VM1_1 | video | 1      | 4.1mbps               | 7.284755ms | 0 (0%)        | -
        | Sub 1 in VM1_1 | video | 1      | 4.1mbps               | 7.288989ms | 0 (0%)        | -
        | Total          | video | 2      | 8.3mbps (4.1mbps avg) | 7.286872ms | 0 (0%)        | 0
```

#### 2. Launch with two publishers in 1080p and 720p resolutions and two subscribers for each publisher with a 1-minute stream interval without simulcasting in room with prefix `VM1`:
```shell
./livekit-cli load-test --duration 1m --video-codec h264 --video-resolution "1080p" --no-simulcast --room-name VM1   --start-publisher 1  --end-publisher 2 --subscribers 2

Statistics for room VM1_1

Sub 0 in VM1_1 | Track             | Kind  | Pkts  | Bitrate | Latency    | Dropped
               | TR_VCJeWm4HeybYY3 | video | 27115 | 4.1mbps | 6.172919ms | 0 (0%)

Sub 1 in VM1_1 | Track             | Kind  | Pkts  | Bitrate | Latency    | Dropped
               | TR_VCJeWm4HeybYY3 | video | 27115 | 4.1mbps | 6.184611ms | 0 (0%)

Statistics for room VM1_2

Sub 0 in VM1_2 | Track             | Kind  | Pkts  | Bitrate | Latency    | Dropped
               | TR_VCBXTX9fgKSkmm | video | 27132 | 4.1mbps | 8.388166ms | 0 (0%)

Sub 1 in VM1_2 | Track             | Kind  | Pkts  | Bitrate | Latency    | Dropped
               | TR_VCBXTX9fgKSkmm | video | 27132 | 4.1mbps | 8.385111ms | 0 (0%)

Summary for room VM1_2

Summary | Tester         | Kind  | Tracks | Bitrate               | Latency    | Total Dropped | Error
        | Sub 0 in VM1_2 | video | 1      | 4.1mbps               | 8.388166ms | 0 (0%)        | -
        | Sub 1 in VM1_2 | video | 1      | 4.1mbps               | 8.385111ms | 0 (0%)        | -
        | Total          | video | 2      | 8.3mbps (4.1mbps avg) | 8.386639ms | 0 (0%)        | 0

Summary for room VM1_1

Summary | Tester         | Kind  | Tracks | Bitrate               | Latency    | Total Dropped | Error
        | Sub 0 in VM1_1 | video | 1      | 4.1mbps               | 6.172919ms | 0 (0%)        | -
        | Sub 1 in VM1_1 | video | 1      | 4.1mbps               | 6.184611ms | 0 (0%)        | -
        | Total          | video | 2      | 8.3mbps (4.1mbps avg) | 6.178765ms | 0 (0%)        | 0
```

#### 3. Launch with two publishers in 1080p and 720p resolutions and three subscribers for each publisher with a 1-minute stream interval with simulcast, where the first subscriber uses high resolution, the second in medium, and the third in low:
```shell
./livekit-cli load-test --duration 1m --video-codec h264 --high=1 --medium=1 --low=1 --video-resolution "1080p 720p" --start-publisher 1  --end-publisher 2 --subscribers 3
Statistics for room load-test_1

Sub 0 in load-test_1 | Track             | Kind  | Pkts  | Bitrate | Latency    | Dropped
                     | TR_VCPT3wHs6sAWvB | video | 27115 | 4.1mbps | 6.908075ms | 0 (0%)

Sub 1 in load-test_1 | Track             | Kind  | Pkts | Bitrate   | Latency    | Dropped
                     | TR_VCPT3wHs6sAWvB | video | 5864 | 792.4kbps | 6.832594ms | 0 (0%)

Sub 2 in load-test_1 | Track             | Kind  | Pkts | Bitrate   | Latency    | Dropped
                     | TR_VCPT3wHs6sAWvB | video | 4095 | 529.3kbps | 6.784838ms | 0 (0%)

Statistics for room load-test_2

Sub 0 in load-test_2 | Track             | Kind  | Pkts  | Bitrate | Latency    | Dropped
                     | TR_VCXFqLsFZUAEGN | video | 11748 | 1.7mbps | 6.699061ms | 0 (0%)

Sub 1 in load-test_2 | Track             | Kind  | Pkts | Bitrate   | Latency   | Dropped
                     | TR_VCXFqLsFZUAEGN | video | 5518 | 740.6kbps | 6.70892ms | 0 (0%)

Sub 2 in load-test_2 | Track             | Kind  | Pkts | Bitrate   | Latency    | Dropped
                     | TR_VCXFqLsFZUAEGN | video | 3755 | 477.6kbps | 6.655672ms | 0 (0%)

Summary for room load-test_1

Summary | Tester               | Kind  | Tracks | Bitrate               | Latency    | Total Dropped | Error
        | Sub 0 in load-test_1 | video | 1      | 4.1mbps               | 6.908075ms | 0 (0%)        | -
        | Sub 1 in load-test_1 | video | 1      | 792.4kbps             | 6.832594ms | 0 (0%)        | -
        | Sub 2 in load-test_1 | video | 1      | 529.3kbps             | 6.784838ms | 0 (0%)        | -
        | Total                | video | 3      | 5.5mbps (1.8mbps avg) | 6.841247ms | 0 (0%)        | 0

Summary for room load-test_2

Summary | Tester               | Kind  | Tracks | Bitrate                 | Latency    | Total Dropped | Error
        | Sub 0 in load-test_2 | video | 1      | 1.7mbps                 | 6.699061ms | 0 (0%)        | -
        | Sub 1 in load-test_2 | video | 1      | 740.6kbps               | 6.70892ms  | 0 (0%)        | -
        | Sub 2 in load-test_2 | video | 1      | 477.6kbps               | 6.655672ms | 0 (0%)        | -
        | Total                | video | 3      | 3.0mbps (987.5kbps avg) | 6.687669ms | 0 (0%)        | 0
```

#### 4. Launch with a single publisher in 1440p resolution, with two data publishers and two subscribers for each publisher with a 1-minute stream interval without simulcasting:
```shell
./livekit-cli load-test --duration 1m --video-codec h264 --video-resolution "1440p" --data-publishers 2 --no-simulcast --start-publisher 1  --end-publisher 1 --subscribers 2

Statistics for room load-test_1

Sub 0 in load-test_1 | Track             | Kind  | Pkts  | Bitrate | Latency    | Dropped
                     | PA_RZQnhwddGtp8   | data  | 7681  | 1.1mbps | 1.269959ms | 0 (0%)
                     | TR_VCWjNe6EiscvXn | video | 46752 | 7.3mbps | 7.347317ms | 0 (0%)

Sub 1 in load-test_1 | Track             | Kind  | Pkts  | Bitrate | Latency    | Dropped
                     | TR_VCWjNe6EiscvXn | video | 46752 | 7.3mbps | 7.339957ms | 0 (0%)
                     | PA_49NB4rjnoUtT   | data  | 7681  | 1.1mbps | 1.276926ms | 0 (0%)

Summary for room load-test_1

Summary | Tester               | Kind  | Tracks | Bitrate                | Latency    | Total Dropped | Error
        | Sub 0 in load-test_1 | video | 1      | 7.3mbps                | 7.347317ms | 0 (0%)        | -
        | Sub 0 in load-test_1 | data  | 1      | 1.1mbps                | 1.269959ms | 0 (0%)        | -
        | Sub 1 in load-test_1 | video | 1      | 7.3mbps                | 7.339957ms | 0 (0%)        | -
        | Sub 1 in load-test_1 | data  | 1      | 1.1mbps                | 1.276926ms | 0 (0%)        | -
        | Total                | video | 2      | 14.7mbps (7.3mbps avg) | 7.343637ms | 0 (0%)        | 0
        | Total                | data  | 2      | 2.1mbps (1.1mbps avg)  | 1.273443ms | 0 (0%)        | 0
```

#### 5. Launching a remote connection with two publishers and two subscribers on different machines. In this case, a connection to one room will be made from one machine, and to the other from a different machine.

##### Machine 1: Running 2 publishers without subscribers

 ```shell
./livekit-cli load-test --duration 1m --video-codec h264 --video-resolution "1080p" --no-simulcast --room-name VM1  --end-publisher 2
Using url, api-key, api-secret from environment
Starting load test with 2 video publishers
publishing video track - fylyz_pubVM1_1_0
publishing video track - fylyz_pubVM1_2_1
Finished connecting to room, waiting 1m0s
No subscribers, skipping stats
 ```

##### Machine 2: Running 2 subscribers to connect to one room on Machine 1

 ```shell
/livekit-cli load-test --duration 1m --video-codec h264 --video-resolution "1080p" --no-simulcast --room-name VM1 --start-room-number 1  --end-room-number 1 --subscribers 2
Statistics for room VM1_1

Sub 0 in VM1_1 | Track             | Kind  | Pkts  | Bitrate | Latency    | Dropped
               | TR_VCRNowWqxAtsE3 | video | 26086 | 4.0mbps | 6.598817ms | 0 (0%)

Sub 1 in VM1_1 | Track             | Kind  | Pkts  | Bitrate | Latency    | Dropped
               | TR_VCRNowWqxAtsE3 | video | 26086 | 4.0mbps | 6.592381ms | 0 (0%)

Summary for room VM1_1

Summary | Tester         | Kind  | Tracks | Bitrate               | Latency    | Total Dropped | Error
        | Sub 0 in VM1_1 | video | 1      | 4.0mbps               | 6.598817ms | 0 (0%)        | -
        | Sub 1 in VM1_1 | video | 1      | 4.0mbps               | 6.592381ms | 0 (0%)        | -
        | Total          | video | 2      | 8.1mbps (4.0mbps avg) | 6.595599ms | 0 (0%)        | 0
```

##### Machine 3: Running 2 subscribers to connect to another room on Machine 1

 ```shell
 ./livekit-cli load-test --duration 1m --video-codec h264 --video-resolution "1080p" --no-simulcast --room-name VM1 --start-room-number 2  --end-room-number 2 --subscribers 2
Statistics for room VM1_2

Sub 0 in VM1_2 | Track             | Kind  | Pkts  | Bitrate | Latency    | Dropped
               | TR_VCfZxVXnPzTVDu | video | 25487 | 3.9mbps | 7.366207ms | 0 (0%)

Sub 1 in VM1_2 | Track             | Kind  | Pkts  | Bitrate | Latency    | Dropped
               | TR_VCfZxVXnPzTVDu | video | 25487 | 3.9mbps | 7.372518ms | 0 (0%)

Summary for room VM1_2

Summary | Tester         | Kind  | Tracks | Bitrate               | Latency    | Total Dropped | Error
        | Sub 0 in VM1_2 | video | 1      | 3.9mbps               | 7.366207ms | 0 (0%)        | -
        | Sub 1 in VM1_2 | video | 1      | 3.9mbps               | 7.372518ms | 0 (0%)        | -
        | Total          | video | 2      | 7.8mbps (3.9mbps avg) | 7.369363ms | 0 (0%)        | 0
```

#### 6. Launch with two publishers in 1080p resolution and two subscribers for each publisher with a 1-minute stream interval without simulcasting in the same room `VM1`.
```shell
./livekit-cli load-test --duration 1m --video-codec h264 --video-resolution "1080p" --no-simulcast --room-name VM1 --same-room  --end-publisher 2 --subscribers 2

Statistics for room VM1

Sub 0 in VM1 | Track             | Kind  | Pkts  | Bitrate | Latency    | Dropped
             | TR_VCJTFJgfWPrLPM | video | 26663 | 4.1mbps | 7.741564ms | 0 (0%)
             | TR_VC46w4STDFW92A | video | 26533 | 4.1mbps | 7.413243ms | 0 (0%)

Sub 1 in VM1 | Track             | Kind  | Pkts  | Bitrate | Latency    | Dropped
             | TR_VCJTFJgfWPrLPM | video | 26663 | 4.1mbps | 7.725936ms | 0 (0%)
             | TR_VC46w4STDFW92A | video | 26533 | 4.1mbps | 7.407041ms | 0 (0%)

Summary for room VM1

Summary | Tester       | Kind  | Tracks | Bitrate                | Latency    | Total Dropped | Error
        | Sub 0 in VM1 | video | 2      | 8.3mbps                | 7.577821ms | 0 (0%)        | -
        | Sub 1 in VM1 | video | 2      | 8.3mbps                | 7.566893ms | 0 (0%)        | -
        | Total        | video | 4      | 16.6mbps (4.1mbps avg) | 7.572357ms | 0 (0%)        | 0
```

#### 7. Launch with two publishers with audio in 1080p resolution and two subscribers for each publisher with a 1-minute stream interval.
```shell
./livekit-cli load-test --duration 1m --video-codec h264 --video-resolution "1080p" --no-simulcast --room-name VM1 --with-audio  --end-publisher 2 --subscribers 2
Statistics for room VM1_1

Sub 0 in VM1_1 | Track             | Kind  | Pkts  | Bitrate  | Latency    | Dropped
               | TR_VCb3jqPexpMsb4 | video | 27132 | 4.1mbps  | 6.595114ms | 0 (0%)
               | TR_AMUyfhHwtrUQMf | audio | 3081  | 23.3kbps | 6.72431ms  | 0 (0%)

Sub 1 in VM1_1 | Track             | Kind  | Pkts  | Bitrate  | Latency    | Dropped
               | TR_AMUyfhHwtrUQMf | audio | 3081  | 23.3kbps | 6.71336ms  | 0 (0%)
               | TR_VCb3jqPexpMsb4 | video | 27132 | 4.1mbps  | 6.591647ms | 0 (0%)

Statistics for room VM1_2

Sub 0 in VM1_2 | Track             | Kind  | Pkts  | Bitrate  | Latency    | Dropped
               | TR_VCiPTnBK3Sr52j | video | 27067 | 4.2mbps  | 7.498247ms | 0 (0%)
               | TR_AMNj7cZdraQz6D | audio | 3071  | 23.3kbps | 6.792933ms | 0 (0%)

Sub 1 in VM1_2 | Track             | Kind  | Pkts  | Bitrate  | Latency    | Dropped
               | TR_AMNj7cZdraQz6D | audio | 3071  | 23.3kbps | 6.79852ms  | 0 (0%)
               | TR_VCiPTnBK3Sr52j | video | 27067 | 4.2mbps  | 7.508409ms | 0 (0%)

Summary for room VM1_1

Summary | Tester         | Kind  | Tracks | Bitrate                 | Latency    | Total Dropped | Error
        | Sub 0 in VM1_1 | video | 1      | 4.1mbps                 | 6.595114ms | 0 (0%)        | -
        | Sub 0 in VM1_1 | audio | 1      | 23.3kbps                | 6.72431ms  | 0 (0%)        | -
        | Sub 1 in VM1_1 | video | 1      | 4.1mbps                 | 6.591647ms | 0 (0%)        | -
        | Sub 1 in VM1_1 | audio | 1      | 23.3kbps                | 6.71336ms  | 0 (0%)        | -
        | Total          | video | 2      | 8.3mbps (4.1mbps avg)   | 6.59338ms  | 0 (0%)        | 0
        | Total          | audio | 2      | 46.6kbps (23.3kbps avg) | 6.718835ms | 0 (0%)        | 0

Summary for room VM1_2

Summary | Tester         | Kind  | Tracks | Bitrate                 | Latency    | Total Dropped | Error
        | Sub 0 in VM1_2 | video | 1      | 4.2mbps                 | 7.498247ms | 0 (0%)        | -
        | Sub 0 in VM1_2 | audio | 1      | 23.3kbps                | 6.792933ms | 0 (0%)        | -
        | Sub 1 in VM1_2 | video | 1      | 4.2mbps                 | 7.508409ms | 0 (0%)        | -
        | Sub 1 in VM1_2 | audio | 1      | 23.3kbps                | 6.79852ms  | 0 (0%)        | -
        | Total          | video | 2      | 8.3mbps (4.2mbps avg)   | 7.503328ms | 0 (0%)        | 0
        | Total          | audio | 2      | 46.6kbps (23.3kbps avg) | 6.795727ms | 0 (0%)        | 0
```
