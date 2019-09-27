package parser

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

import "../sunrise"

// padInfo - information about pad
// id       : pad id string
// offset   : the offset of the register address relative to the base
// function : the string that means the pad function
// dw0      : DW0 register value
// dw1      : DW1 register value
type padInfo struct {
	id        string
	offset    uint16
	function  string
	dw0       uint32
	dw1       uint32
}

// add - add information about pad to data structure
// line : string from inteltool log file
func (info *padInfo) add(line string) {
	var val uint64
	if strings.HasPrefix(line, "----") {
		// ------- GPIO Group GPP_A -------
		// ------- GPIO Community 0 -------
		// Add header to define GPIO group or community
		info.function = line
		return
	}
	// 0x0520: 0x0000003c44000600 GPP_B12  SLP_S0#
	// 0x0438: 0xffffffffffffffff GPP_C7   RESERVED
	fmt.Sscanf(line,
		"0x%x: 0x%x %s %s",
		&info.offset,
		&val,
		&info.id,
		&info.function)
	info.dw0 = uint32(val & 0xffffffff)
	info.dw1 = uint32(val >> 32)
}

// titleFprint - print GPIO group title to file
// gpio : gpio.c file descriptor
func (info *padInfo) titleFprint(gpio *os.File) {
	fmt.Fprintf(gpio, "\n\t/* %s */\n", info.function)
}

// reservedFprint - print reserved GPIO to file as comment
// gpio : gpio.c file descriptor
func (info *padInfo) reservedFprint(gpio *os.File) {
	// small comment about reserved port
	fmt.Fprintf(gpio, "\t/* %s - %s */\n", info.id, info.function)
}

// padInfoRawFprint - print information about current pad to file using
// raw format:
// _PAD_CFG_STRUCT(GPP_F1, 0x84000502, 0x00003026), /* SATAXPCIE4 */
// gpio : gpio.c file descriptor
func (info *padInfo) padInfoRawFprint(gpio *os.File) {
	fmt.Fprintf(gpio,
		"\t_PAD_CFG_STRUCT(%s, 0x%0.8x, 0x%0.8x), /* %s */\n",
		info.id,
		info.dw0,
		(info.dw1 & 0xffffff00), // Interrupt Select - RO
		info.function)
}

// padInfoMacroFprint - print information about current pad to file using
// special macros:
// PAD_CFG_NF(GPP_F1, 20K_PU, PLTRST, NF1), /* SATAXPCIE4 */
// gpio : gpio.c file descriptor
func (info *padInfo) padInfoMacroFprint(gpio *os.File) {
	fmt.Fprintf(gpio, "\t/* %s - %s */\n\t%s\n",
		info.id,
		info.function,
		sunrise.GetMacro(info.id, info.dw0, info.dw1))
}

// ParserData - global data
// padmap  : pad info map
// dbgFlag : gebug flag, currently not used
type ParserData struct {
	padmap  []padInfo
	DbgFlag bool
}

// padInfoAdd - adds a new entry to pad info map
// line      : string/line from the inteltool log file
// community : pads community number
func (parser *ParserData) padInfoAdd(line string) {
	pad := padInfo{}
	pad.add(line)
	parser.padmap = append(parser.padmap, pad)
}

// getCommunity - scans the string and returns the pads community number
// line   : string from inteltool log file
// return
// community number
func (parser *ParserData) getCommunity(line string) uint8 {
	var comm uint8
	fmt.Sscanf(line, "------- GPIO Community %d -------", &comm)
	return comm
}

// PadMapFprint - print pad info map to file
// gpio : gpio.c descriptor file
// raw  : in the case when this flag is false, pad information will be print
//        as macro
func (parser *ParserData) PadMapFprint(gpio *os.File, raw bool) {
	gpio.WriteString("\n/* Pad configuration in ramstage */\n")
	gpio.WriteString("static const struct pad_config gpio_table[] = {\n")
	for _, pad := range parser.padmap {
		switch pad.dw0 {
		case 0:
			pad.titleFprint(gpio)
		case 0xffffffff:
			pad.reservedFprint(gpio)
		default:
			if raw {
				pad.padInfoRawFprint(gpio)
			} else {
				pad.padInfoMacroFprint(gpio)
			}
		}
	}
	gpio.WriteString("};\n")

	// FIXME: need to add early configuration
	gpio.WriteString(`/* Early pad configuration in romstage. */
static const struct pad_config early_gpio_table[] = {
	/* TODO: Add early pad configuration */
};

const struct pad_config *get_gpio_table(size_t *num)
{
	*num = ARRAY_SIZE(gpio_table);
	return gpio_table;
}

const struct pad_config *get_early_gpio_table(size_t *num)
{
	*num = ARRAY_SIZE(early_gpio_table);
	return early_gpio_table;
}

`)
}

// Parse pads groupe information in the inteltool log file
// logFile : name of inteltool log file
// return
// err : error
func (parser *ParserData) Parse(logFile string) (err error) {
	file, err := os.Open(logFile)
	if err != nil {
		return err
	}
	defer file.Close()

	// Read all lines from inteltool log file
	fmt.Println("Parse IntelTool Log File...")
	scanner := bufio.NewScanner(file)
	var line string
	for scanner.Scan() {
		line = scanner.Text()
		// Use only the string that contains the GPP information
		if !strings.Contains(line, "GPP_") && !strings.Contains(line, "GPD") {
				continue
		}
		parser.padInfoAdd(line)
	}
	fmt.Println("...done!")
	return nil
}