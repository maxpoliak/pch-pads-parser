package sunrise

import (
	"strings"
	"fmt"
)

// Local packages
import "../config"
import "../common"

const (
	PAD_CFG_DW0_RO_FIELDS = (0x1 << 27) | (0x1 << 24) | (0x3 << 21) | (0xf << 16) | 0xfe
	PAD_CFG_DW1_RO_FIELDS = 0xfffffc3f
)

const (
	PAD_CFG_DW0 = common.PAD_CFG_DW0
	PAD_CFG_DW1 = common.PAD_CFG_DW1
	MAX_DW_NUM  = common.MAX_DW_NUM
)

type PlatformSpecific struct {}

// Adds the PADRSTCFG parameter from PAD_CFG_DW0 to the macro as a new argument
func (PlatformSpecific) Rstsrc(macro *common.Macro) {
	dw0 := macro.Register(PAD_CFG_DW0)

	if config.IsPlatformSunrise() && strings.Contains(macro.PadIdGet(), "GPD") {
		// The pads from the GPD group in the Sunrise point PCH have a specific Pad 
		// Reset Config Map, which differs from that in other groups. PADRSTCFG field
		// in PAD_CFG_DW0 register should implements 3h value as RSMRST and 0h as PWROK
		var gpdResetSrcMap = map[uint8]string{
			0x0: "PWROK",
			0x1: "DEEP",
			0x2: "PLTRST",
			0x3: "RSMRST",
		}
		macro.Separator().Add(gpdResetSrcMap[dw0.GetResetConfig()])
		return
	}

	var resetSrcMap = map[uint8]string{
		0x0: "RSMRST",
		0x1: "DEEP",
		0x2: "PLTRST",
	}
	str, valid := resetSrcMap[dw0.GetResetConfig()]
	if !valid {
		str = "RESERVED"
	}
	macro.Separator().Add(str)
}

// Adds The Pad Termination (TERM) parameter from PAD_CFG_DW1 to the macro
// as a new argument
func (PlatformSpecific) Pull(macro *common.Macro) {
	dw1 := macro.Register(PAD_CFG_DW1)
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
	dw0 := macro.Register(PAD_CFG_DW0)
	isEdge := dw0.GetRXLevelEdgeConfiguration()
	macro.Add("_GPI")
	switch {
	case dw0.GetGPIOInputRouteNMI():
		// e.g. PAD_CFG_GPI_NMI(GPIO_24, UP_20K, DEEP, LEVEL, INVERT),
		macro.Add("_NMI").Add("(").Id().Pull().Rstsrc().Trig().Invert()

	case dw0.GetGPIOInputRouteIOxAPIC():
		// e.g. PAD_CFG_GPI_APIC(GPP_B3, NONE, PLTRST)
		macro.Add("_APIC")
		if isEdge == 0 {
			if dw0.GetRXLevelConfiguration() {
				// e.g. PAD_CFG_GPI_APIC_INVERT(GPP_C5, DN_20K, DEEP),
				macro.Add("_INVERT")
			}
			macro.Add("(").Id().Pull().Rstsrc()
		} else {
			// PAD_CFG_GPI_APIC_IOS(GPP_C20, NONE, DEEP, LEVEL, INVERT,
			// TxLASTRxE, SAME) macro isn't used for this chipset. But, in
			// the case when SOC_INTEL_COMMON_BLOCK_GPIO_LEGACY_MACROS is
			// defined in the config, this macro allows you to set trig and
			// invert parameters.
			macro.Add("_IOS").Add("(").Id().Pull().Rstsrc().Trig().Invert()
			// IOStandby Config will be ignored for the Sunrise PCH, so use
			// any values
			macro.Add(", TxLASTRxE, SAME")
		}

	case dw0.GetGPIOInputRouteSCI():

		// e.g. PAD_CFG_GPI_SCI(GPP_B18, UP_20K, PLTRST, LEVEL, INVERT),
		if (isEdge & 0x1) != 0 {
			// e.g. PAD_CFG_GPI_ACPI_SCI(GPP_G2, NONE, DEEP, YES),
			macro.Add("_ACPI")
		}
		macro.Add("_SCI").Add("(").Id().Pull().Rstsrc()
		if (isEdge & 0x1) == 0 {
			macro.Trig()
		}
		macro.Invert()

	case dw0.GetGPIOInputRouteSMI():
		if (isEdge & 0x1) != 0 {
			// e.g. PAD_CFG_GPI_ACPI_SMI(GPP_I3, NONE, DEEP, YES),
			macro.Add("_ACPI")
		}
		macro.Add("_SMI").Add("(").Id().Pull().Rstsrc()
		if (isEdge & 0x1) == 0 {
			// e.g. PAD_CFG_GPI_SMI(GPP_E7, NONE, DEEP, LEVEL, NONE),
			macro.Trig()
		}
		macro.Invert()

	case isEdge != 0 || macro.IsOwnershipDriver():
		// e.g. PAD_CFG_GPI_TRIG_OWN(pad, pull, rst, trig, own)
		macro.Add("_TRIG_OWN").Add("(").Id().Pull().Rstsrc().Trig().Own()

	default:
		if macro.IsOwnershipDriver() {
			// e. g. PAD_CFG_GPI_GPIO_DRIVER(GPIO_212, UP_20K, DEEP),
			macro.Add("_GPIO_DRIVER")
		}
		// e.g. PAD_CFG_GPI(GPP_A7, NONE, DEEP),
		macro.Add("(").Id().Pull().Rstsrc()
	}
	macro.Add("),")
}

// Adds PAD_CFG_GPO macro with arguments
func (PlatformSpecific) GpoMacroAdd(macro *common.Macro) {
	term := macro.Register(PAD_CFG_DW1).GetTermination()
	// FIXME: don`t understand how to get PAD_CFG_GPI_GPIO_DRIVER(..)
	if term != 0 {
		// e.g. PAD_CFG_TERM_GPO(GPP_B23, 1, DN_20K, DEEP),
		macro.Add("_TERM")
	}
	macro.Add("_GPO")
	if macro.IsOwnershipDriver() {
		// PAD_CFG_GPO_GPIO_DRIVER(pad, val, rst, pull)
		macro.Add("_GPIO_DRIVER")
	}
	macro.Add("(").Id().Val()
	if term != 0 {
		macro.Pull()
	}
	macro.Rstsrc()
	if macro.IsOwnershipDriver() {
		macro.Pull()
	}
	macro.Add("),")

	// Fix mask for RX Level/Edge Configuration (RXEVCFG)
	// See https://github.com/coreboot/coreboot/commit/3820e3c
	macro.Register(PAD_CFG_DW0).MaskTrigFix()
}

// Adds PAD_CFG_NF macro with arguments
func (PlatformSpecific) NativeFunctionMacroAdd(macro *common.Macro) {
	dw0 := macro.Register(PAD_CFG_DW0)
	isEdge := dw0.GetRXLevelEdgeConfiguration() != 0
	isTxRxBufDis := dw0.GetGPIORxTxDisableStatus() != 0
	// e.g. PAD_CFG_NF(GPP_D23, NONE, DEEP, NF1)
	macro.Add("_NF")
	if isEdge || isTxRxBufDis {
		// e.g. PCHHOT#
		// PAD_CFG_NF_BUF_TRIG(GPP_B23, 20K_PD, PLTRST, NF2, RX_DIS, OFF),
		macro.Add("_BUF_TRIG")
	}
	macro.Add("(").Id().Pull().Rstsrc().Padfn()
	if isEdge || isTxRxBufDis {
		macro.Bufdis().Trig()
	}
	macro.Add("),")
}

// Generates "Advanced" macro if the standard ones can't define the values from
// the DW0/1 register
func (PlatformSpecific) AdvancedMacroGenerate(macro *common.Macro) {
	macro.Set("_PAD_CFG_STRUCT(").Id().Add(",\n\t\tPAD_FUNC(").Padfn().
		Add(") | PAD_RESET(").Rstsrc().Add(") |\n\t\t")

	var str string
	var argument = true

	switch dw0 := macro.Register(PAD_CFG_DW0); {
	case dw0.GetGPIOInputRouteIOxAPIC():
		str = "IOAPIC"
	case dw0.GetGPIOInputRouteSCI():
		str = "SCI"
	case dw0.GetGPIOInputRouteSMI():
		str = "SMI"
	case dw0.GetGPIOInputRouteNMI():
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
	if macro.Register(PAD_CFG_DW0).GetGPIOTXState() != 0 {
		macro.Add(" | 1")
	}
	macro.Add(",\n\t\tPAD_CFG_OWN_GPIO(").Own().Add(") | ").Add("PAD_PULL(").Pull().Add(")),")

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
