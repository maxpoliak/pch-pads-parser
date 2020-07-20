package sunrise

import (
	"strings"
	"fmt"
)

// Local packages
import "../common"
import "../config"
import "../fields/cb"

const (
	PAD_CFG_DW0_RO_FIELDS = (0x1 << 27) | (0x1 << 24) | (0x3 << 21) | (0xf << 16) | 0xfe
	PAD_CFG_DW1_RO_FIELDS = 0xffffc3ff
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
	var resetsrc = map[uint8]string{
		0: "PWROK",
		1: "DEEP",
		2: "PLTRST",
		3: "RSMRST",
	}

	if config.TemplateGet() != config.TempInteltool {
		// Use reset source remapping only if the input file is inteltool.log dump
		macro.Separator().Add(resetsrc[dw0.GetResetConfig()])
		return
	}

	if config.IsPlatformSunrise() && strings.Contains(macro.PadIdGet(), "GPD") {
		// See reset map for the GPD Group in the Community 2:
		// https://github.com/coreboot/coreboot/blob/master/src/soc/intel/skylake/gpio.c#L15
		// static const struct reset_mapping rst_map_com2[] = {
		// { .logical = PAD_CFG0_LOGICAL_RESET_PWROK,  .chipset = 0U << 30},
		// { .logical = PAD_CFG0_LOGICAL_RESET_DEEP,   .chipset = 1U << 30},
		// { .logical = PAD_CFG0_LOGICAL_RESET_PLTRST, .chipset = 2U << 30},
		// { .logical = PAD_CFG0_LOGICAL_RESET_RSMRST, .chipset = 3U << 30},
		// };
		macro.Separator().Add(resetsrc[dw0.GetResetConfig()])
		return
	}

	// For other pads of the Sunrise Point chipset, as well as for all pads of the Lewisburg
	// chipset, remapping of the card requires the port reset source value.
	// See https://github.com/coreboot/coreboot/blob/master/src/soc/intel/skylake/gpio.c#L15
	// and https://github.com/coreboot/coreboot/blob/master/src/soc/intel/xeon_sp/gpio.c#L16
	//
	// static const struct reset_mapping rst_map[] = {
	// { .logical = PAD_CFG0_LOGICAL_RESET_RSMRST, .chipset = 0U << 30 },
	// { .logical = PAD_CFG0_LOGICAL_RESET_DEEP,   .chipset = 1U << 30 },
	// { .logical = PAD_CFG0_LOGICAL_RESET_PLTRST, .chipset = 2U << 30 },
	// };
	var remapping = map[uint8]string{
		0: "RSMRST",
		1: "DEEP",
		2: "PLTRST",
	}
	str, valid := remapping[dw0.GetResetConfig()]
	if !valid {
		str = "RESERVED" // from intel doc: 3h = Reserved
		dw0.CntrMaskFieldsClear(common.PadRstCfgMask)
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

	macro.Add("_APIC")
	if dw0.GetRXLevelEdgeConfiguration() == common.TRIG_LEVEL {
		if dw0.GetRxInvert() != 0 {
			// PAD_CFG_GPI_APIC_INVERT(pad, pull, rst)
			macro.Add("_INVERT")
		}
		// PAD_CFG_GPI_APIC(pad, pull, rst)
		macro.Add("(").Id().Pull().Rstsrc().Add("),")
		return true
	}

	// e.g. PAD_CFG_GPI_APIC_IOS(pad, pull, rst, trig, inv, iosstate, iosterm)
	macro.Add("_IOS(").Id().Pull().Rstsrc().Trig().Invert().Add(", TxLASTRxE, SAME),")
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
	if (dw0.GetRXLevelEdgeConfiguration() & common.TRIG_EDGE_SINGLE) != 0 {
		// e.g. PAD_CFG_GPI_ACPI_SCI(GPP_G2, NONE, DEEP, YES),
		// #define PAD_CFG_GPI_ACPI_SCI(pad, pull, rst, inv)	\
		//             PAD_CFG_GPI_SCI(pad, pull, rst, EDGE_SINGLE, inv)
		macro.Add("_ACPI")
	}
	macro.Add("_SCI").Add("(").Id().Pull().Rstsrc()
	if (dw0.GetRXLevelEdgeConfiguration() & common.TRIG_EDGE_SINGLE) == 0 {
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
	if (dw0.GetRXLevelEdgeConfiguration() & common.TRIG_EDGE_SINGLE) != 0 {
		// e.g. PAD_CFG_GPI_ACPI_SMI(GPP_I3, NONE, DEEP, YES),
		macro.Add("_ACPI")
	}
	macro.Add("_SMI").Add("(").Id().Pull().Rstsrc()
	if (dw0.GetRXLevelEdgeConfiguration() & common.TRIG_EDGE_SINGLE) == 0 {
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

	switch argc := len(ids); argc {
	case 0:
		// e.g. PAD_CFG_GPI_TRIG_OWN(pad, pull, rst, trig, own)
		macro.Add("_TRIG_OWN").Add("(").Id().Pull().Rstsrc().Trig().Own().Add("),")
	case 1:
		// GPI with IRQ route
		if config.AreFieldsIgnored() {
			// Set Host Software Ownership to ACPI mode
			macro.SetPadOwnership(common.PAD_OWN_ACPI)
		}

	case 2:
		// PAD_CFG_GPI_DUAL_ROUTE(pad, pull, rst, trig, inv, route1, route2)
		macro.Set("PAD_CFG_GPI_DUAL_ROUTE(").Id().Pull().Rstsrc().Trig().Invert()
		macro.Add(", " + ids[0] + ", " + ids[1] + "),")
		if config.AreFieldsIgnored() {
			// Set Host Software Ownership to ACPI mode
			macro.SetPadOwnership(common.PAD_OWN_ACPI)
		}
	default:
		// Clear the control mask so that the check fails and "Advanced" macro is
		// generated
		macro.Register(PAD_CFG_DW0).CntrMaskFieldsClear(common.AllFields)
	}
}

// Adds PAD_CFG_GPO macro with arguments
func (PlatformSpecific) GpoMacroAdd(macro *common.Macro) {
	dw0 := macro.Register(PAD_CFG_DW0)
	term := macro.Register(PAD_CFG_DW1).GetTermination()

	// #define PAD_CFG_GPO(pad, val, rst)                \
	//    _PAD_CFG_STRUCT(pad,                           \
	//      PAD_FUNC(GPIO) | PAD_RESET(rst) |            \
	//      PAD_TRIG(OFF) | PAD_BUF(RX_DISABLE) | !!val, \
	//      PAD_PULL(NONE) | PAD_IOSSTATE(TxLASTRxE))
	if dw0.GetRXLevelEdgeConfiguration() != common.TRIG_OFF {
		dw0.CntrMaskFieldsClear(common.RxLevelEdgeConfigurationMask)
	}
	macro.Set("PAD_CFG")
	if macro.IsOwnershipDriver() {
		// PAD_CFG_GPO_GPIO_DRIVER(pad, val, rst, pull)
		macro.Add("_GPO_GPIO_DRIVER").Add("(").Id().Val().Rstsrc().Pull().Add("),")
		return
	}
	if term != 0 {
		// e.g. PAD_CFG_TERM_GPO(GPP_B23, 1, DN_20K, DEEP),
		macro.Add("_TERM")
	}
	macro.Add("_GPO").Add("(").Id().Val()
	if term != 0 {
		macro.Pull()
	}
	macro.Rstsrc().Add("),")
}

// Adds PAD_CFG_NF macro with arguments
func (PlatformSpecific) NativeFunctionMacroAdd(macro *common.Macro) {
	// e.g. PAD_CFG_NF(GPP_D23, NONE, DEEP, NF1)
	macro.Set("PAD_CFG_NF").Add("(").Id().Pull().Rstsrc().Padfn().Add("),")
}

// Adds PAD_NC macro
func (PlatformSpecific) NoConnMacroAdd(macro *common.Macro) {
	// #define PAD_NC(pad, pull)
	// _PAD_CFG_STRUCT(pad,
	//     PAD_FUNC(GPIO) | PAD_RESET(DEEP) | PAD_TRIG(OFF) | PAD_BUF(TX_RX_DISABLE),
	//     PAD_PULL(pull) | PAD_IOSSTATE(TxDRxE)),
	dw0 := macro.Register(PAD_CFG_DW0)

	// Some fields of the configuration registers are hidden inside the macros,
	// we should check them to update the corresponding bits in the control mask.
	if dw0.GetRXLevelEdgeConfiguration() != common.TRIG_OFF {
		dw0.CntrMaskFieldsClear(common.RxLevelEdgeConfigurationMask)
	}
	if dw0.GetResetConfig() != 1 { // 1 = RST_DEEP
		dw0.CntrMaskFieldsClear(common.PadRstCfgMask)
	}

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
	cb.SetFieldsIface(&macro)
	macro.PadIdSet(id).SetPadOwnership(ownership)
	macro.Register(PAD_CFG_DW0).ValueSet(dw0).ReadOnlyFieldsSet(PAD_CFG_DW0_RO_FIELDS)
	macro.Register(PAD_CFG_DW1).ValueSet(dw1).ReadOnlyFieldsSet(PAD_CFG_DW1_RO_FIELDS)
	return macro.Generate()
}
