package apollo

import (
	"fmt"
	"strconv"
	"strings"
)

// macro - contains macro information and methods
// padID : pad ID string
// str   : macro string entirely
// dwcfg : configuration data structure
type macro struct {
	padID string
	str   string
	dwcfg
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
	var resetSrc = map[uint8]string{
		0x0: "PWROK",
		0x1: "DEEP",
		0x2: "PLTRST",
	}
	str, valid := resetSrc[macro.dw().getResetConfig()]
	if !valid {
		// Pad Reset Config field in DW0 register should implements 3h value
		// as RSMRST for GPD pads group, but this value is reserved for other
		// groups
		if strings.Contains(macro.padID, "GPD") {
			str = "RSMRST"
		} else {
			str = resetSrc[0]
			fmt.Println("Warning:", macro.padID, "PADRSTCFG = 3h Reserved")
			fmt.Println("It is implemented as setting 0h (PWROK) for Sunrise PCH")
		}
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
		return macro.add("YES")
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

// no common

// Add a line to the macro that defines IO Standby State
// return: macro
func (macro *macro) iosstate() *macro {
	var stateMacro = map[uint8]string{
		0x0: "TxLASTRxE",
		0x1: "Tx0RxDCRx0",
		0x2: "Tx0RxDCRx1",
		0x3: "Tx1RxDCRx0",
		0x4: "Tx1RxDCRx1",
		0x5: "Tx0RxE",
		0x6: "Tx1RxE",
		0x7: "HIZCRx0",
		0x8: "HIZCRx1",
		0x9: "TxDRxE",
		0xf: "IGNORE",
	}
	str, valid := stateMacro[macro.dw().getIOStandbyState()]
	if !valid {
		// ignore setting for incorrect value
		str = "IGNORE"
	}
	return macro.separator().add(str)
}

// Add a line to the macro that defines IO Standby Termination
// return: macro
func (macro *macro) ioterm() *macro {
	var iotermMacro = map[uint8]string{
		0x0: "SAME",
		0x1: "DISPUPD",
		0x2: "ENPD",
		0x3: "ENPU",
	}
	return macro.separator().add(iotermMacro[macro.dw().getIOStandbyTermination()])
}
// end no common

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
		if dw.getIOStandbyState() != 0 || dw.getIOStandbyTermination() != 0 {
			// e.g. H1_PCH_INT_ODL
			// PAD_CFG_GPI_APIC_IOS(GPIO_63, NONE, DEEP, LEVEL, INVERT, TxDRxE, DISPUPD),
			macro.add("_IOS")
			macro.add("(").id().pull().rstsrc().trig().invert().iosstate().ioterm()
		} else {
			if dw.getRXLevelConfiguration() {
				// e.g. PAD_CFG_GPI_APIC_INVERT(GPP_C5, DN_20K, DEEP),
				macro.add("_INVERT")
			}
			// PAD_CFG_GPI_APIC(GPP_C5, DN_20K, DEEP),
			macro.add("(").id().pull().rstsrc()
		}

	case dw.getGPIOInputRouteSCI():
		if dw.getIOStandbyState() != 0 || dw.getIOStandbyTermination() != 0 {
			// PAD_CFG_GPI_SCI_IOS(GPIO_141, NONE, DEEP, EDGE_SINGLE, INVERT, IGNORE, DISPUPD),
			macro.add("_SCI_IOS")
			macro.add("(").id().pull().rstsrc().trig().invert().iosstate().ioterm()
		} else if (isEdge & 0x1) != 0 {
			// e.g. PAD_CFG_GPI_ACPI_SCI(GPP_G2, NONE, DEEP, YES),
			macro.add("_ACPI_SCI").add("(").id().pull().rstsrc().invert()
		} else {
			// e.g. PAD_CFG_GPI_SCI(GPP_B18, UP_20K, PLTRST, LEVEL, INVERT),
			macro.add("_SCI").add("(").id().pull().rstsrc().trig().invert()
		}

	case dw.getGPIOInputRouteSMI():

		if dw.getIOStandbyState() != 0 || dw.getIOStandbyTermination() != 0 {
			// PAD_CFG_GPI_SMI_IOS(GPIO_41, UP_20K, DEEP, EDGE_SINGLE, NONE, IGNORE, SAME),
			macro.add("_SMI_IOS")
			macro.add("(").id().pull().rstsrc().trig().invert().iosstate().ioterm()
		} else if (isEdge & 0x1) != 0 {
			// e.g. PAD_CFG_GPI_ACPI_SMI(GPP_I3, NONE, DEEP, YES),
			macro.add("_ACPI_SMI").add("(").id().pull().rstsrc().invert()
		} else {
			// e. g. PAD_CFG_GPI_SMI(GPP_E3, NONE, PLTRST, EDGE_SINGLE, NONE),
			macro.add("_SMI").add("(").id().pull().rstsrc().trig().invert()
		}

	case isEdge != 0:
		// e.g. ISH_SPI_CS#
		// PAD_CFG_GPI_INT(GPP_D9, NONE, PLTRST, EDGE),
		macro.add("_INT").add("(").id().pull().rstsrc().trig()

	default:
		// no common
		// e.g. PAD_CFG_GPI(GPP_A7, NONE, DEEP),
		macro.add("(").id().pull().rstsrc()
		// end
	}
}

// Adds suffix to GPO macro with arguments
func (macro *macro) addSuffixOutput() {
	dw := macro.dw()
	term := dw.getTermination()
	if dw.getGPIORxTxDisableStatus() == 3 {
		// if Rx and Tx buffers are disabled
		// e.g. PAD_CFG_GPIO_HI_Z(GPIO_91, NATIVE, DEEP, IGNORE, SAME),
		macro.add("_GPO_HI_Z").id().pull().rstsrc().iosstate().ioterm()
	}
	// FIXME: add _DRIVER(..)
	if term != 0 {
		// e.g. PAD_CFG_TERM_GPO(GPP_B23, 1, DN_20K, DEEP),
		macro.add("_TERM")
	}
	macro.add("_GPO")
	if dw.getIOStandbyState() != 0 || dw.getIOStandbyTermination() != 0 {
		// e.g. FST_SPI_CS1_B -- SPK_PA_EN_R
		// PAD_CFG_GPO_IOSSTATE_IOSTERM(GPIO_91, 0, DEEP, NONE, Tx0RxDCRx0, DISPUPD),
		macro.add("_IOSSTATE").add("_IOSTERM")
	}
	macro.add("(").id().val()
	if term != 0 {
		macro.pull()
	}
	macro.rstsrc()
	if dw.getIOStandbyState() != 0 || dw.getIOStandbyTermination() != 0 {
		// e.g. FST_SPI_CS1_B -- SPK_PA_EN_R
		// PAD_CFG_GPO_IOSSTATE_IOSTERM(GPIO_91, 0, DEEP, NONE, Tx0RxDCRx0, DISPUPD),
		macro.iosstate().ioterm()
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

	if val := int(macro.dw().getRxTxEnableConfig()); val != 0 {
		macro.add(" | (").add(strconv.Itoa(val)).add(") << 21 | ")
	}

	if macro.dw().getGPIOTXState() != 0 {
		macro.add(" | 1")
	}

	macro.add(",\n\t\tPAD_PULL(").pull().add(") | PAD_IOSSTATE(").iosstate()
	macro.add(") | PAD_IOSTERM(").ioterm().add(")),")
	return macro
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
		isIOStandbyStateUsed := dw.getIOStandbyState() != 0
		isIOStandbyTerminationUsed := dw.getIOStandbyTermination() != 0
		macro.add("_NF")
		if !isEdge || !isTxRxBufDis {
			if !isIOStandbyTerminationUsed {
				// PAD_CFG_NF_IOSSTATE(GPIO_22, UP_20K, DEEP, NF2, TxDRxE),
				macro.add("_IOSSTATE")
				macro.add("(").id().pull().rstsrc().padfn().iosstate().add("),")
				macro.check()
				return macro.get()
			}
			// e.g. PAD_CFG_NF(GPP_D23, NONE, DEEP, NF1)
			macro.add("(").id().pull().rstsrc().padfn()
		} else {
			macro.add("_BUF")
			if !isIOStandbyStateUsed && !isIOStandbyTerminationUsed {
				// PAD_CFG_NF_BUF_TRIG(GPP_B23, 20K_PD, PLTRST, NF2, RX_DIS, OFF),
				macro.add("_TRIG").add("(").id().pull().rstsrc().padfn().bufdis().trig()
			} else {
				// PAD_CFG_NF_BUF_IOSSTATE_IOSTERM(GPIO_103, NATIVE, DEEP, NF1,
				//         TX_DISABLE, MASK, SAME),
				macro.add("_IOSSTATE_IOSTERM").add("(").id().pull().rstsrc().padfn()
				macro.bufdis().iosstate().ioterm()
			}
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
func GenMacro(id string, dw0 uint32, dw1 uint32) string {
	macro := macro{padID: id, dwcfg: dwcfg{reg: [MaxDWNum]uint32{dw0, dw1}}}
	return macro.generate()
}
