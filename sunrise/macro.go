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

const (
	TRIG_LEVEL       = 0
	TRIG_EDGE_SINGLE = 1
	TRIG_OFF         = 2
	TRIG_EDGE_BOTH   = 3
)

type PlatformSpecific struct {}

// Adds the PADRSTCFG parameter from PAD_CFG_DW0 to the macro as a new argument
func (PlatformSpecific) Rstsrc(macro *common.Macro) {
	dw0 := macro.Register(PAD_CFG_DW0)

	if config.IsPlatformSunrise() && strings.Contains(macro.PadIdGet(), "GPD") {
		// The pads from the GPD group in the Sunrise Point PCH have a specific Pad
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

// Generate macro to cause peripheral IRQ when configured in GPIO input mode
func ioApicRoute(macro *common.Macro) bool {
	dw0 := macro.Register(PAD_CFG_DW0)
	if dw0.GetGPIOInputRouteIOxAPIC() == 0 {
		return false
	}
	// e.g. PAD_CFG_GPI_APIC(GPP_B3, NONE, PLTRST)
	macro.Add("_APIC")
	if dw0.GetRXLevelEdgeConfiguration() == 0 {
		if dw0.GetRXLevelConfiguration() != 0 {
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
	macro.Add("),")
	return true
}

// Generate macro to cause NMI when configured in GPIO input mode
func nmiRoute(macro *common.Macro) bool {
	if macro.Register(PAD_CFG_DW0).GetGPIOInputRouteNMI() == 0 {
		return false
	}
	// PAD_CFG_GPI_NMI(GPIO_24, UP_20K, DEEP, LEVEL, INVERT),
	macro.Add("_NMI").Add("(").Id().Pull().Rstsrc().Trig().Invert().Add("),")
	return true
}

// Generate macro to cause SCI when configured in GPIO input mode
func sciRoute(macro *common.Macro) bool {
	dw0 := macro.Register(PAD_CFG_DW0)
	if dw0.GetGPIOInputRouteSCI() == 0 {
		return false
	}
	// e.g. PAD_CFG_GPI_SCI(GPP_B18, UP_20K, PLTRST, LEVEL, INVERT),
	if (dw0.GetRXLevelEdgeConfiguration() & TRIG_EDGE_SINGLE) != 0 {
		// e.g. PAD_CFG_GPI_ACPI_SCI(GPP_G2, NONE, DEEP, YES),
		// #define PAD_CFG_GPI_ACPI_SCI(pad, pull, rst, inv)	\
		//             PAD_CFG_GPI_SCI(pad, pull, rst, EDGE_SINGLE, inv)
		macro.Add("_ACPI")
	}
	macro.Add("_SCI").Add("(").Id().Pull().Rstsrc()
	if (dw0.GetRXLevelEdgeConfiguration() & TRIG_EDGE_SINGLE) == 0 {
		macro.Trig()
	}
	macro.Invert().Add("),")
	return true
}

// Generate macro to cause SMI when configured in GPIO input mode
func smiRoute(macro *common.Macro) bool {
	dw0 := macro.Register(PAD_CFG_DW0)
	if dw0.GetGPIOInputRouteSMI() == 0 {
		return false
	}
	if (dw0.GetRXLevelEdgeConfiguration() & TRIG_EDGE_SINGLE) != 0 {
		// e.g. PAD_CFG_GPI_ACPI_SMI(GPP_I3, NONE, DEEP, YES),
		macro.Add("_ACPI")
	}
	macro.Add("_SMI").Add("(").Id().Pull().Rstsrc()
	if (dw0.GetRXLevelEdgeConfiguration() & TRIG_EDGE_SINGLE) == 0 {
		// e.g. PAD_CFG_GPI_SMI(GPP_E7, NONE, DEEP, LEVEL, NONE),
		macro.Trig()
	}
	macro.Invert().Add("),")
	return true
}

// Adds PAD_CFG_GPI macro with arguments
func (PlatformSpecific) GpiMacroAdd(macro *common.Macro) {
	var ids []string

	macro.Set("PAD_CFG_GPI")
	for routeid, isRoute := range map[string]func(macro *common.Macro) (bool) {
		"IOAPIC": ioApicRoute,
		"SCI":    sciRoute,
		"SMI":    smiRoute,
		"NMI":    nmiRoute,
	} {
		if isRoute(macro) {
			ids = append(ids, routeid)
		}
	}

	if argc := len(ids); argc == 0 {
		// e.g. PAD_CFG_GPI_TRIG_OWN(pad, pull, rst, trig, own)
		macro.Add("_TRIG_OWN").Add("(").Id().Pull().Rstsrc().Trig().Own().Add("),")
	} else if argc == 2 {
		// PAD_CFG_GPI_DUAL_ROUTE(pad, pull, rst, trig, inv, route1, route2)
		macro.Set("PAD_CFG_GPI_DUAL_ROUTE(").Id().Pull().Rstsrc().Trig()
		macro.Add(", " + ids[0] + ", " + ids[1] + "),")
	} else {
		// Clear the control mask so that the check fails and "Advanced" macro is
		// generated
		macro.Register(PAD_CFG_DW0).ClearCntrMask()
	}
}

// Adds PAD_CFG_GPO macro with arguments
func (PlatformSpecific) GpoMacroAdd(macro *common.Macro) {
	term := macro.Register(PAD_CFG_DW1).GetTermination()

	macro.Set("PAD_CFG")
	// FIXME: don`t understand how to get PAD_CFG_GPI_GPIO_DRIVER(..)
	if term != 0 {
		// e.g. PAD_CFG_TERM_GPO(GPP_B23, 1, DN_20K, DEEP),
		macro.Add("_TERM")
	}
	macro.Add("_GPO").Add("(").Id().Val()
	if term != 0 {
		macro.Pull()
	}
	macro.Rstsrc().Add("),")

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
	macro.Set("PAD_CFG_NF")
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

// Adds PAD_NC macro
func (PlatformSpecific) NoConnMacroAdd(macro *common.Macro) {
	// PAD_NC(OSC_CLK_OUT_1, DN_20K)
	macro.Set("PAD_NC").Add("(").Id().Pull().Add("),")
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
