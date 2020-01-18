# Reverse Engineering remote control Electric Radiators

The radiators in my apartment are electric and remotely controlled by battery powered devices mounted nearby on the wall.

This setup is not supported by any of the popular home thermostats so I decided to build my own control system.

## Understanding the RF controllers

Each controller is mounted on the wall and has a display showing the current temperature and has buttons for configuring the target temperature.
One controller can be paired to multiple radiators which will receive the same settings whenever the controller sends an update.

There is no documentation, as far as I know, as to the protocol, frequency, modulation, etc.

### Discovering the frequency

Following a [guide by someone reverse engineering their home alarm](https://funoverip.net/2014/11/reverse-engineer-a-verisure-wireless-alarm-part-1-radio-communications/)
I managed to pinpoint the frequency of the communications using [SDR#](http://airspy.com/quickstart/) and an [RTL SDR](http://www.rtl-sdr.com/about-rtl-sdr/).

### Capturing a packet

At this point I disassembled a broken remote control in the hope that it would enlighten me some more as to the protocol and modulation.
Inside I found a [TI CC1100](http://www.ti.com/product/CC1100), similar to in the other reverse engineering article!

From here I played around with [GNURadio](https://www.gnuradio.org/) until I managed to capture and demodulate (2-FSK) the signal and save the waveform.

### Decoding the packet

I struggled to get GNURadio to fully decode the packet so I [opened the waveform in Audacity](https://raw.githubusercontent.com/hatstand/shinywaffle/master/data/winter-bedroom-waveform.png) and tried to decode it by hand.
It was tedious work but it was fairly easy to [separate the waveform into the different parts of the packet](https://raw.githubusercontent.com/hatstand/shinywaffle/master/data/sdr-notes.png) [as documented by TI](http://www.ti.com/lit/ds/symlink/cc1101.pdf):

  * Preamble: 16 bits alternating high/low
  * Sync Word: An arbitrarily chosen 16 bit word
  * Length: 1 single byte packet length indicator (not including CRC)
  * Data: 9 bytes of data
  * CRC: 2 bytes of CRC
  
At this point I was clearly on the right track as the sync word came out as `0xd391`, the same as the other blog and also the default sync word for the CC1101
(clearly, the developers were lazy on both products).

### Building a receiver

So now we know the rough parameters of the signal and the packet format:

|||
|---|---|
|Frequency |868.3MHz|
|Deviation |40KHz   |
|Modulation|2-FSK   |

With this we should be able to receive packets.

Initially I tried to get this to work with GNURadio but gave up and discovered that you can buy a CC1101 (the successor to the CC1100) on a board on ebay for a few bucks.
The first one I ordered turned out to be configured for 433MHz and it turns out the RF filter is different between a 433MHz and an 868MHz board.

On receiving an 868MHz CC1101 I hooked it up to an Arduino and started playing around but it was behaving weirdly and I gave up.

A while later it occurred to me that the voltage levels on a standard Arduino Uno are 5V and the CC1101 expects 3.3V!

From here I connected it to a Raspberry Pi Zero W (which has 3.3V levels).
To add to the fun I decided to try to interface with the CC1101 using go after discovering the [EMBD](http://embd.kidoman.io/) framework.
This makes interfacing with SPI and I2C and so on relatively easy and with a much nicer development environment than the generally awful Arduino code lying around the web.

### A working CC1101

After some hacking around there was great rejoicing at the point where I could read the `PARTNUM` and `VERSION` registers
from the CC1101 over SPI and see that they matched the expected values from the datasheet.

From here I could more or less fully control the CC1101 to do what I wanted.
Unfortunately, the values for the frequency configuration registers are fairly opaque.
The only real way to configure them is to download [TI's SmartRF Studio](http://www.ti.com/tool/smartrftm-studio).
This GUI tool lets you configure the CC1101 registers and then export the values to a file.
The nice part is that it lets you write your own template for exporting the register values so it was trivial to generate a file of go constants for my guessed settings.

### Receiving Packets

Now my CC1101 was configured correctly and when set to receive mode, I would actually see packets arriving.
I could see packets that correlated well with one of the remote controls sending a packet but sadly there was also a lot of other junk.

After digging through the datasheet I discovered it is possible to make the CC1101 only interrupt when it receives a full packet with a valid CRC.
Configuring this option suddenly cleared all the chaff and I could see just the packets that were good candidates for being from a controller.

### Decoding Packets

So now I could see packets arriving.
From here, I tried each remote in turn and each setting on a single remote in turn and [documented the packets generated](https://github.com/hatstand/shinywaffle/blob/master/data/protocol).

From inspection it was easy to see that the packet format was roughly:

|0x0|0x1|0x2|0x3|0x4|0x5|0x6|0x7|0x8|
|---|---|---|---|---|---|---|---|---|
|IDENT2|IDENT1|IDENT0|ADDR1|ADDR0|MODE|DAYTEMP|NIGHTTEMP|DEFROSTTEMP|

The first 3 bytes are some unique identifier so the system knows these packets are for it (they probably should have just set a unique sync word and let the CC1101 do the hard work).
The next 2 bytes are the target address of a radiator.
The next byte is the mode that the radiator should change to (day, night, defrost or off).
The last 3 bytes are the target temperatures in degrees Celsius for each mode multiplied by two (so it can represent half degrees).

### Sending Packets

So now we understand and can follow the packets flying around my apartment.
The next step is being able to send our own packets so we can replace the remote controls.

With only a small amount more code it was simple to start sending identical packets via the CC1101.
The only hiccup was that occasionally a radiator would miss a packet. This was easily solved by just transmitting each packet three times as they are idempotent.

### More Work

An unfortunate part of this protocol is that, while the remote controls can display the _current_ room temperature, they do not transmit it at any point.
So, in order to build an automated thermostat I needed to record the temperature some other way.

After a little research, I stumbled upon the [SHT3x series](https://www.sensirion.com/en/environmental-sensors/humidity-sensors/digital-humidity-sensors-for-various-applications/)
of temperature and humidity sensors.

These sensors work using I2C so it ended up being simple to interface with one from both a Raspberry Pi using EMBD and an Arduino using the builtin I2C library.

The final steps will be combining a CC1101, a SHT3x and an Arduino Pro Mini to try to create a low power (coin battery for ~1 year) wireless temperature sensor
to complement the remote control of the radiators.





