# Shark 100 Power Meter Monitoring and Display

The Shark 100 power meter is network-connected but does not have a web server.
It does, however, provide a TCP modbus interface.

This is a simple go server that reads power draw, voltage, and frequency
from the Shark and either logs it to the console or makes it available
via a webserver running on (default) port 8081.

Note that this version displays the voltage across both hot lines,
as I'm using it in a 240V application.  You can change the registers
it reads to show different values.  (They're listed in the manual
for the meter).

The IP address of the shark meter is hard-coded.  If you find
this program useful and would like me to generalize that a little,
let me know.  Otherwise, I assume I'm either the only one using it,
or that anyone else is using it as reference code for hacking their
own thing up.

The included 'powermon.html' web page has a bit of javascript
that auto-refreshes from the Go web server.  It's formatted
to fit nicely on a horizontally-oriented iPad.
