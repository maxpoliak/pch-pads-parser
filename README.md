README Pads Configuration Parser for Intel PCH
==============================================

This is a small utility that parses the inteltool log file and creates
a GPIO configuration for the [coreboot] project.

```bash
(shell)$ git clone https://github.com/maxpoliak/pch-pads-parser.git -b stable_1.0
```
```bash
(shell)$ go run pch-pads-parser.go dw_operations.go sunrise_macro.go -file /path/to/inteltool.log
```
or
```bash
(shell)$ go build
(shell)$ ./pch-pads-parser -file /path/to/inteltool.log
```

Test:
```bash
(shell)$ ./pch-pads-parser -file examples/inteltool-asrock-h110m-dvs.log
(shell)$ ./pch-pads-parser -file examples/inteltool-asrock-h110m-stx.log
```
[coreboot]: https://github.com/coreboot/coreboot
