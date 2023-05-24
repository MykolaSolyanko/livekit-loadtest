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

- `video-publishers`: Specifies the number of video streamers.
- `subscribers`: Specifies the number of viewers for each publisher. This command depends on the `video-publishers` command.
- `video-resolution`: Specifies the resolution of the video streamed by the publisher. In this option, the resolution is separated by space and indicated for each publisher specified in the `video-publishers` command.
- `no-simulcast`: Indicates that the publisher streams without Simulcasting and in high resolution.
- `duration`: Specifies the test duration.
- `video-codec`: Specifies the video codec used by the video publisher.
- `high`, `medium`, `low`: If the `no-simulcast` option is not selected, it specifies the resolution at which the subscriber will consume the video. These parameters depend on the `subscribers` parameter. With the `high` option, we specify how many subscribers will consume the video in high resolution, etc.
- `data-publishers`: Specifies the number of publishers for the data channel.
- `data-packet-bytes`, `data-bitrate`: These parameters specify the size of the data packet and how many of these packets will be sent per second.

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

#### 1. Launch with a single publisher in 1080p resolution and two subscribers with a 1-minute stream interval without simulcasting:
```shell
./livekit-cli load-test --duration 1m --video-codec h264 --video-resolution "1080p" --no-simulcast --video-publishers 1 --subscribers 2

Statistics for room load-test_0

Sub 0 in load-test_0 | Track                | Kind  | Pkts  | Bitrate | Latency     | Dropped | Data Pkts | Data Bitrate | Latency
                     | 0V TR_VCZk3mpLDYZ6QG | video | 26245 | 4.1mbps | 38.245066ms | 0 (0%)  | 0         | 0bps         |  -

Sub 1 in load-test_0 | Track                | Kind  | Pkts  | Bitrate | Latency     | Dropped | Data Pkts | Data Bitrate | Latency
                     | 0V TR_VCZk3mpLDYZ6QG | video | 26245 | 4.1mbps | 38.181305ms | 0 (0%)  | 0         | 0bps         |  -

Summary for room load-test_0

Summary | Tester               | Bitrate               | Latency     | Total Dropped | Data Bitrate    | Latency | Error
        | Sub 0 in load-test_0 | 4.1mbps               | 38.245066ms | 0 (0%)        | 0bps            |  -      | -
        | Sub 1 in load-test_0 | 4.1mbps               | 38.181305ms | 0 (0%)        | 0bps            |  -      | -
        | Total                | 8.3mbps (8.3mbps avg) | 38.213186ms | 0 (0%)        | 0bps (0bps avg) |  -      | 0
```

#### 2. Launch with two publishers in 1080p and 720p resolutions and two subscribers for each publisher with a 1-minute stream interval without simulcasting:
```shell
./livekit-cli load-test --duration 1m --video-codec h264 --video-resolution "1080p 720p" --no-simulcast --video-publishers 2 --subscribers 2

Statistics for room load-test_0

Sub 1 in load-test_0 | Track                | Kind  | Pkts  | Bitrate | Latency     | Dropped | Data Pkts | Data Bitrate | Latency
                     | 0V TR_VCmzea9vSy82Zr | video | 26533 | 4.1mbps | 37.175218ms | 0 (0%)  | 0         | 0bps         |  -

Sub 0 in load-test_0 | Track                | Kind  | Pkts  | Bitrate | Latency     | Dropped | Data Pkts | Data Bitrate | Latency
                     | 0V TR_VCmzea9vSy82Zr | video | 26533 | 4.1mbps | 37.955722ms | 0 (0%)  | 0         | 0bps         |  -

Statistics for room load-test_1

Sub 0 in load-test_1 | Track                | Kind  | Pkts  | Bitrate | Latency     | Dropped | Data Pkts | Data Bitrate | Latency
                     | 1V TR_VCM6fYqsEJSxE6 | video | 11530 | 1.7mbps | 38.003357ms | 0 (0%)  | 0         | 0bps         |  -

Sub 1 in load-test_1 | Track                | Kind  | Pkts  | Bitrate | Latency     | Dropped | Data Pkts | Data Bitrate | Latency
                     | 1V TR_VCM6fYqsEJSxE6 | video | 11530 | 1.7mbps | 37.166911ms | 0 (0%)  | 0         | 0bps         |  -

Summary for room load-test_0

Summary | Tester               | Bitrate               | Latency     | Total Dropped | Data Bitrate    | Latency | Error
        | Sub 1 in load-test_0 | 4.1mbps               | 37.175218ms | 0 (0%)        | 0bps            |  -      | -
        | Sub 0 in load-test_0 | 4.1mbps               | 37.955722ms | 0 (0%)        | 0bps            |  -      | -
        | Total                | 8.3mbps (4.1mbps avg) | 37.56547ms  | 0 (0%)        | 0bps (0bps avg) |  -      | 0

Summary for room load-test_1

Summary | Tester               | Bitrate               | Latency     | Total Dropped | Data Bitrate    | Latency | Error
        | Sub 0 in load-test_1 | 1.7mbps               | 38.003357ms | 0 (0%)        | 0bps            |  -      | -
        | Sub 1 in load-test_1 | 1.7mbps               | 37.166911ms | 0 (0%)        | 0bps            |  -      | -
        | Total                | 3.5mbps (1.7mbps avg) | 37.585134ms | 0 (0%)        | 0bps (0bps avg) |  -      | 0
```

#### 3. Launch with two publishers in 1080p and 720p resolutions and three subscribers for each publisher with a 1-minute stream interval with simulcast, where the first subscriber uses high resolution, the second in medium, and the third in low:
```shell
./livekit-cli load-test --duration 1m --video-codec h264 --high=1 --medium=1 --low=1 --video-resolution "1080p 720p" --video-publishers 2 --subscribers 3

Statistics for room load-test_0

Sub 0 in load-test_0 | Track                | Kind  | Pkts  | Bitrate | Latency    | Dropped | Data Pkts | Data Bitrate | Latency
                     | 0V TR_VCmLbkoPbcotGd | video | 26538 | 4.1mbps | 37.33238ms | 0 (0%)  | 0         | 0bps         |  -

Sub 1 in load-test_0 | Track                | Kind  | Pkts | Bitrate   | Latency     | Dropped | Data Pkts | Data Bitrate | Latency
                     | 0V TR_VCmLbkoPbcotGd | video | 5757 | 795.0kbps | 36.941458ms | 0 (0%)  | 0         | 0bps         |  -

Sub 2 in load-test_0 | Track                | Kind  | Pkts | Bitrate   | Latency     | Dropped | Data Pkts | Data Bitrate | Latency
                     | 0V TR_VCmLbkoPbcotGd | video | 4026 | 531.8kbps | 37.663974ms | 0 (0%)  | 0         | 0bps         |  -

Statistics for room load-test_1

Sub 0 in load-test_1 | Track                | Kind  | Pkts  | Bitrate | Latency     | Dropped | Data Pkts | Data Bitrate | Latency
                     | 1V TR_VCYwqCPtbGmNxh | video | 11542 | 1.7mbps | 38.059655ms | 0 (0%)  | 0         | 0bps         |  -

Sub 1 in load-test_1 | Track                | Kind  | Pkts | Bitrate   | Latency     | Dropped | Data Pkts | Data Bitrate | Latency
                     | 1V TR_VCYwqCPtbGmNxh | video | 5434 | 741.8kbps | 38.849474ms | 0 (0%)  | 0         | 0bps         |  -

Sub 2 in load-test_1 | Track                | Kind  | Pkts | Bitrate   | Latency     | Dropped | Data Pkts | Data Bitrate | Latency
                     | 1V TR_VCYwqCPtbGmNxh | video | 3696 | 478.2kbps | 37.955024ms | 0 (0%)  | 0         | 0bps         |  -

Summary for room load-test_0

Summary | Tester               | Bitrate               | Latency     | Total Dropped | Data Bitrate    | Latency | Error
        | Sub 0 in load-test_0 | 4.1mbps               | 37.33238ms  | 0 (0%)        | 0bps            |  -      | -
        | Sub 1 in load-test_0 | 795.0kbps             | 36.941458ms | 0 (0%)        | 0bps            |  -      | -
        | Sub 2 in load-test_0 | 531.8kbps             | 37.663974ms | 0 (0%)        | 0bps            |  -      | -
        | Total                | 5.5mbps (2.7mbps avg) | 37.31622ms  | 0 (0%)        | 0bps (0bps avg) |  -      | 0

Summary for room load-test_1

Summary | Tester               | Bitrate               | Latency     | Total Dropped | Data Bitrate    | Latency | Error
        | Sub 0 in load-test_1 | 1.7mbps               | 38.059655ms | 0 (0%)        | 0bps            |  -      | -
        | Sub 1 in load-test_1 | 741.8kbps             | 38.849474ms | 0 (0%)        | 0bps            |  -      | -
        | Sub 2 in load-test_1 | 478.2kbps             | 37.955024ms | 0 (0%)        | 0bps            |  -      | -
        | Total                | 3.0mbps (1.5mbps avg) | 38.282338ms | 0 (0%)        | 0bps (0bps avg) |  -      | 0
```

#### 4. Launch with a single publisher in 1440p resolution, with two data publishers and two subscribers for each publisher with a 1-minute stream interval without simulcasting:
```shell
./livekit-cli load-test --duration 1m --video-codec h264 --video-resolution "1440p" --data-publishers 2 --no-simulcast --video-publishers 1 --subscribers 2

Statistics for room load-test_0

Sub 0 in load-test_0 | Track            | Kind  | Pkts  | Bitrate | Latency     | Dropped | Data Pkts | Data Bitrate | Latency
                     |  PA_c7x4wmJALxWr | video | 46461 | 7.3mbps | 37.256332ms | 0 (0%)  | 7677      | 1.1mbps      | 31.789094ms

Sub 1 in load-test_0 | Track            | Kind  | Pkts  | Bitrate | Latency     | Dropped | Data Pkts | Data Bitrate | Latency
                     |  PA_jmenhWMbuhXd | video | 46461 | 7.3mbps | 37.163675ms | 0 (0%)  | 7676      | 1.1mbps      | 31.704591ms

Summary for room load-test_0

Summary | Tester               | Bitrate                 | Latency     | Total Dropped | Data Bitrate          | Latency     | Error
        | Sub 0 in load-test_0 | 7.3mbps                 | 37.256332ms | 0 (0%)        | 1.1mbps               | 31.789094ms | -
        | Sub 1 in load-test_0 | 7.3mbps                 | 37.163675ms | 0 (0%)        | 1.1mbps               | 31.704591ms | -
        | Total                | 14.7mbps (14.7mbps avg) | 37.210004ms | 0 (0%)        | 2.1mbps (2.1mbps avg) | 31.746845ms | 0
```


