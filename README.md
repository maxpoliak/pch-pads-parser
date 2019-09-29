Pads Configuration Parser for Intel PCH
=======================================

A small utility for converting a pad configuration from the inteltool
dump to the PAD_CFG macro for [coreboot] project.

```bash
(shell)$ git clone https://github.com/maxpoliak/pch-pads-parser.git -b stable_2.0
(shell)$ go build
(shell)$ ./pch-pads-parser -file /path/to/inteltool.log
```

To generate the gpio.c with raw DW0/1 register values you need to use
the -raw option:

```bash
  (shell)$ ./pch-pads-parser -raw -file /path/to/inteltool.log
```

Test:
```bash
(shell)$ ./pch-pads-parser -file examples/inteltool-asrock-h110m-dvs.log
(shell)$ ./pch-pads-parser -file examples/inteltool-asrock-h110m-stx.log
```

In the [coreboot], this utility is called `intelp2m` (Intel Pad to Macro):

```bash
  (shell)$ git clone https://review.coreboot.org/coreboot
  (shell)$ cd coreboot/util/inteltool; make
  (shell)$ sudo ./inteltool -G > /path/to/inteltool.log
  (shell)$ cd ../intelp2m
  (shell)$ go build
  (shell)$ ./intelp2m -h
  (shell)$ ./intelp2m -file /path/to/inteltool.log
  (shell)$ cp ./generate/gpio.* ../../src/mainboard/you_mainboard
```


### Supports Chipsets

  Sunrise PCH

TODO:
  Lewisburg PCH
  Apollo Lake SoC PCH

[coreboot]: https://github.com/coreboot/coreboot
