# ESP8266 Updater

Distribute binaries for ESP8266 OTA.

## Build

Run `make`. The binary is compiled for ARM to run on a Raspberry Pi. If you want to build for eg `amd64` run `GOARCH=amd64 make`.

## Configure

Add a config file to `/etc/esp8266-updater/config.yml`.

The config file contains the binary versions for your esp8266.

```
versions:
  "12:23:45:67:89:0A": "foo-bar-0.0.1"
```

Then place a binary called `foo-bar-0.0.1.bin` into `/var/lib/esp8266-updater/`.

## Run

Run `make install` or copy the binary to `/usr/local/bin/` manually.

Add the systemd unit file `systemd/esp8266-updater.service` and start the unit.

```
sudo cp systemd/esp8266-updater.service /etc/systemd/system
sudo systemctl enable esp8266-updater
sudo systemctl start esp8266-updater
```
