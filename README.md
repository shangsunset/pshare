# pshare

> A peer to peer file sharing command line tool.

## Installation

```
go get -u github.com/shangsunset/pshare/cmd/psh
```

## Usage

```
NAME:
   psh - A peer to peer file sharing cli application

USAGE:
   psh [global options] command [command options] [arguments...]

VERSION:
   0.0.0

COMMANDS:
     share, s  share content with peers
     recv, r   receive content from peer
     help, h   Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h     show help
   --version, -v  print the version
```


Broadcast to peers
```
psh share {filename}

Your service hash to share: {hashA}
```

Receive from broadcast
```
psh recv -s {hashA}
```

Private sharing
```
psh share -p {filename}

Your service hash to share: {hashA}
Your private instance hash to share: {hashB}
```

Receive from private sharing
```
psh recv -s {hashA} -p {hashB}
```

[MIT](./LICENSE)
