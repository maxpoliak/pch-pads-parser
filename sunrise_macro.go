package main

import (
	"fmt"
	"strconv"
)

type macro struct {
	padID string
	dw    *configData
	str   string
}

func (macro *macro) initData(padID string, data *configData) *macro {
	macro.padID = padID
	macro.dw = data
	return macro
}

func (macro *macro) getData() *configData {
	return macro.dw
}

func (macro *macro) add(str string) *macro {
	macro.str += str
	return macro
}

func (macro *macro) set(str string) *macro {
	macro.str = str
	return macro
}

// Adds PAD Id to the macro as a new argument
// return: macro
func (macro *macro) id() *macro {
	return macro.add(macro.padID)
}

// Add separator to macro if needed
func (macro *macro) separator() *macro {
	c := macro.str[len(macro.str)-1]
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
	str, valid := resetSrc[macro.getData().getResetConfig()]
	if !valid {
		// 3h = Reserved (implement as setting 0h) for Sunrice PCH
		str = resetSrc[0]
		fmt.Println("Warning:",
			macro.padID,
			"PADRSTCFG = 3h Reserved")
		fmt.Println("It is implemented as setting 0h (PWROK) for Sunrise PCH")
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
		0xf: "NATIVE"}
	str, valid := pull[macro.getData().getTermination()]
	if !valid {
		str = "INVALID"
		fmt.Println("Error",
			macro.padID,
			" invalid TERM value = ",
			int(macro.getData().getTermination()))
	}
	return macro.separator().add(str)
}

// Adds Pad GPO value to macro string as a new argument
// return: macro
func (macro *macro) val() *macro {
	return macro.separator().add(strconv.Itoa(macro.getData().getGPIOTXState()))
}

// Adds Pad GPO value to macro string as a new argument
// return: macro
func (macro *macro) trig() *macro {
	var dw = macro.getData()
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
	if macro.getData().getRXLevelConfiguration() {
		macro.add("YES")
	} else {
		macro.add("NONE")
	}
	return macro
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
	state := macro.getData().getGPIORxTxDisableStatus()
	return macro.separator().add(buffDisStat[state])
}

//Adds pad native function (PMODE) as a new argument
//return: macro
func (macro *macro) padfn() *macro {
	nfnum := int(macro.getData().getPadMode())
	if nfnum != 0 {
		return macro.separator().add("NF"+strconv.Itoa(nfnum))
	}
	// GPIO used only for PAD_FUNC(x) macro
	return macro.add("GPIO")
}

// Adds suffix to GPI macro with arguments
func (macro *macro) addSuffixInput() {
	dw := macro.getData()
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
		macro.add("_SMI").add("(").id().pull().rstsrc().invert()

	case isEdge != 0:
		// e.g. ISH_SPI_CS#
		// PAD_CFG_GPI_INT(GPP_D9, NONE, PLTRST, EDGE),
		macro.add("_INT").add("(").id().pull().rstsrc().trig()

	default:
		// e.g. PAD_CFG_GPI(GPP_A7, NONE, DEEP),
		macro.add("(").id().pull().rstsrc()
	}
}

// Adds suffix to GPO macro with arguments
func (macro *macro) addSuffixOutput() {
	term := macro.getData().getTermination()
	// FIXME: don`t understand how to get PAD_CFG_GPI_GPIO_DRIVER(..)
	if term != 0 {
		// e.g. PAD_CFG_TERM_GPO(GPP_B23, 1, DN_20K, DEEP),
		macro.add("_TERM")
	}
	macro.add("_GPO").add("(").id().val()
	if term != 0 {
		macro.pull()
	}
	macro.rstsrc()

	// Fix mask for RX Level/Edge Configuration (RXEVCFG)
	// See https://github.com/coreboot/coreboot/commit/3820e3c
	macro.getData().maskTrigFix()
}

const (
	RX_DIS uint8 = 0x2
	TX_DIS uint8 = 0x1
)

// Set common macros macros if the standard ones can't define the values from
// the DW0/1 register
func (macro *macro) common() *macro {
	macro.set("_PAD_CFG_STRUCT(").id().
		add(",\n\t\tPAD_FUNC(").padfn().
		add(") | PAD_RESET(").rstsrc().add(") |\n\t\t")

	var str string
	var argument = true
	switch dw := macro.getData(); {
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

	return macro.add(" |\n\t\t").add("PAD_BUF(").bufdis().add("),").
		add("\n\t\tPAD_PULL(").pull().add(")),")
}

// Check created macro
func (macro *macro) check() {
	if !macro.getData().maskCheck(0) {
		macro.common()
		// Debug message about this
		fmt.Printf(
			"\nCoreboot macros can't define the configuration for pad: %s\n",
			macro.padID)
		fmt.Printf("Use %s\n", macro.str)
	}
}

// Gets base string of current macro
// return: string of macro
//         error
func (macro *macro) getBase() string {
	dw := macro.getData()
	macro.set("PAD_CFG")
	if dw.getPadMode() == 0 {
		/* GPIO */
		switch dw.getGPIORxTxDisableStatus() {
		case TX_DIS:
			/* GPI */
			macro.addSuffixInput()

		case RX_DIS:
			/* GPO */
			macro.addSuffixOutput()

		case RX_DIS | TX_DIS:
			/* NC */
			macro.set("PAD_NC").add("(").id().pull()
			// Fix mask for RX Level/Edge Configuration (RXEVCFG)
			// and Pad Reset Config (PADRSTCFG)
			// There is no need to check these fields if the pad
			// is in the NC state
			macro.getData().maskResetFix().maskTrigFix()

		default:
			// In case the rule isn't found, a common macro is used
			// to create the pad configuration
			return macro.common().str
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
	return macro.str
}

// Get pad macro
// return: string of macro
//         error
func getMacro(padID string, data *configData) string {
	var macro macro
	macro.initData(padID, data)
	return macro.getBase()
}
