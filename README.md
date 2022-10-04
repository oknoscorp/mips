# mips
Miners information polling service. (cgminer/bmminer/asic miners)

Note: this a pilot project with no long term plans for project maintenance.

This software polls information about cgminer/bmminer (ASIC) workers.
Polled information can be pushed to third party endpoint with all of the data listed in cgminer stats API.

Default port for the tool daemon is :3011

Mechanism is simple:
- you define endpoints that contains list of machines IP addresses (one IP per line)
- you configure how often poller will collect the data (default is every 3 minutes) per machine
- you define third party endpoint, collected data will be pushed to that endpoint in JSON format
- in case postback request fails (third party service responded with error code) it will retry sending, but with a little bit of delay (we assume gone server will come to life ASAP)

Caution: we suggest you to test tool on dozens of machines and expand it slowly. There is a chance you can get a network flood
in case your network infrastructure is poorly configred (slow router, slow DHCP leases etc...)

You can post your questions to "Issues" section.

### Installation
1. rename `configuration_example.yaml` to `configuration.yaml` so tool can fetch appropriate configuration
2. run `make` command to compile the program
3. setup `systemd` daemon so you can execute the process in the background, or install `supervisor` for your version of OS