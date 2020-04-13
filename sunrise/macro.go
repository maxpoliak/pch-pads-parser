package sunrise

import (
	"fmt"
	"strconv"
	"strings"
)

import "../config"

// macro - contains macro information and methods
// padID : pad ID string
// str   : macro string entirely
// dwcfg : configuration data structure
type macro struct {
	padID  string
	str    string
	dwcfg
	driver bool
}

// dw - returns dwcfg data configuration structure with PAD_CFG_DW0/1 register
// values
func (macro *macro) dw() *dwcfg {
	return &macro.dwcfg
}

// add a string to macro
func (macro *macro) add(str string) *macro {
	macro.str += str
	return macro
}

// set a string in a macro instead of its previous contents
func (macro *macro) set(str string) *macro {
	macro.str = str
	return macro
}

// get macro string
func (macro *macro) get() string {
	return macro.str
}

// Adds PAD Id to the macro as a new argument
// return: macro
func (macro *macro) id() *macro {
	return macro.add(macro.padID)
}

// Add separator to macro if needed
func (macro *macro) separator() *macro {
	str := macro.get()
	c := str[len(str)-1]
	if c != '(' && c != '_' {
		macro.add(", ")
	}
	return macro
}

// Adds the PADRSTCFG parameter from DW0 to the macro as a new argument
// return: macro
func (macro *macro) rstsrc() *macro {
	// The pads from the GPD group in the Sunrise point PCH have a specific Pad Reset
	// Config Map, which differs from that in other groups. PADRSTCFG field in DW0
	// register should implements 3h value as RSMRST and 0h as PWROK
	if config.IsPlatformSunrise() && strings.Contains(macro.padID, "GPD") {
		var gpdResetSrcMap = map[uint8]string{
			0x0: "PWROK",
			0x1: "DEEP",
			0x2: "PLTRST",
			0x3: "RSMRST",
		}
		return macro.separator().add(gpdResetSrcMap[macro.dw().getResetConfig()])
	}

	var resetSrcMap = map[uint8]string{
		0x0: "RSMRST",
		0x1: "DEEP",
		0x2: "PLTRST",
	}
	str, valid := resetSrcMap[macro.dw().getResetConfig()]
	if !valid {
			str = "RESERVED"
	}
	return macro.separator().add(str)
}

// Adds The Pad Termination (TERM) parameter from DW1 to the macro as a new argument
// return: macro
func (macro *macro) pull() *macro {
	var pull = map[uint8]string{
		0x0: "NONE",
		0x2: "5K_PD",
		0x4: "20K_PD",
		0x9: "1K_PU",
		0xa: "5K_PU",
		0xb: "2K_PU",
		0xc: "20K_PU",
		0xd: "667_PU",
		0xf: "NATIVE",
	}
	str, valid := pull[macro.dw().getTermination()]
	if !valid {
		str = "INVALID"
		fmt.Println("Error",
			macro.padID,
			" invalid TERM value = ",
			int(macro.dw().getTermination()))
	}
	return macro.separator().add(str)
}

// Adds Pad GPO value to macro string as a new argument
// return: macro
func (macro *macro) val() *macro {
	return macro.separator().add(strconv.Itoa(macro.dw().getGPIOTXState()))
}

// Adds Pad GPO value to macro string as a new argument
// return: macro
func (macro *macro) trig() *macro {
	var dw = macro.dw()
	var trig = map[uint8]string{
		0x0: "LEVEL",
		0x1: "EDGE_SINGLE",
		0x2: "OFF",
		0x3: "EDGE_BOTH",
	}
	return macro.separator().add(trig[dw.getRXLevelEdgeConfiguration()])
}

// Adds Pad Polarity Inversion Stage (RXINV) to macro string as a new argument
// return: macro
func (macro *macro) invert() *macro {
	macro.separator()
	if macro.dw().getRXLevelConfiguration() {
		return macro.add("INVERT")
	}
	return macro.add("NONE")
}

// Adds input/output buffer state
// return: macro
func (macro *macro) bufdis() *macro {
	var buffDisStat = map[uint8]string{
		0x0: "NO_DISABLE",    // both buffers are enabled
		0x1: "TX_DISABLE",    // output buffer is disabled
		0x2: "RX_DISABLE",    // input buffer is disabled
		0x3: "TX_RX_DISABLE", // both buffers are disabled
	}
	state := macro.dw().getGPIORxTxDisableStatus()
	return macro.separator().add(buffDisStat[state])
}

// Adds macro to set the host software ownership
// return: macro
func (macro *macro) own() *macro {
	if macro.driver {
		return macro.separator().add("DRIVER")
	}
	return macro.separator().add("ACPI")
}

//Adds pad native function (PMODE) as a new argument
//return: macro
func (macro *macro) padfn() *macro {
	nfnum := int(macro.dw().getPadMode())
	if nfnum != 0 {
		return macro.separator().add("NF" + strconv.Itoa(nfnum))
	}
	// GPIO used only for PAD_FUNC(x) macro
	return macro.add("GPIO")
}

// Adds suffix to GPI macro with arguments
func (macro *macro) addSuffixInput() {
	dw := macro.dw()
	isEdge := dw.getRXLevelEdgeConfiguration()
	macro.add("_GPI")
	switch {
	case dw.getGPIOInputRouteNMI():
		// e.g. PAD_CFG_GPI_NMI(GPIO_24, UP_20K, DEEP, LEVEL, INVERT),
		macro.add("_NMI").add("(").id().pull().rstsrc().trig().invert()

	case dw.getGPIOInputRouteIOxAPIC():
		// e.g. PAD_CFG_GPI_APIC(GPP_B3, NONE, PLTRST)
		macro.add("_APIC")
		if isEdge == 0 {
			if dw.getRXLevelConfiguration() {
				// e.g. PAD_CFG_GPI_APIC_INVERT(GPP_C5, DN_20K, DEEP),
				macro.add("_INVERT")
			}
			macro.add("(").id().pull().rstsrc()
		} else {
			// PAD_CFG_GPI_APIC_IOS(GPP_C20, NONE, DEEP, LEVEL, INVERT,
			// TxDRxE, DISPUPD) macro isn't used for this chipset. But, in
			// the case when SOC_INTEL_COMMON_BLOCK_GPIO_LEGACY_MACROS is
			// defined in the config, this macro allows you to set trig and
			// invert parameters.
			macro.add("_IOS").add("(").id().pull().rstsrc().trig().invert()
			// IOStandby Config will be ignored for the Sunrise PCH, so use
			// any values
			macro.add(", TxDRxE, DISPUPD")
		}

	case dw.getGPIOInputRouteSCI():

		// e.g. PAD_CFG_GPI_SCI(GPP_B18, UP_20K, PLTRST, LEVEL, INVERT),
		if (isEdge & 0x1) != 0 {
			// e.g. PAD_CFG_GPI_ACPI_SCI(GPP_G2, NONE, DEEP, YES),
			macro.add("_ACPI")
		}
		macro.add("_SCI").add("(").id().pull().rstsrc()
		if (isEdge & 0x1) == 0 {
			macro.trig()
		}
		macro.invert()

	case dw.getGPIOInputRouteSMI():
		if (isEdge & 0x1) != 0 {
			// e.g. PAD_CFG_GPI_ACPI_SMI(GPP_I3, NONE, DEEP, YES),
			macro.add("_ACPI")
		}
		macro.add("_SMI").add("(").id().pull().rstsrc()
		if (isEdge & 0x1) == 0 {
			// e.g. PAD_CFG_GPI_SMI(GPP_E7, NONE, DEEP, LEVEL, NONE),
			macro.trig()
		}
		macro.invert()

	case isEdge != 0 || macro.driver:
		// e.g. PAD_CFG_GPI_TRIG_OWN(pad, pull, rst, trig, own)
		macro.add("_TRIG_OWN").add("(").id().pull().rstsrc().trig().own()

	default:
		if macro.driver {
			// e. g. PAD_CFG_GPI_GPIO_DRIVER(GPIO_212, UP_20K, DEEP),
			macro.add("_GPIO_DRIVER")
		}
		// e.g. PAD_CFG_GPI(GPP_A7, NONE, DEEP),
		macro.add("(").id().pull().rstsrc()
	}
}

// Adds suffix to GPO macro with arguments
func (macro *macro) addSuffixOutput() {
	term := macro.dw().getTermination()
	// FIXME: don`t understand how to get PAD_CFG_GPI_GPIO_DRIVER(..)
	if term != 0 {
		// e.g. PAD_CFG_TERM_GPO(GPP_B23, 1, DN_20K, DEEP),
		macro.add("_TERM")
	}
	macro.add("_GPO")
	if macro.driver {
		// PAD_CFG_GPO_GPIO_DRIVER(pad, val, rst, pull)
		macro.add("_GPIO_DRIVER")
	}
	macro.add("(").id().val()
	if term != 0 {
		macro.pull()
	}
	macro.rstsrc()
	if macro.driver {
		macro.pull()
	}

	// Fix mask for RX Level/Edge Configuration (RXEVCFG)
	// See https://github.com/coreboot/coreboot/commit/3820e3c
	macro.dw().maskTrigFix()
}

const (
	rxDisable uint8 = 0x2
	txDisable uint8 = 0x1
)

// Set common macros macros if the standard ones can't define the values from
// the DW0/1 register
func (macro *macro) common() *macro {
	macro.set("_PAD_CFG_STRUCT(").id().
		add(",\n\t\tPAD_FUNC(").padfn().
		add(") | PAD_RESET(").rstsrc().add(") |\n\t\t")

	var str string
	var argument = true
	switch dw := macro.dw(); {
	case dw.getGPIOInputRouteIOxAPIC():
		str = "IOAPIC"
	case dw.getGPIOInputRouteSCI():
		str = "SCI"
	case dw.getGPIOInputRouteSMI():
		str = "SMI"
	case dw.getGPIOInputRouteNMI():
		str = "NMI"
	default:
		argument = false
	}

	if argument {
		macro.add("PAD_IRQ_CFG(").add(str).trig().invert().add(")")
	} else {
		macro.add("PAD_CFG0_TRIG_").trig().add(" | PAD_CFG0_RX_POL_").invert()
	}

	macro.add(" |\n\t\t").add("PAD_BUF(").bufdis().add(")")
	if macro.dw().getGPIOTXState() != 0 {
		macro.add(" | 1")
	}
	macro.add(",\n\t\tPAD_CFG_OWN_GPIO(").own().add(") | ")
	return macro.add("PAD_PULL(").pull().add(")),")
}

// Check created macro
func (macro *macro) check() {
	if !macro.dw().maskCheck(0) {
		macro.common()
		// Debug message about this
		fmt.Printf("\nNo configuration for pad: %s\n", macro.padID)
		fmt.Printf("Use %s\n", macro.get())
	}
}

// Gets base string of current macro
// return: string of macro
//         error
func (macro *macro) generate() string {
	dw := macro.dw()
	macro.set("PAD_CFG")
	if dw.getPadMode() == 0 {
		// GPIO
		switch dw.getGPIORxTxDisableStatus() {
		case txDisable:
			// GPI
			macro.addSuffixInput()

		case rxDisable:
			// GPO
			macro.addSuffixOutput()

		case rxDisable | txDisable:
			// NC
			macro.set("PAD_NC").add("(").id().pull()
			// Fix mask for RX Level/Edge Configuration (RXEVCFG)
			// and Pad Reset Config (PADRSTCFG)
			// There is no need to check these fields if the pad
			// is in the NC state
			macro.dw().maskResetFix().maskTrigFix()

		default:
			// In case the rule isn't found, a common macro is used
			// to create the pad configuration
			return macro.common().get()
		}
	} else {
		isEdge := dw.getRXLevelEdgeConfiguration() != 0
		isTxRxBufDis := dw.getGPIORxTxDisableStatus() != 0
		// e.g. PAD_CFG_NF(GPP_D23, NONE, DEEP, NF1)
		macro.add("_NF")
		if isEdge || isTxRxBufDis {
			// e.g. PCHHOT#
			// PAD_CFG_NF_BUF_TRIG(GPP_B23, 20K_PD, PLTRST, NF2, RX_DIS, OFF),
			macro.add("_BUF_TRIG")
		}
		macro.add("(").id().pull().rstsrc().padfn()
		if isEdge || isTxRxBufDis {
			macro.bufdis().trig()
		}
	}
	macro.add("),")
	macro.check()
	return macro.get()
}

// GenMacro - generate pad macro
// dw0 : DW0 config register value
// dw1 : DW1 config register value
// return: string of macro
//         error
func GenMacro(id string, dw0 uint32, dw1 uint32, driver bool) string {
	macro := macro{padID: id, dwcfg: dwcfg{reg: [MaxDWNum]uint32{dw0, dw1}}, driver : driver}
	return macro.generate()
}
