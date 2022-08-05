# mips
Miners information polling service. (cgminer/bmminer/asic miners)

Note: this a pilot projewith no long terms plans for project maintenance.


This software is useful for polling informations about cgminer/bmminer (asics) workers.
Polled information can be pushed to third party endpoint with all of the data listed in cgminer stats API.

Default port for the tool daemon :3011 and it is fetching information from :4028 workers API port.

Mechanism is simple:
- you define endpoint that contains list of IP addresses assigned to machines (one IP per line)
- you configure how often poller will collect the data (default is every 3 minutes) per machine
- you define third party endpoint where collected data should be pushed.
- in case postback request fails (third party service responded with error code) it will retry sending but with a little bit delay (we assume gone server will come to life ASAP)

Caution: we suggest you to test tool on couple of dozens machines and expand it slowly. There is a chance you can get a network flood
in case you network infrastructure is poorly configred (slow router, slow DHCP leases etc...)

You can post your questions to "Issues" section.

### Installation
1. run `make` command to compile the program
2. setup `systemd` daemon so you can execute the process in the background
3. run `systemctl start mips.service`