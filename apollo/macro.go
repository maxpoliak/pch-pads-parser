package apollo

import "fmt"

// Local packages
import "../common"

const (
	PAD_CFG_DW0_RO_FIELDS = (0x1 << 27) | (0x1 << 24) | (0x3 << 21) | (0xf << 16) | 0xfe
	PAD_CFG_DW1_RO_FIELDS = 0xfffc000f
)

const (
	PAD_CFG_DW0 = common.PAD_CFG_DW0
	PAD_CFG_DW1 = common.PAD_CFG_DW1
	MAX_DW_NUM  = common.MAX_DW_NUM
)

type PlatformSpecific struct {}

// Adds the PADRSTCFG parameter from DW0 to the macro as a new argument
// return: macro
func (PlatformSpecific) Rstsrc(macro *common.Macro) {
	dw0 := macro.Register(PAD_CFG_DW0)
	var resetSrcMap = map[uint8]string{
		0x0: "PWROK",
		0x1: "DEEP",
		0x2: "PLTRST",
	}
	str, valid := resetSrcMap[dw0.GetResetConfig()]
	if !valid {
			str = "RESERVED"
	}
	macro.Separator().Add(str)
}

// Adds The Pad Termination (TERM) parameter from DW1 to the macro as a new argument
// return: macro
func (PlatformSpecific) Pull(macro *common.Macro) {
	dw1 := macro.Register(PAD_CFG_DW1)
	var pull = map[uint8]string{
		0x0: "NONE",
		0x2: "DN_5K",
		0x4: "DN_20K",
		0x9: "UP_1K",
		0xa: "UP_5K",
		0xb: "UP_2K",
		0xc: "UP_20K",
		0xd: "UP_667",
		0xf: "NATIVE",
	}
	str, valid := pull[dw1.GetTermination()]
	if !valid {
		str = "INVALID"
		fmt.Println("Error",
			macro.PadIdGet(),
			" invalid TERM value = ",
			int(dw1.GetTermination()))
	}
	macro.Separator().Add(str)
}

// Adds PAD_CFG_GPI macro with arguments
func (PlatformSpecific) GpiMacroAdd(macro *common.Macro) {
	dw0 :=  macro.Register(PAD_CFG_DW0)
	dw1 :=  macro.Register(PAD_CFG_DW1)
	isEdge := dw0.GetRXLevelEdgeConfiguration()
	macro.Add("_GPI")
	switch {
	case dw0.GetGPIOInputRouteNMI() != 0:
		// e.g. PAD_CFG_GPI_NMI(GPIO_24, UP_20K, DEEP, LEVEL, INVERT),
		macro.Add("_NMI").Add("(").Id().Pull().Rstsrc().Trig().Invert()

	case dw0.GetGPIOInputRouteIOxAPIC() != 0:
		// e.g. PAD_CFG_GPI_APIC(GPP_B3, NONE, PLTRST)
		macro.Add("_APIC")
		if dw1.GetIOStandbyState() != 0 || dw1.GetIOStandbyTermination() != 0 {
			// e.g. H1_PCH_INT_ODL
			// PAD_CFG_GPI_APIC_IOS(GPIO_63, NONE, DEEP, LEVEL, INVERT, TxDRxE, DISPUPD),
			macro.Add("_IOS")
			macro.Add("(").Id().Pull().Rstsrc().Trig().Invert().IOSstate().IOTerm()
		} else {
			if dw0.GetRXLevelConfiguration() != 0 {
				// e.g. PAD_CFG_GPI_APIC_INVERT(GPP_C5, DN_20K, DEEP),
				macro.Add("_INVERT")
			}
			// PAD_CFG_GPI_APIC(GPP_C5, DN_20K, DEEP),
			macro.Add("(").Id().Pull().Rstsrc()
		}

	case dw0.GetGPIOInputRouteSCI() != 0:
		if dw0.GetIOStandbyState() != 0 || dw0.GetIOStandbyTermination() != 0 {
			// PAD_CFG_GPI_SCI_IOS(GPIO_141, NONE, DEEP, EDGE_SINGLE, INVERT, IGNORE, DISPUPD),
			macro.Add("_SCI_IOS")
			macro.Add("(").Id().Pull().Rstsrc().Trig().Invert().IOSstate().IOTerm()
		} else if (isEdge & 0x1) != 0 {
			// e.g. PAD_CFG_GPI_ACPI_SCI(GPP_G2, NONE, DEEP, YES),
			macro.Add("_ACPI_SCI").Add("(").Id().Pull().Rstsrc().Invert()
		} else {
			// e.g. PAD_CFG_GPI_SCI(GPP_B18, UP_20K, PLTRST, LEVEL, INVERT),
			macro.Add("_SCI").Add("(").Id().Pull().Rstsrc().Trig().Invert()
		}

	case dw0.GetGPIOInputRouteSMI() != 0:
		if dw1.GetIOStandbyState() != 0 || dw1.GetIOStandbyTermination() != 0 {
			// PAD_CFG_GPI_SMI_IOS(GPIO_41, UP_20K, DEEP, EDGE_SINGLE, NONE, IGNORE, SAME),
			macro.Add("_SMI_IOS")
			macro.Add("(").Id().Pull().Rstsrc().Trig().Invert().IOSstate().IOTerm()
		} else if (isEdge & 0x1) != 0 {
			// e.g. PAD_CFG_GPI_ACPI_SMI(GPP_I3, NONE, DEEP, YES),
			macro.Add("_ACPI_SMI").Add("(").Id().Pull().Rstsrc().Invert()
		} else {
			// e.g. PAD_CFG_GPI_SMI(GPP_E3, NONE, PLTRST, EDGE_SINGLE, NONE),
			macro.Add("_SMI").Add("(").Id().Pull().Rstsrc().Trig().Invert()
		}

	default:
		// e.g. PAD_CFG_GPI_TRIG_OWN(pad, pull, rst, trig, own)
		macro.Add("_TRIG_OWN").Add("(").Id().Pull().Rstsrc().Trig().Own()
	}
	macro.Add("),")
}

// Adds PAD_CFG_GPO macro with arguments
func (PlatformSpecific) GpoMacroAdd(macro *common.Macro) {
	dw0 :=  macro.Register(PAD_CFG_DW0)
	dw1 :=  macro.Register(PAD_CFG_DW1)
	term := dw1.GetTermination()
	if dw1.GetIOStandbyState() != 0 || dw1.GetIOStandbyTermination() != 0 {
		// PAD_CFG_GPO_IOSSTATE_IOSTERM(GPIO_91, 0, DEEP, NONE, Tx0RxDCRx0, DISPUPD),
		macro.Add("_IOSSTATE_IOSTERM(").Id().Val().Pull().IOSstate().IOTerm()
	} else {
		if term != 0 {
			// e.g. PAD_CFG_TERM_GPO(GPP_B23, 1, DN_20K, DEEP),
			macro.Add("_TERM")
		}
		macro.Add("_GPO(").Id().Val()
		if term != 0 {
			macro.Pull()
		}
		macro.Rstsrc().Add("),")
	}
	// Fix mask for RX Level/Edge Configuration (RXEVCFG)
	// See https://github.com/coreboot/coreboot/commit/3820e3c
	dw0.MaskTrigFix()
}

// Adds PAD_CFG_NF macro with arguments
func (PlatformSpecific) NativeFunctionMacroAdd(macro *common.Macro) {
	dw1 := macro.Register(PAD_CFG_DW1)
	isIOStandbyStateUsed := dw1.GetIOStandbyState() != 0
	isIOStandbyTerminationUsed := dw1.GetIOStandbyTermination() != 0
	macro.Add("_NF")
	if !isIOStandbyTerminationUsed && isIOStandbyStateUsed {
		if dw1.GetIOStandbyState() == common.StandbyIgnore {
			// PAD_CFG_NF_IOSTANDBY_IGNORE(PMU_SLP_S0_B, NONE, DEEP, NF1),
			macro.Add("_IOSTANDBY_IGNORE(").Id().Pull().Rstsrc().Padfn()
		} else {
			// PAD_CFG_NF_IOSSTATE(GPIO_22, UP_20K, DEEP, NF2, TxDRxE),
			macro.Add("_IOSSTATE(").Id().Pull().Rstsrc().Padfn().IOSstate()
		}
	} else if isIOStandbyTerminationUsed {
		// PAD_CFG_NF_IOSSTATE_IOSTERM(GPIO_103, NATIVE, DEEP, NF1, MASK, SAME),
		macro.Add("_IOSSTATE_IOSTERM(").Id().Pull().Rstsrc().Padfn().IOSstate().IOTerm()
	} else {
		// e.g. PAD_CFG_NF(GPP_D23, NONE, DEEP, NF1)
		macro.Add("(").Id().Pull().Rstsrc().Padfn()
	}
	macro.Add("),")
}

// Adds PAD_NC macro
func (PlatformSpecific) NoConnMacroAdd(macro *common.Macro) {
	iosstate := macro.Register(PAD_CFG_DW1).GetIOStandbyState()
	if iosstate == common.TxDRxE {
		// PAD_NC(OSC_CLK_OUT_1, DN_20K)
		macro.Set("PAD_NC").Add("(").Id().Pull().Add("),")
		return
	}
	// PAD_CFG_GPIO_HI_Z(GPIO_81, UP_20K, DEEP, HIZCRx0, DISPUPD),
	macro.Set("PAD_CFG_GPIO_")
	if macro.IsOwnershipDriver() {
		// PAD_CFG_GPIO_DRIVER_HI_Z(GPIO_55, UP_20K, DEEP, HIZCRx1, ENPU),
		macro.Add("DRIVER_")
	}
	macro.Add("HI_Z(").Id().Pull().Rstsrc().IOSstate().IOTerm().Add("),")
}

// Generates "Advanced" macro if the standard ones can't define the values from
// the DW0/1 register
func (PlatformSpecific) AdvancedMacroGenerate(macro *common.Macro) {
	dw0 := macro.Register(PAD_CFG_DW0)
	ioApicRoute := dw0.GetGPIOInputRouteIOxAPIC() != 0
	nmiRoute := dw0.GetGPIOInputRouteNMI() != 0
	sciRoute := dw0.GetGPIOInputRouteSCI() != 0
	smiRoute := dw0.GetGPIOInputRouteSMI() != 0

	macro.Set("_PAD_CFG_STRUCT(").Id().
		Add(",\n\t\tPAD_FUNC(").Padfn().
		Add(") | PAD_RESET(").Rstsrc().Add(") |\n\t\t")

	var str string
	var argument = true
	switch {
	case ioApicRoute:
		str = "IOAPIC"
	case nmiRoute:
		str = "SCI"
	case sciRoute:
		str = "SMI"
	case smiRoute:
		str = "NMI"
	default:
		argument = false
	}

	if argument {
		macro.Add("PAD_IRQ_CFG(").Add(str).Trig().Invert().Add(")")
	} else {
		macro.Add("PAD_CFG0_TRIG_").Trig().Add(" | PAD_CFG0_RX_POL_").Invert()
	}

	macro.Add(" |\n\t\t").Add("PAD_BUF(").Bufdis().Add(")")

	if int(dw0.GetRxTxEnableConfig()) != 0 {
		macro.Add(" | (1 << 21)")
	}

	if dw0.GetGPIOTXState() != 0 {
		macro.Add(" | 1")
	}
	macro.Add(",\n\t\tPAD_CFG_OWN_GPIO(").Own().Add(") | ")
	macro.Add("PAD_PULL(").Pull().Add(") |\n\t\tPAD_IOSSTATE(").IOSstate()
	macro.Add(") | PAD_IOSTERM(").IOTerm().Add(")),")

	fmt.Printf("\nNo configuration for pad: %s\n", macro.PadIdGet())
	fmt.Printf("Use %s\n", macro.Get())
}

// GenMacro - generate pad macro
// dw0 : DW0 config register value
// dw1 : DW1 config register value
// return: string of macro
//         error
func (PlatformSpecific) GenMacro(id string, dw0 uint32, dw1 uint32, ownership uint8) string {
	var macro common.Macro
	// use platform-specific interface in Macro struct
	macro.Platform = PlatformSpecific {}
	macro.PadIdSet(id).SetPadOwnership(ownership)
	macro.Register(PAD_CFG_DW0).ValueSet(dw0).ReadOnlyFieldsSet(PAD_CFG_DW0_RO_FIELDS)
	macro.Register(PAD_CFG_DW1).ValueSet(dw1).ReadOnlyFieldsSet(PAD_CFG_DW1_RO_FIELDS)
	return macro.Generate()
}
