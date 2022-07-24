# Failing Disk Reporter (FDR)

[![FDR](https://img.shields.io/aur/version/failing-disk-reporter?color=1793d1&logo=arch-linux&style=for-the-badge)](https://aur.archlinux.org/packages/failing-disk-reporter/)
[![MPLv2](https://img.shields.io/aur/license/failing-disk-reporter?color=1793d1&style=for-the-badge)](https://mozilla.org/MPL/2.0/)

Get a report on Matrix or Slack when a drive (HDD or SDD) is failing.

## Rationale

For large infrastructures with 24/7 availability, available monitoring solutions are designed to keep them up-and-running by analyzing numerous parameters in real-time. For smaller setup, without high availability requirement and/or low resources, maintaining the integrity of the data is the primary resource to monitor. In that setup, other failures (such as RAM, network etc) are dealt with as they arise. *Failing Disk Reporter* is a simple tool checking periodically that drives are functional and reports when a failing drive is detected. Reporting to a [Matrix](https://www.matrix.org) room and a Slack channel are supported.

Failing drives are detected using [Smartmontools](https://www.smartmontools.org) using the [S.M.A.R.T.](https://en.wikipedia.org/wiki/S.M.A.R.T.) interface. Smartmontools supports drives connected directly on the motherboard using SATA ports from the [southbridge](https://en.wikipedia.org/wiki/Southbridge_(computing)) and drives connected on hardware RAID cards.

To identify failing drives, criteria defined by [Blackblaze](https://www.backblaze.com/blog/what-smart-stats-indicate-hard-drive-failures) are used, as translated in this [post](https://superuser.com/questions/1171760/how-to-determine-how-dead-a-hdd-is-from-smartctl-report). Users can define different criteria.

## Download

See [refs](https://git.sr.ht/~vejnar/failing-disk-reporter/refs) page for tarball and executable.

Executables are statically linked binaries obtained with disabled [cgo](https://golang.org/cmd/cgo):
```bash
CGO_ENABLED=0 go build *go
```

## Installation

### AUR (Archlinux)

Install the [failing-disk-reporter](https://aur.archlinux.org/packages/failing-disk-reporter) package available on the AUR.

### Manual

1. Install `fdr` executable in `/usr/bin` (or `/usr/local/bin`, in that case change path to `fdr` in [failing-disk-reporter.service](/../../raw/master/systemd/failing-disk-reporter.service))
2. Edit FDR configuration file [fdr.toml](/../../raw/master/config/fdr.toml), then copy it to `/etc`
3. Copy systemd [failing-disk-reporter.service](/../../raw/master/systemd/failing-disk-reporter.service) and [failing-disk-reporter.timer](/../../raw/master/systemd/failing-disk-reporter.timer) to `/etc/systemd/system`

## Configuration

1. Configure FDR in `/etc/fdr.toml`:
    * *[smart]*
        * *ignored_protocols*: List of protocols ignored
    * *[[smart.criteria]]*: List of criteria to identify failing drives
        * *protocol*: For example `ATA`, `NVMe`
        * *key*: SMART attribute
        * *id*: SMART attribute, optional
        * *name*: SMART attribute
        * *label*: Label for report
        * *max*: Threshold for failure
    * *[[reporters]]* Configuration of reporters

2. Enable and start the timer:
    ```bash
    systemctl enable failing-disk-reporter.timer
    systemctl start failing-disk-reporter.timer
    ```

### Matrix reporter

1. Get the access token from the *Help & About* tab in the user config (details in this [post](https://webapps.stackexchange.com/questions/131056/how-to-get-an-access-token-for-riot-matrix)). Input this token in the `TOKEN` parameter of the Matrix reporter.
2. Get the *Internal room ID* from the *Advanced* tab in the config page of the room messages should be sent. Input this room ID in the `ROOM` parameter of the Matrix reporter.

### Slack reporter

1. Create a [Webhook](https://api.slack.com/messaging/webhooks).
2. Input the `TOKENxxx/Bxxx/Gxxx` in `url` parameter of the Slack reporter.

## Test

FDR can be tested with (`-debug` for increasing verbosity and `-report` for sending reports ignoring intervals configured in `fdr.toml`):
```bash
fdr -config config/fdr.toml -debug -report
```

## License

*Failing Disk Reporter* is distributed under the Mozilla Public License Version 2.0 (see /LICENSE).

Copyright (C) 2020-2022 Charles E. Vejnar
