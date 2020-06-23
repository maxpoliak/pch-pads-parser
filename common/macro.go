package common

import "strconv"
import "../config"

const (
	PAD_OWN_ACPI   = 0
	PAD_OWN_DRIVER = 1
)

const (
	TxLASTRxE     = 0x0
	Tx0RxDCRx0    = 0x1
	Tx0RxDCRx1    = 0x2
	Tx1RxDCRx0    = 0x3
	Tx1RxDCRx1    = 0x4
	Tx0RxE        = 0x5
	Tx1RxE        = 0x6
	HIZCRx0       = 0x7
	HIZCRx1       = 0x8
	TxDRxE        = 0x9
	StandbyIgnore = 0xf
)

const (
	IOSTERM_SAME	= 0x0
	IOSTERM_DISPUPD	= 0x1
	IOSTERM_ENPD	= 0x2
	IOSTERM_ENPU    = 0x3
)

const (
	RST_PWROK  = 0
	RST_DEEP   = 1
	RST_PLTRST = 2
	RST_RSMRST = 3
)

const (
	TRIG_LEVEL       = 0
	TRIG_EDGE_SINGLE = 1
	TRIG_OFF         = 2
	TRIG_EDGE_BOTH   = 3
)

// PlatformSpecific - platform-specific interface
type PlatformSpecific interface {
	Rstsrc(macro *Macro)
	Pull(macro *Macro)
	GpiMacroAdd(macro *Macro)
	GpoMacroAdd(macro *Macro)
	NativeFunctionMacroAdd(macro *Macro)
	NoConnMacroAdd(macro *Macro)
}

// Macro - contains macro information and methods
// Platform : platform-specific interface
// padID    : pad ID string
// str      : macro string entirely
// Reg      : structure of configuration register values and their masks
type Macro struct {
	Platform  PlatformSpecific
	Reg       [MAX_DW_NUM]Register
	padID     string
	str       string
	ownership uint8
}

func (macro *Macro) PadIdGet() string {
	return macro.padID
}

func (macro *Macro) PadIdSet(padid string) *Macro {
	macro.padID = padid
	return macro
}

func (macro *Macro) SetPadOwnership(own uint8) *Macro {
	macro.ownership = own
	return macro
}

func (macro *Macro) IsOwnershipDriver() bool {
	return macro.ownership == PAD_OWN_DRIVER
}

// returns <Register> data configuration structure
// number : register number
func (macro *Macro) Register(number uint8) *Register {
	return &macro.Reg[number]
}

// add a string to macro
func (macro *Macro) Add(str string) *Macro {
	macro.str += str
	return macro
}

// set a string in a macro instead of its previous contents
func (macro *Macro) Set(str string) *Macro {
	macro.str = str
	return macro
}

// get macro string
func (macro *Macro) Get() string {
	return macro.str
}

// Adds PAD Id to the macro as a new argument
// return: Macro
func (macro *Macro) Id() *Macro {
	return macro.Add(macro.padID)
}

// Add Separator to macro if needed
func (macro *Macro) Separator() *Macro {
	str := macro.Get()
	c := str[len(str)-1]
	if c != '(' && c != '_' {
		macro.Add(", ")
	}
	return macro
}

// Adds the PADRSTCFG parameter from DW0 to the macro as a new argument
// return: Macro
func (macro *Macro) Rstsrc() *Macro {
	macro.Platform.Rstsrc(macro)
	return macro
}

// Adds The Pad Termination (TERM) parameter from DW1 to the macro as a new argument
// return: Macro
func (macro *Macro) Pull() *Macro {
	macro.Platform.Pull(macro)
	return macro
}

// Adds Pad GPO value to macro string as a new argument
// return: Macro
func (macro *Macro) Val() *Macro {
	dw0 := macro.Register(PAD_CFG_DW0)
	return macro.Separator().Add(strconv.Itoa(dw0.GetGPIOTXState()))
}

// Adds Pad GPO value to macro string as a new argument
// return: Macro
func (macro *Macro) Trig() *Macro {
	dw0 := macro.Register(PAD_CFG_DW0)
	var trig = map[uint8]string{
		0x0: "LEVEL",
		0x1: "EDGE_SINGLE",
		0x2: "OFF",
		0x3: "EDGE_BOTH",
	}
	return macro.Separator().Add(trig[dw0.GetRXLevelEdgeConfiguration()])
}

// Adds Pad Polarity Inversion Stage (RXINV) to macro string as a new argument
// return: Macro
func (macro *Macro) Invert() *Macro {
	macro.Separator()
	if macro.Register(PAD_CFG_DW0).GetRXLevelConfiguration() !=0 {
		return macro.Add("INVERT")
	}
	return macro.Add("NONE")
}

// Adds input/output buffer state
// return: Macro
func (macro *Macro) Bufdis() *Macro {
	var buffDisStat = map[uint8]string{
		0x0: "NO_DISABLE",    // both buffers are enabled
		0x1: "TX_DISABLE",    // output buffer is disabled
		0x2: "RX_DISABLE",    // input buffer is disabled
		0x3: "TX_RX_DISABLE", // both buffers are disabled
	}
	state := macro.Register(PAD_CFG_DW0).GetGPIORxTxDisableStatus()
	return macro.Separator().Add(buffDisStat[state])
}

// Adds macro to set the host software ownership
// return: Macro
func (macro *Macro) Own() *Macro {
	if macro.IsOwnershipDriver() {
		return macro.Separator().Add("DRIVER")
	}
	return macro.Separator().Add("ACPI")
}

//Adds pad native function (PMODE) as a new argument
//return: Macro
func (macro *Macro) Padfn() *Macro {
	dw0 := macro.Register(PAD_CFG_DW0)
	nfnum := int(dw0.GetPadMode())
	if nfnum != 0 {
		return macro.Separator().Add("NF" + strconv.Itoa(nfnum))
	}
	// GPIO used only for PAD_FUNC(x) macro
	return macro.Add("GPIO")
}

// Add a line to the macro that defines IO Standby State
// return: macro
func (macro *Macro) IOSstate() *Macro {
	var stateMacro = map[uint8]string{
		TxLASTRxE:     "TxLASTRxE",
		Tx0RxDCRx0:    "Tx0RxDCRx0",
		Tx0RxDCRx1:    "Tx0RxDCRx1",
		Tx1RxDCRx0:    "Tx1RxDCRx0",
		Tx1RxDCRx1:    "Tx1RxDCRx1",
		Tx0RxE:        "Tx0RxE",
		Tx1RxE:        "Tx1RxE",
		HIZCRx0:       "HIZCRx0",
		HIZCRx1:       "HIZCRx1",
		TxDRxE:        "TxDRxE",
		StandbyIgnore: "IGNORE",
	}
	dw1 := macro.Register(PAD_CFG_DW1)
	str, valid := stateMacro[dw1.GetIOStandbyState()]
	if !valid {
		// ignore setting for incorrect value
		str = "IGNORE"
	}
	return macro.Separator().Add(str)
}

// Add a line to the macro that defines IO Standby Termination
// return: macro
func (macro *Macro) IOTerm() *Macro {
	var ioTermMacro = map[uint8]string{
		IOSTERM_SAME:    "SAME",
		IOSTERM_DISPUPD: "DISPUPD",
		IOSTERM_ENPD:    "ENPD",
		IOSTERM_ENPU:    "ENPU",
	}
	dw1 := macro.Register(PAD_CFG_DW1)
	return macro.Separator().Add(ioTermMacro[dw1.GetIOStandbyTermination()])
}

// Check created macro
func check(macro *Macro) {
	if !macro.Register(PAD_CFG_DW0).MaskCheck() {
		macro.Advanced()
	}
}

// Add Input Route
func (macro *Macro) irqInputRoute() *Macro {
	dw0 := macro.Register(PAD_CFG_DW0)
	var options []string

	for key, isRoute := range map[string]func()uint8{
		"IOAPIC": dw0.GetGPIOInputRouteIOxAPIC,
		"SCI":    dw0.GetGPIOInputRouteSCI,
		"SMI":    dw0.GetGPIOInputRouteSMI,
		"NMI":    dw0.GetGPIOInputRouteNMI,
	} {
		if isRoute() != 0 {
			options = append(options, key)
		}
	}
	switch len(options) {
	case 0:
		break
	case 1:
		// PAD_IRQ_CFG(route, trig, inv)
		macro.Add("PAD_IRQ_CFG(" + options[0]).Trig().Invert()
		return macro
	case 2:
		// PAD_IRQ_CFG_DUAL_ROUTE(route1, route2, trig, inv)
		macro.Add("PAD_IRQ_CFG_DUAL_ROUTE(" + options[0]+ ", " + options[1])
		macro.Trig().Invert()
		return macro
	default:
		// PAD_IRQ_CFG(route1) | PAD_IRQ_CFG(route2) | PAD_IRQ_CFG(route3) ...
		for _, option := range options {
			macro.Add("PAD_IRQ_CFG(" + option)
		}
	}
	// PAD_TRIG(trig) | PAD_RX_POL(inv) | ...
	macro.Add("PAD_TRIG(").Trig()
	if dw0.GetRXLevelConfiguration() != 0 {
		macro.Add(") | PAD_RX_POL(").Invert()
	}
	return macro
}

// Generate Advanced Macro
func (macro *Macro) Advanced() {
	dw0 := macro.Register(PAD_CFG_DW0)
	dw1 := macro.Register(PAD_CFG_DW1)

	// Add string of a reference macro as a comment
	reference := macro.Get()
	macro.Set("/* ").Add(reference).Add(" */")

	macro.Add("\n\t_PAD_CFG_STRUCT(").Id().Add(",\n\t\tPAD_FUNC(").Padfn()
	macro.Add(") | PAD_RESET(").Rstsrc().Add(") | ").irqInputRoute().Add(")")
	if dw0.GetGPIORxTxDisableStatus() != 0 {
		macro.Add(" | PAD_BUF(").Bufdis().Add(")")
	}
	if dw0.GetRXPadStateSelect() != 0 {
		macro.Add(" | (1 << 29)")
	}
	if dw0.GetRXRawOverrideStatus() != 0 {
		macro.Add(" | (1 << 28)")
	}

	if dw0.GetGPIOTXState() != 0 {
		macro.Add(" | 1")
	}
	if dw1.ValueGet() == 0 {
		macro.Add(", 0")
	} else {
		macro.Add(",\n\t\t").Add("PAD_PULL(").Pull().Add(")")
		if dw1.GetIOStandbyState() != 0 {
			macro.Add(" | PAD_IOSSTATE(").IOSstate().Add(")")
		}
		if dw1.GetIOStandbyTermination() != 0 {
			macro.Add(" | PAD_IOSTERM(").IOTerm().Add(")")
		}
		if macro.IsOwnershipDriver() {
			macro.Add(" | PAD_CFG_OWN_GPIO(").Own().Add(")")
		}
	}
	macro.Add("),")
}

// Generate macro for bi-directional GPIO port
func (macro *Macro) Bidirection() {
	dw1 := macro.Register(PAD_CFG_DW1)
	ios := dw1.GetIOStandbyState() != 0 || dw1.GetIOStandbyTermination() != 0
	macro.Set("PAD_CFG_GPIO_BIDIRECT")
	if ios {
		macro.Add("_IOS")
	}
	// PAD_CFG_GPIO_BIDIRECT(pad, val, pull, rst, trig, own)
	macro.Add("(").Id().Val().Pull().Rstsrc().Trig()
	if ios {
		// PAD_CFG_GPIO_BIDIRECT_IOS(pad, val, pull, rst, trig, iosstate, iosterm, own)
		macro.IOSstate().IOTerm()
	}
	macro.Own().Add("),")
}

const (
	rxDisable uint8 = 0x2
	txDisable uint8 = 0x1
)

// Gets base string of current macro
// return: string of macro
func (macro *Macro) Generate() string {
	macro.Set("PAD_CFG")
	dw0 := macro.Register(PAD_CFG_DW0)
	if dw0.GetPadMode() == 0 {
		// GPIO
		switch dw0.GetGPIORxTxDisableStatus() {
		case txDisable:
			macro.Platform.GpiMacroAdd(macro) // GPI

		case rxDisable:
			macro.Platform.GpoMacroAdd(macro) // GPO

		case rxDisable | txDisable:
			macro.Platform.NoConnMacroAdd(macro) // NC

		default:
			macro.Bidirection()
		}
	} else {
		macro.Platform.NativeFunctionMacroAdd(macro)
	}

	if config.IsAdvancedFormatUsed() {
		// Clear control mask to generate advanced macro only
		macro.Register(PAD_CFG_DW0).ClearCntrMask()
	}

	if config.IsNonCheckingFlagUsed() {
		// Generate macros without checking
		return macro.Get()
	}

	check(macro)
	return macro.Get()
}
