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
	macro.add(macro.padID)
	return macro
}

// Adds the PADRSTCFG parameter from DW0 to the macro as a new argument
// return: macro
func (macro *macro) rstsrc() *macro {
	var dw = macro.getData()
	var resetSrc = map[uint8]string{
		0x0: "PWROK",
		0x1: "DEEP",
		0x2: "PLTRST",
	}
	if str, valid := resetSrc[dw.getResetConfig()]; valid {
		macro.add(", " + str)
	} else {
		// 3h = Reserved (implement as setting 0h) for sky pch
		macro.add(", PWROK")
		fmt.Println("Warning:", macro.padID, "PADRSTCFG = 3h Reserved")
		fmt.Println("It is implemented as setting 0h (PWROK) for Skylake/Kaby Lake PCH")
	}
	return macro
}

// Adds The Pad Termination (TERM) parameter from DW1 to the macro as a new argument
// return: macro
func (macro *macro) pull() *macro {
	var dw = macro.getData()
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
	if str, valid := pull[dw.getTermination()]; valid {
		macro.add(", " + str)
	} else {
		macro.add(", INVALID")
		fmt.Println("Error", macro.padID, " invalid TERM value = ",
			int(dw.getTermination()))
	}
	return macro
}

// Adds Pad GPO value to macro string as a new argument
// return: macro
func (macro *macro) val() *macro {
	var state int
	if macro.getData().getGPIOTXState() {
		state = 1
	}
	macro.add(", " + strconv.Itoa(state))
	return macro
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
	macro.add(", " + trig[dw.getRXLevelEdgeConfiguration()])
	return macro
}

// Adds Pad Polarity Inversion Stage (RXINV) to macro string as a new argument
// return: macro
func (macro *macro) invert() *macro {
	if macro.getData().getRXLevelConfiguration() {
		macro.add(", YES")
	} else {
		macro.add(", NONE")
	}
	return macro
}

//Adds pad native function (PMODE) as a new argument
//return: macro
func (macro *macro) padfn() *macro {
	var nfnum = int(macro.getData().getPadMode())
	macro.add(", NF" + strconv.Itoa(nfnum))
	return macro
}

// Adds suffix to GPI macro with arguments
func (macro *macro) addSuffixInput() {
	var dw = macro.getData()
	var isEdge = (dw.getRXLevelEdgeConfiguration() & 0x1 != 0)

	macro.add("_GPI")
	switch {
	case dw.getGPIOInputRouteNMI():
		// e.g. PAD_CFG_GPI_NMI(GPIO_24, UP_20K, DEEP, LEVEL, INVERT),
		macro.add("_NMI").add("(").id().pull().rstsrc().trig().invert()

	case dw.getGPIOInputRouteIOxAPIC():
		// e.g. PAD_CFG_GPI_APIC(GPP_B3, NONE, PLTRST)
		macro.add("_APIC")
		if dw.getRXLevelConfiguration() {
			// e.g. PAD_CFG_GPI_APIC_INVERT(GPP_C5, DN_20K, DEEP),
			macro.add("_INVERT")
		}
		macro.add("(").id().pull().rstsrc()

	case dw.getGPIOInputRouteSCI():

		// e.g. PAD_CFG_GPI_SCI(GPP_B18, UP_20K, PLTRST, LEVEL, INVERT),
		if isEdge {
			// e.g. PAD_CFG_GPI_ACPI_SCI(GPP_G2, NONE, DEEP, YES),
			macro.add("_ACPI")
		}
		macro.add("_SCI").add("(").id().pull().rstsrc()
		if !isEdge {
			macro.trig()
		}
		macro.invert()

	case dw.getGPIOInputRouteSMI():
		if isEdge {
			// e.g. PAD_CFG_GPI_ACPI_SMI(GPP_I3, NONE, DEEP, YES),
			macro.add("_ACPI")
		}
		macro.add("_SMI").add("(").id().pull().rstsrc().invert()

	default:
		// e.g. PAD_CFG_GPI(GPP_A7, NONE, DEEP),
		macro.add("(").id().pull().rstsrc()
	}
}

// Adds suffix to GPO macro with arguments
func (macro *macro) addSuffixOutput() {
	// FIXME: don`t understand how to get PAD_CFG_GPI_GPIO_DRIVER(..)
	if macro.getData().getTermination() != 0 {
		// e.g. PAD_CFG_TERM_GPO(GPP_B23, 1, DN_20K, DEEP),
		macro.add("_TERM")
	}
	macro.add("_GPO").add("(").id().val().rstsrc()
}

const (
	RX_DIS uint8 = 0x2
	TX_DIS uint8 = 0x1
)

// Gets base string of current macro
// return: string of macro
//         error
func (macro *macro) getBase() (string, error) {
	var dw = macro.getData()
	var err error

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
		default:
			macro.set("PAD_INVALID").add("(").id()
			err = fmt.Errorf("Error: Missing template for creating macros")
		}
	} else {
		/* PAD NF e.g. PAD_CFG_NF(GPP_D23, NONE, DEEP, NF1) */
		macro.add("_NF").add("(").id().pull().rstsrc().padfn()
		err = nil
	}
	macro.add("),")
	return macro.str, err
}

// Get pad macro
// return: string of macro
//         error
func getMacro(padID string, data *configData) (string, error) {
	var macro macro
	macro.initData(padID, data)
	return macro.getBase()
}
