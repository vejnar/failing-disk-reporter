# Failing Disk Reporter (FDR)

Get a report when a drive (HDD or SDD) is failing.

## Rationale

For large infrastructures with 24/7 availability, available monitoring solutions are designed to keep them up-and-running by analyzing numerous parameters in real-time. For smaller setup, without high availability requirement and/or low resources, maintaining the integrity of the data is the primary resource to monitor. In that setup, other failures (such as RAM, network etc) are dealt with as they arise. *Failing Disk Reporter* is a simple tool checking periodically that drives are functional and reports when a failing drive is detected. Reporting to a [Matrix](https://www.matrix.org) room and a Slack channel are supported.

Failing drives are detected using [Smartmontools](https://www.smartmontools.org) using the [S.M.A.R.T.](https://en.wikipedia.org/wiki/S.M.A.R.T.) interface. Smartmontools supported drives connected using a SATA HBA or a hardware RAID card.

To identify failing drives, criteria defined by [Blackblaze](https://www.backblaze.com/blog/what-smart-stats-indicate-hard-drive-failures) are used, as translated in this [post](https://superuser.com/questions/1171760/how-to-determine-how-dead-a-hdd-is-from-smartctl-report). Users can define different criteria.

## Download

See [tags](/../../tags) page.

## Installation

1. Install `fdr` executable in `/usr/bin`
2. Edit FDR configuration file [fdr.toml](/../../raw/master/config/fdr.toml), then copy it to `/etc`
3. Copy systemd [systemd/failing-disk-reporter.service](./-/raw/master/failing-disk-reporter.service) and [systemd/failing-disk-reporter.timer](./-/raw/master/failing-disk-reporter.timer) to `/etc/systemd`
4. Enable and start the timer:
    ```bash
    systemctl enable failing-disk-reporter.timer
    systemctl start failing-disk-reporter.timer
    ```

## Configuration

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

### Matrix reporter

1. Get the access token from the *Help & About* tab in the user config (details in this [post](https://webapps.stackexchange.com/questions/131056/how-to-get-an-access-token-for-riot-matrix)). Input this tocken in the `TOKEN` parameter of the Matrix reporter.
2. Get the *Internal room ID* from the *Advanced* tab in the config page of the room messages should be sent. Input this room ID in the `ROOM` parameter of the Matrix reporter.

### Slack reporter

1. Create a [Webhook](https://api.slack.com/messaging/webhooks).
2. Input the `TOKEN` in `url` parameter of the Slack reporter.

## License

*Failing Disk Reporter* is distributed under the Mozilla Public License Version 2.0 (see /LICENSE).

Copyright (C) 2020 Charles E. Vejnar
