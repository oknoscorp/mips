# mips
Miners info polling service.

Server implementation is very simple and for now only handles
cgminer/bmminer machines.

Lucid chart scheme: https://lucid.app/lucidchart/5dd1968a-4f32-4b08-b787-3d31972484be/edit?viewport_loc=-398%2C-1753%2C2827%2C1343%2C0_0&invitationId=inv_25567ef1-4a69-48ef-96e4-b601984aa5ec#

Features:
<ol>
    <li>Fetch IP list from remote destinations</li>
    <li>Execute API command to fetch current miner info</li>
    <li>Push collected data to remote location</li>
    <li>Execute worker Firmware upgrade</li>
    <li>Automatic worker reboot based on predefined rules</li>
    <li>Local file copy to remote location</li>
</ol>

Execute `make all` command to generate binary.

This service is standalone application that can compile for any type of processor supported by GO compiler.

1. Fetch IP list from remote desitnation
In configuration.yaml we can specify several endpoints where lists of IP's are stored. Application automatically pushes and removes IP's from the queue.
We attach listener to each IP which will trigger API request to machine every 3 minutes to obtain current status and information.
Informations are save inside .json files on local machine. They are processed and pushed to remote endpoint from Queue list. In case remote server responds with
error message is pushed to queue again, but repetition time is extended. After every successful push message is removed from the queue and file is deleted.
2. 
