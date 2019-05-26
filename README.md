README Pads Configuration Parser for Intel PCH
==============================================

This is a small utility that parses the inteltool log file and creates
a GPIO configuration for the [coreboot] project.

```bash
(shell)$ git clone https://github.com/maxpoliak/pch-pads-parser.git
```
```bash
(shell)$ go run pch-pads-parser.go dw_operations.go sunrise_macro.go -file /path/to/inteltool.log
```
or
```bash
(shell)$ go build
(shell)$ ./pch-pads-parser -file /path/to/inteltool.log
```
[coreboot]: https://github.com/coreboot/coreboot
