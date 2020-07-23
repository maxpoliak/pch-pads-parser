Pads Configuration Parser for Intel PCH (intelp2m)
==================================================

This utility allows to convert the configuration DW0/1 registers value
from [inteltool] dump to [coreboot] macros.

```bash
(shell)$ git clone https://github.com/maxpoliak/pch-pads-parser.git -b stable_2.3
(shell)$ make
(shell)$ ./intelp2m -h
(shell)$ ./intelp2m -file /path/to/inteltool.log
```

To generate the gpio.c with raw DW0/1 register values you need to use
the -raw option:

```bash
  (shell)$ ./intelp2m -raw -file /path/to/inteltool.log
```

Test:
```bash
(shell)$ ./intelp2m -file examples/inteltool-asrock-h110m-dvs.log
(shell)$ ./intelp2m -file examples/inteltool-asrock-h110m-stx.log
```

It is possible to use templates for parsing files of excellent inteltool.log.
To specify such a pattern, use the option -t <template number>. For example,
using template type # 1, you can parse gpio.h from an already added board in
the coreboot project.

```bash
(shell)$ ./intelp2m -h
	-t
	template type number
		0 - inteltool.log (default)
		1 - gpio.h
		2 - your template
(shell)$ ./intelp2m -t 1 -file coreboot/src/mainboard/youboard/gpio.h
```
You can also add add a template to 'parser/template.go' for your file type with
the configuration of the pads.

platform type is set using the -p option (Sunrise by default):

```bash
	-p string
	set up a platform
		snr - Sunrise PCH with Skylake/Kaby Lake CPU
		lbg - Lewisburg PCH with Xeon SP CPU
		apl - Apollo Lake SoC
	(default "snr")

(shell)$./intelp2m -p <platform> -file path/to/inteltool.log
```
### Bit fields in macros

Use the -fld option to only generate a sequence of bit fields in a new macro:

```bash
(shell)$./intelp2m -fld -p apl -file ../apollo-inteltool.log
```

```c
_PAD_CFG_STRUCT(GPIO_37, PAD_FUNC(NF1) | PAD_TRIG(OFF) | PAD_TRIG(OFF), PAD_PULL(DN_20K)), /* LPSS_UART0_TXD */
```

### FSP-style macro

The utility allows to generate macros that include fsp/edk2-palforms/slimbootloader-style bitfields:

```c
{ GPIO_SKL_H_GPP_A12, { GpioPadModeGpio, GpioHostOwnAcpi, GpioDirInInvOut, GpioOutLow, GpioIntSci | GpioIntLvlEdgDis, GpioResetNormal, GpioTermNone,  GpioPadConfigLock },	/* GPIO */
```

To do this, use the -fsp option on the command line:

```bash
(shell)$./intelp2m -fsp -p apl -file ../apollo-inteltool.log
```

### Macro Check

After generating the macro, the utility checks all used
fields of the configuration registers. If some field has been
ignored, the utility generates field macros. To not check
macros, use the -n option:

```bash
(shell)$./intelp2m -n -file /path/to/inteltool.log
```

In this case, some fields of the configuration registers
DW0 will be ignored.

```c
PAD_CFG_NF_IOSSTATE_IOSTERM(GPIO_38, UP_20K, DEEP, NF1, HIZCRx1, DISPUPD),		/* LPSS_UART0_RXD */
PAD_CFG_NF_IOSSTATE_IOSTERM(GPIO_39, UP_20K, DEEP, NF1, TxLASTRxE, DISPUPD),	/* LPSS_UART0_TXD */
```

### Information level

The utility can generate additional information about the bit
fields of the DW0 and DW1 configuration registers:

```c
/* GPIO_39 - LPSS_UART0_TXD (DW0: 0x44000400, DW1: 0x00003100) */ --> (2)
/* PAD_CFG_NF_IOSSTATE_IOSTERM(GPIO_39, UP_20K, DEEP, NF1, TxLASTRxE, DISPUPD), */ --> (3)
/* DW0 : PAD_TRIG(OFF) - IGNORED */ --> (4)
_PAD_CFG_STRUCT(GPIO_39, PAD_FUNC(NF1) | PAD_RESET(DEEP) | PAD_TRIG(OFF), PAD_PULL(UP_20K) | PAD_IOSTERM(DISPUPD)),
```

Using the options -i, -ii, -iii, -iiii you can set the info level
from (1) to (4):

```bash
(shell)$./intelp2m -i -file /path/to/inteltool.log
(shell)$./intelp2m -ii -file /path/to/inteltool.log
(shell)$./intelp2m -iii -file /path/to/inteltool.log
(shell)$./intelp2m -iiii -file /path/to/inteltool.log
```
(1) : print /* GPIO_39 - LPSS_UART0_TXD */

(2) : print initial raw values of configuration registers from
inteltool dump
DW0: 0x44000400, DW1: 0x00003100

(3) : print the target macro that will generate if you use the
-n option
PAD_CFG_NF_IOSSTATE_IOSTERM(GPIO_39, UP_20K, DEEP, NF1, TxLASTRxE, DISPUPD),

(4) : print decoded fields from (3) as macros
DW0 : PAD_TRIG(OFF) - IGNORED

### Ignoring Fields

Utilities can generate the _PAD_CFG_STRUCT macro and exclude fields
from it that are not in the corresponding PAD_CFG_*() macro:

```bash
(shell)$./intelp2m -iiii -fld -ign -file /path/to/inteltool.log
```

```c
/* GPIO_39 - LPSS_UART0_TXD DW0: 0x44000400, DW1: 0x00003100 */
/* PAD_CFG_NF_IOSSTATE_IOSTERM(GPIO_39, UP_20K, DEEP, NF1, TxLASTRxE, DISPUPD), */
/* DW0 : PAD_TRIG(OFF) - IGNORED */
_PAD_CFG_STRUCT(GPIO_39, PAD_FUNC(NF1) | PAD_RESET(DEEP), PAD_PULL(UP_20K) | PAD_IOSTERM(DISPUPD)),
```

If you generate macros without checking, you can see bit fields that
were ignored:

```bash
(shell)$./intelp2m -iiii -n -file /path/to/inteltool.log
```

```c
/* GPIO_39 - LPSS_UART0_TXD DW0: 0x44000400, DW1: 0x00003100 */
PAD_CFG_NF_IOSSTATE_IOSTERM(GPIO_39, UP_20K, DEEP, NF1, TxLASTRxE, DISPUPD),
/* DW0 : PAD_TRIG(OFF) - IGNORED */
```

```bash
(shell)$./intelp2m -n -file /path/to/inteltool.log
```

```c
/* GPIO_39 - LPSS_UART0_TXD */
PAD_CFG_NF_IOSSTATE_IOSTERM(GPIO_39, UP_20K, DEEP, NF1, TxLASTRxE, DISPUPD),
```


### Supports Chipsets

  Sunrise PCH, Lewisburg PCH, Apollo Lake SoC

[coreboot]: https://github.com/coreboot/coreboot
[inteltool]: https://github.com/coreboot/coreboot/tree/master/util/inteltool
