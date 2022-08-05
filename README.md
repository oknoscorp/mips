# mips
Miners Information Polling Service

Main goal of the service is to poll data from CGMINER machines. It can handle thousands of machines, as much
as your network infrastructure allow.

Note: this is a pilot project, mostly tested on farms with up to 2k workers with well configured local network.
Caution: we do not take any responsibility for damage this software may cause to your internal network traffic - we recommend testing on small amount of machines so you can fully avoid network flooding.

Process is very simple, and can be explained in couple of paragraphs:

1. you setup configuration file (define URL from where to pull IP addresses of the workers)
2. define postback URL's (endpoint to a third party service where data will be saved/presented)
3. queue system that saves the data on local disk until third party service responded with success
4. it automatically updates list of workers, in case you shring or expand your IP list it will adapt itself
and request data only from defined machines

