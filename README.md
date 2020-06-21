Pads Configuration Parser for Intel PCH
=======================================

A small utility for converting a pad configuration from the inteltool
dump to the PAD_CFG macro for [coreboot] project.

```bash
(shell)$ git clone https://github.com/maxpoliak/pch-pads-parser.git -b stable_2.1
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

It is possible to use templates for parsing files of excellent inteltool.log.
To specify such a pattern, use the option -t <template number>. For example,
using template type # 1, you can parse gpio.h from an already added board in
the coreboot project.

```bash
(shell)$ ./pch-pads-parser -h
	-t
	template type number
		0 - inteltool.log (default)
		1 - gpio.h
		2 - your template
(shell)$ ./pch-pads-parser -t 1 -file coreboot/src/mainboard/youboard/gpio.h
```
You can also add add a template to 'parser/template.go' for your file type with
the configuration of the pads.

platform type is set using the -p option (Sunrise by default):

```bash
	-p string
	set up a platform
		snr - Sunrise PCH or Skylake/Kaby Lake SoC
		lbg - Lewisburg PCH with Xeon SP
		apl - Apollo Lake SoC
	(default "snr")

./pch-pads-parser -p apl -file path/to/inteltool.log
```

Use the -adv option to only generate extended macros:

```bash
./pch-pads-parser -adv -p apl -file ../apollo-inteltool.log
```

### Supports Chipsets

  Sunrise PCH, Lewisburg PCH, Apollo Lake SoC

[coreboot]: https://github.com/coreboot/coreboot
