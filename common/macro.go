package common

import (
	"strconv"
)

const (
	PAD_OWN_ACPI   = 0
	PAD_OWN_DRIVER = 1
)

// Platform - platform-specific interface
type PlatformSpecific interface {
	Rstsrc(macro *Macro)
	Pull(macro *Macro)
	GpiMacroAdd(macro *Macro)
	GpoMacroAdd(macro *Macro)
	NativeFunctionMacroAdd(macro *Macro)
	AdvancedMacroGenerate(macro *Macro)
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
	var dw0 = macro.Register(PAD_CFG_DW0)
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
	if macro.Register(PAD_CFG_DW0).GetRXLevelConfiguration() {
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
		0x0: "SAME",
		0x1: "DISPUPD",
		0x2: "ENPD",
		0x3: "ENPU",
	}
	dw1 := macro.Register(PAD_CFG_DW1)
	return macro.Separator().Add(ioTermMacro[dw1.GetIOStandbyTermination()])
}

// Check created macro
func check(macro *Macro) {
	if !macro.Register(PAD_CFG_DW0).MaskCheck() {
		macro.Platform.AdvancedMacroGenerate(macro)
	}
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
			// NC
			macro.Set("PAD_NC").Add("(").Id().Pull().Add("),")
			return macro.Get()

		default:
			// In case the rule isn't found, a common macro is used
			// to create the pad configuration
			macro.Platform.AdvancedMacroGenerate(macro)
			return macro.Get()
		}
	} else {
		macro.Platform.NativeFunctionMacroAdd(macro)
		return macro.Get()
	}
	check(macro)
	return macro.Get()
}
