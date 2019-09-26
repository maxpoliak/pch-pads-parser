package sunrise

// Bit field constants for DW0 register
const (
	padRstCfgShift uint8  = 30
	padRstCfgMask  uint32 = 0x3 << padRstCfgShift

	rxPadStateSelectShift uint8  = 29
	rxPadStateSelectMask  uint32 = 0x1 << rxPadStateSelectShift

	rxRawOverrideTo1Shift uint8  = 28
	rxRawOverrideTo1Mask  uint32 = 0x1 << rxRawOverrideTo1Shift

	rxLevelEdgeConfigurationShift uint8  = 25
	rxLevelEdgeConfigurationMask  uint32 = 0x3 << rxLevelEdgeConfigurationShift

	rxInvertShift uint8  = 23
	rxInvertMask  uint32 = 0x1 << rxInvertShift

	gpioInputRouteIOxApicShift uint8  = 20
	gpioInputRouteIOxApicMask  uint32 = 0x1 << gpioInputRouteIOxApicShift

	gpioInputRouteSCIShift uint8  = 19
	gpioInputRouteSCIMask  uint32 = 0x1 << gpioInputRouteSCIShift

	gpioInputRouteSMIShift uint8  = 18
	gpioInputRouteSMIMask  uint32 = 0x1 << gpioInputRouteSMIShift

	gpioInputRouteNMIShift uint8  = 17
	gpioInputRouteNMIMask  uint32 = 0x1 << gpioInputRouteNMIShift

	padModeShift uint8  = 10
	padModeMask  uint32 = 0x7 << padModeShift

	gpioRxTxDisableShift uint8  = 8
	gpioRxTxDisableMask  uint32 = 0x3 << gpioRxTxDisableShift

	gpioRxStateShift uint8  = 1
	gpioRxStateMask  uint32 = 0x1 << gpioRxStateShift

	gpioTxStateMask uint32 = 0x1
)

// Bit field constants for DW1 register
const (
	termShift            uint8  = 10
	termMask             uint32 = 0xF << termShift
	interruptSelectShift uint32 = 0xFF
)

// Maximum numbers of the config DW register for Sunrise chipset
const (
	MaxDWNum = 2
)

// cfgDW - information about pad
// reg   : register value
// mask  : bit fileds mask
type cfgDW struct {
	reg  [MaxDWNum]uint32
	mask [MaxDWNum]uint32
}

// getResetConfig - get Reset Configuration from PADRSTCFG field in PAD_CFG_DW0_GPx register
func (data *cfgDW) getFieldVal(regNum uint8, mask uint32, shift uint8) uint8 {
	if regNum < MaxDWNum {
		data.mask[regNum] |= mask
		return uint8((data.reg[regNum] & mask) >> shift)
	}
	return 0
}

// Check the mask of the new macro
// regNum : DW register number
// Returns true if the macro is generated correctly
func (data *cfgDW) maskCheck(regNum uint8) bool {
	if regNum >= MaxDWNum {
		return false
	}

	// Take into account the bits that are read-only
	readonly := [MaxDWNum]uint32{
		(0x1 << 27) | (0x1 << 24) | (0x3 << 21) | (0xf << 16) | 0xfe,
		0xfffffc3f,
	}
	mask := ^(data.mask[regNum] | readonly[regNum])
	return data.reg[regNum]&mask == 0
}

// Fix Pad Reset Config field in mask for DW0 register
// Returns *cfgDW
func (data *cfgDW) maskResetFix() *cfgDW {
	data.mask[0] |= padRstCfgMask
	return data
}

// Fix RX Level/Edge Configuration field in mask for DW0 register
// Returns *cfgDW
func (data *cfgDW) maskTrigFix() *cfgDW {
	data.mask[0] |= rxLevelEdgeConfigurationMask
	return data
}

// getResetConfig - returns type reset source for corresponding pad
// PADRSTCFG field in PAD_CFG_DW0 register
// 00 = RSMRST#, 01 = DEEP#, 10 = PLTRST#, 11 = RSMRST# for GPDx pads
// and Reserved for others
func (data *cfgDW) getResetConfig() uint8 {
	return data.getFieldVal(0, padRstCfgMask, padRstCfgShift)
}

// getRXPadStateSelect - returns RX Pad State (RXINV)
// 0 = Raw RX pad state directly from RX buffer
// 1 = Internal RX pad state
func (data *cfgDW) getRXPadStateSelect() uint8 {
	return data.getFieldVal(0, rxPadStateSelectMask, rxPadStateSelectShift)
}

// getRXRawOverrideStatus - returns 1 if the selected pad state is being
// overridden to '1' (RXRAW1 field)
func (data *cfgDW) getRXRawOverrideStatus() uint8 {
	return data.getFieldVal(0, rxRawOverrideTo1Mask, rxRawOverrideTo1Shift)
}

// getRXLevelEdgeConfiguration - returns RX Level/Edge Configuration (RXEVCFG)
// 0h = Level, 1h = Edge, 2h = Drive '0', 3h = Reserved (implement as setting 0h)
func (data *cfgDW) getRXLevelEdgeConfiguration() uint8 {
	return data.getFieldVal(0, rxLevelEdgeConfigurationMask, rxLevelEdgeConfigurationShift)
}

// getRXLevelConfiguration - returns RX Invert state (RXINV)
// 1 - Inversion, 0 - No inversion
func (data *cfgDW) getRXLevelConfiguration() bool {
	return data.getFieldVal(0, rxInvertMask, rxInvertShift) != 0
}

// getGPIOInputRouteIOxAPIC - returns 1 if the pad can be routed to cause
// peripheral IRQ when configured in GPIO input mode.
func (data *cfgDW) getGPIOInputRouteIOxAPIC() bool {
	return data.getFieldVal(0, gpioInputRouteIOxApicMask, gpioInputRouteIOxApicShift) != 0
}

// getGPIOInputRouteSCI - returns 1 if the pad can be routed to cause SCI when
// configured in GPIO input mode.
func (data *cfgDW) getGPIOInputRouteSCI() bool {
	return data.getFieldVal(0, gpioInputRouteSCIMask, gpioInputRouteSCIShift) != 0
}

// getGPIOInputRouteSMI - returns 1 if the pad can be routed to cause SMI when
// configured in GPIO input mode
func (data *cfgDW) getGPIOInputRouteSMI() bool {
	return data.getFieldVal(0, gpioInputRouteSMIMask, gpioInputRouteSMIShift) != 0
}

// getGPIOInputRouteNMI - returns 1 if the pad can be routed to cause NMI when
// configured in GPIO input mode
func (data *cfgDW) getGPIOInputRouteNMI() bool {
	return data.getFieldVal(0, gpioInputRouteNMIMask, gpioInputRouteNMIShift) != 0
}

// getPadMode - reutrns pad mode or one of the native functions
// 0h = GPIO control the Pad.
// 1h = native function 1, if applicable, controls the Pad
// 2h = native function 2, if applicable, controls the Pad
// 3h = native function 3, if applicable, controls the Pad
// 4h = enable GPIO blink/PWM capability if applicable
func (data *cfgDW) getPadMode() uint8 {
	return data.getFieldVal(0, padModeMask, padModeShift)
}

// getGPIORxTxDisableStatus - returns GPIO RX/TX buffer state (GPIORXDIS | GPIOTXDIS)
// 0 - both are enabled, 1 - TX Disable, 2 - RX Disable, 3 - both are disabled
func (data *cfgDW) getGPIORxTxDisableStatus() uint8 {
	return data.getFieldVal(0, gpioRxTxDisableMask, gpioRxTxDisableShift)
}

// getGPIORXState - returns GPIO RX State (GPIORXSTATE)
func (data *cfgDW) getGPIORXState() uint8 {
	return data.getFieldVal(0, gpioRxStateMask, gpioRxStateShift)
}

// getGPIOTXState - returns GPIO TX State (GPIOTXSTATE)
func (data *cfgDW) getGPIOTXState() int {
	return int(data.getFieldVal(0, gpioTxStateMask, 0))
}

// getTermination - returns the pad termination state defines the different weak
// pull-up and pull-down settings that are supported by the buffer
// 0000 = none; 0010 = 5k PD; 0100 = 20k PD; 1010 = 5k PU; 1100 = 20k PU;
// 1111 = Native controller selected
func (data *cfgDW) getTermination() uint8 {
	return data.getFieldVal(1, termMask, termShift)
}

// getInterruptSelect - returns Interrupt Line number from the GPIO controller
func (data *cfgDW) getInterruptSelect() uint8 {
	return data.getFieldVal(1, interruptSelectShift, 0)
}
