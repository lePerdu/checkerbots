## Links

- https://hackingroomba.com/
- https://wiki.ros.org/Robots/Roomba
- https://github.com/koalazak/dorita980
- https://cdn.hackaday.io/files/1747287475562752/Roomba_SCI_manual.pdf
- Hack Your Old Roomba to a Smart Robotic Vacuum for $5: https://www.youtube.com/watch?v=t2NgA8qYcFI
- You Should Hack Your Roomba: https://www.youtube.com/watch?v=mTpkV7xZln0
- I put Javascript on my Roomba Vacuum: https://www.youtube.com/watch?v=4jAM5P7PcK0
- Hacking my Roomba to add ONE MISSING Feature!: https://www.youtube.com/watch?v=2SWj49ugr3A

### Serial spec

- 600: https://github.com/ShonP40/ESPRoomba/blob/master/iRobot%20Roomba%20600%20Open%20Interface%20Spec.pdf
- Original?: https://www.cgl.cs.tau.ac.il/wp-content/uploads/2025/06/create-open-interface_v2.pdf
- Original?: https://cdn.hackaday.io/files/1747287475562752/Roomba_SCI_manual.pdf

## Roomba Open Interface

### Wiring

(Do I need an oscilloscope to verify voltage levels before I accidentally fry something by mis-matching voltages?)

#### -> USB -> Computer

USB-TTL adapter (5V), such as CH340/CH341.

#### -> Arduino

Need to convert from 5V to 3.3V.
- 5V -> 3.3V: Voltage divider.
- 3.3V -> 5V: Try nothing.

Some concrete wiring suggestions:
- https://electronics.stackexchange.com/questions/186168/how-to-convert-uart-voltage-from-5v-to-3-3v
- https://cdn.sparkfun.com/tutorialimages/BD-LogicLevelConverter/an97055.pdf

Also need to configure one of the serial controllers to use specific pins:
- https://github.com/ostaquet/Arduino-Nano-33-IoT-Ultimate-Guide#how-to-use-serial-communication-why-there-is-no-softwareserialh-in-the-arduino-nano-33-iot
- (maybe, not sure if it's for the same board): https://docs.arduino.cc/tutorials/communication/SamdSercom/

Level converters:
- 2N7001T: https://www.digikey.com/en/products/detail/texas-instruments/2N7001TDCKR/9554632

### Power -> Arduino

"Buck converter"

### ESP-01(s)

Small, cheap ESP8266-based board with WiFi to serve as a bridge between the Roomba and the game server.

Arduino Nano 33 boards I have work as well, they just have extra peripherals (IMU, Bluetooth, more GPIO) and are more expensive.
