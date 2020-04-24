package common

import (
	"fmt"
	"strconv"
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
	Platform PlatformSpecific
	Reg      [MAX_DW_NUM]Register
	padID    string
	str      string
	Driver   bool
}

func (macro *Macro) PadIdGet() string {
	return macro.padID
}

func (macro *Macro) PadIdSet(padid string) {
	macro.padID = padid
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
	if macro.Driver {
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

// Check created macro
func check(macro *Macro) {
	if !macro.Register(PAD_CFG_DW0).MaskCheck() {
		macro.Platform.AdvancedMacroGenerate(macro)
		// Debug message about this
		fmt.Printf("\nNo configuration for pad: %s\n", macro.padID)
		fmt.Printf("Use %s\n", macro.Get())
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
			macro.Set("PAD_NC").Add("(").Id().Pull()
			// Fix mask for RX Level/Edge Configuration (RXEVCFG)
			// and Pad Reset Config (PADRSTCFG)
			// There is no need to check these fields if the pad
			// is in the NC state
			dw0.MaskResetFix().MaskTrigFix()
			macro.Add("),")

		default:
			// In case the rule isn't found, a common macro is used
			// to create the pad configuration
			macro.Platform.AdvancedMacroGenerate(macro)
			return macro.Get()
		}
	} else {
		macro.Platform.NativeFunctionMacroAdd(macro)
	}
	check(macro)
	return macro.Get()
}