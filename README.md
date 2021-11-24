# YTSync Tool
[![Build Status](https://travis-ci.com/lbryio/ytsync.svg?branch=master)](https://travis-ci.com/lbryio/ytsync)

This tool serves lbry to parse youtube channels that want their content mirrored on LBRY.

The tool downloads the entire set of public videos from a given channel, publishes them to LBRY and populates our private database in order to keep track of what's publishes.
With the support of said database, the tool is also able to keep all the channels updated.


# Requirements
- lbrynet SDK https://github.com/lbryio/lbry-sdk/releases (We strive to keep the latest release of ytsync compatible with the latest major release of the SDK)
- a lbrycrd node running (localhost or on a remote machine) with credits in it
- internal-apis (you cannot run this one yourself)
- python3-pip
- yt-dlp (`pip3 install -U yt-dlp`)
- ffmpeg (latest)

# Setup
- make sure daemon is stopped and can be controlled through `systemctl` (find example below)
- extract the ytsync binary anywhere
- create and fill `config.json` using [this example](config.json.example)

## systemd script example
`/etc/systemd/system/lbrynet.service`
```
[Unit]
Description="LBRYnet daemon"
After=network.target

[Service]
Environment="HOME=/home/lbry"
ExecStart=/opt/lbry/lbrynet start
User=lbry
Group=lbry
Restart=on-failure
KillMode=process

[Install]
WantedBy=multi-user.target
```

# Instructions

```
Publish youtube channels into LBRY network automatically.

Usage:
  ytsync [flags]

Flags:
      --after int                   Specify from when to pull jobs [Unix time](Default: 0)
      --before int                  Specify until when to pull jobs [Unix time](Default: current Unix time) (default 1669311891)
      --channelID string            If specified, only this channel will be synced.
      --concurrent-jobs int         how many jobs to process concurrently (default 1)
  -h, --help                        help for ytsync
      --limit int                   limit the amount of channels to sync
      --max-length int              Maximum video length to process (in hours) (default 2)
      --max-size int                Maximum video size to process (in MB) (default 2048)
      --max-tries int               Number of times to try a publish that fails (default 3)
      --no-transfers                Skips the transferring process of videos, channels and supports
      --quick                       Look up only the last 50 videos from youtube
      --remove-db-unpublished       Remove videos from the database that are marked as published but aren't really published
      --run-once                    Whether the process should be stopped after one cycle or not
      --skip-space-check            Do not perform free space check on startup
      --status string               Specify which queue to pull from. Overrides --update
      --status2 string              Specify which secondary queue to pull from.
      --takeover-existing-channel   If channel exists and we don't own it, take over the channel
      --update                      Update previously synced channels instead of syncing new ones
      --upgrade-metadata            Upgrade videos if they're on the old metadata version
      --videos-limit int            how many videos to process per channel (leave 0 for automatic detection)

```

## Running from Source

Clone the repository and run `make` 

## License

This project is MIT licensed. For the full license, see [LICENSE](LICENSE).

## Contributing

Contributions to this project are welcome, encouraged, and compensated. For more details, see [CONTRIBUTING](https://lbry.tech/contribute).

## Security

We take security seriously. Please contact [security@lbry.io](mailto:security@lbry.io) regarding any security issues. Our PGP key is [here](https://lbry.com/faq/pgp-key) if you need it.

## Contact

The primary contact for this project is [Niko Storni](https://github.com/nikooo777) (niko@lbry.com).

## Additional Info and Links

- [https://lbry.com](https://lbry.com) - The live LBRY website
- [Discord Chat](https://chat.lbry.com) - A chat room for the LBRYians
- [Email us](mailto:hello@lbry.com) - LBRY Support email
- [Twitter](https://twitter.com/@lbryio) - LBRY Twitter page
- [Facebook](https://www.facebook.com/lbryio/) - LBRY Facebook page
- [Reddit](https://reddit.com/r/lbry) - LBRY Reddit page
- [Telegram](https://t.me/lbryofficial) - Telegram group
