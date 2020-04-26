package common

// Bit field constants for PAD_CFG_DW0 register
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

	rxtxEnableConfigShift uint8  = 21
	rxtxEnableConfigMask  uint32 = 0x3 << rxtxEnableConfigShift

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

// Bit field constants for PAD_CFG_DW1 register
const (
	ioStandbyStateShift uint8  = 14
	ioStandbyStateMask  uint32 = 0xF << ioStandbyStateShift

	termShift uint8  = 10
	termMask  uint32 = 0xF << termShift

	ioStandbyTerminationShift uint8  = 8
	ioStandbyTerminationMask  uint32 = 0x3 << ioStandbyTerminationShift

	interruptSelectShift uint32 = 0xFF
)

// config DW registers
const (
	PAD_CFG_DW0 = 0
	PAD_CFG_DW1 = 1
	MAX_DW_NUM  = 2
)

// Register - configuration data structure based on DW0/1 dw value
// value    : register value
// mask     : bit fileds mask
// roFileds : read only fields mask
type Register struct {
	value    uint32
	mask     uint32
	roFileds uint32
}

func (reg *Register) ValueSet(value uint32) *Register {
	reg.value = value
	return reg
}

func (reg *Register) ValueGet() uint32 {
	return reg.value
}

func (reg *Register) ReadOnlyFieldsSet(fileldMask uint32) *Register {
	reg.roFileds = fileldMask
	return reg
}

func (reg *Register) ReadOnlyFieldsGet() uint32 {
	return reg.roFileds
}

// Check the mask of the new macro
// Returns true if the macro is generated correctly
func (reg *Register) MaskCheck() bool {
	mask := ^(reg.mask | reg.roFileds)
	return reg.value&mask == 0
}

// getResetConfig - get Reset Configuration from PADRSTCFG field in PAD_CFG_DW0_GPx register
func (reg *Register) getFieldVal(mask uint32, shift uint8) uint8 {
	reg.mask |= mask
	return uint8((reg.value & mask) >> shift)
}

// Fix Pad Reset Config field in mask for DW0 register
// Returns *Register
func (reg *Register) MaskResetFix() *Register {
	reg.mask |= padRstCfgMask
	return reg
}

// Fix RX Level/Edge Configuration field in mask for DW0 register
// Returns *Register
func (reg *Register) MaskTrigFix() *Register {
	reg.mask |= rxLevelEdgeConfigurationMask
	return reg
}

// getResetConfig - returns type reset source for corresponding pad
// PADRSTCFG field in PAD_CFG_DW0 register
func (reg *Register) GetResetConfig() uint8 {
	return reg.getFieldVal(padRstCfgMask, padRstCfgShift)
}

// getRXPadStateSelect - returns RX Pad State (RXINV)
// 0 = Raw RX pad state directly from RX buffer
// 1 = Internal RX pad state
func (reg *Register) GetRXPadStateSelect() uint8 {
	return reg.getFieldVal(rxPadStateSelectMask, rxPadStateSelectShift)
}

// getRXRawOverrideStatus - returns 1 if the selected pad state is being
// overridden to '1' (RXRAW1 field)
func (reg *Register) GetRXRawOverrideStatus() uint8 {
	return reg.getFieldVal(rxRawOverrideTo1Mask, rxRawOverrideTo1Shift)
}

// getRXLevelEdgeConfiguration - returns RX Level/Edge Configuration (RXEVCFG)
// 0h = Level, 1h = Edge, 2h = Drive '0', 3h = Reserved (implement as setting 0h)
func (reg *Register) GetRXLevelEdgeConfiguration() uint8 {
	return reg.getFieldVal(rxLevelEdgeConfigurationMask, rxLevelEdgeConfigurationShift)
}

// getRXLevelConfiguration - returns RX Invert state (RXINV)
// 1 - Inversion, 0 - No inversion
func (reg *Register) GetRXLevelConfiguration() bool {
	return reg.getFieldVal(rxInvertMask, rxInvertShift) != 0
}

// getRxTxEnableConfig - returns RX/TX Enable Config (RXTXENCFG)
// 0 = Function defined in Pad Mode controls TX and RX Enables
// 1 = Function controls TX Enable and RX Disabled with RX drive 0 internally
// 2 = Function controls TX Enable and RX Disabled with RX drive 1 internally
// 3 = Function controls TX Enabled and RX is always enabled
func (reg *Register) GetRxTxEnableConfig() uint8 {
	return reg.getFieldVal(rxtxEnableConfigMask, rxtxEnableConfigShift)
}

// getGPIOInputRouteIOxAPIC - returns 1 if the pad can be routed to cause
// peripheral IRQ when configured in GPIO input mode.
func (reg *Register) GetGPIOInputRouteIOxAPIC() bool {
	return reg.getFieldVal(gpioInputRouteIOxApicMask, gpioInputRouteIOxApicShift) != 0
}

// getGPIOInputRouteSCI - returns 1 if the pad can be routed to cause SCI when
// configured in GPIO input mode.
func (reg *Register) GetGPIOInputRouteSCI() bool {
	return reg.getFieldVal(gpioInputRouteSCIMask, gpioInputRouteSCIShift) != 0
}

// getGPIOInputRouteSMI - returns 1 if the pad can be routed to cause SMI when
// configured in GPIO input mode
func (reg *Register) GetGPIOInputRouteSMI() bool {
	return reg.getFieldVal(gpioInputRouteSMIMask, gpioInputRouteSMIShift) != 0
}

// getGPIOInputRouteNMI - returns 1 if the pad can be routed to cause NMI when
// configured in GPIO input mode
func (reg *Register) GetGPIOInputRouteNMI() bool {
	return reg.getFieldVal(gpioInputRouteNMIMask, gpioInputRouteNMIShift) != 0
}

// getPadMode - reutrns pad mode or one of the native functions
// 0h = GPIO control the Pad.
// 1h = native function 1, if applicable, controls the Pad
// 2h = native function 2, if applicable, controls the Pad
// 3h = native function 3, if applicable, controls the Pad
// 4h = enable GPIO blink/PWM capability if applicable
func (reg *Register) GetPadMode() uint8 {
	return reg.getFieldVal(padModeMask, padModeShift)
}

// getGPIORxTxDisableStatus - returns GPIO RX/TX buffer state (GPIORXDIS | GPIOTXDIS)
// 0 - both are enabled, 1 - TX Disable, 2 - RX Disable, 3 - both are disabled
func (reg *Register) GetGPIORxTxDisableStatus() uint8 {
	return reg.getFieldVal(gpioRxTxDisableMask, gpioRxTxDisableShift)
}

// getGPIORXState - returns GPIO RX State (GPIORXSTATE)
func (reg *Register) GetGPIORXState() uint8 {
	return reg.getFieldVal(gpioRxStateMask, gpioRxStateShift)
}

// getGPIOTXState - returns GPIO TX State (GPIOTXSTATE)
func (reg *Register) GetGPIOTXState() int {
	return int(reg.getFieldVal(gpioTxStateMask, 0))
}

// getIOStandbyState - return IO Standby State (IOSSTATE)
// 0 = Tx enabled driving last value driven, Rx enabled
// 1 = Tx enabled driving 0, Rx disabled and Rx driving 0 back to its controller internally
// 2 = Tx enabled driving 0, Rx disabled and Rx driving 1 back to its controller internally
// 3 = Tx enabled driving 1, Rx disabled and Rx driving 0 back to its controller internally
// 4 = Tx enabled driving 1, Rx disabled and Rx driving 1 back to its controller internally
// 5 = Tx enabled driving 0, Rx enabled
// 6 = Tx enabled driving 1, Rx enabled
// 7 = Hi-Z, Rx driving 0 back to its controller internally
// 8 = Hi-Z, Rx driving 1 back to its controller internally
// 9 = Tx disabled, Rx enabled
// 15 = IO-Standby is ignored for this pin (same as functional mode)
// Others reserved
func (reg *Register) GetIOStandbyState() uint8 {
	return reg.getFieldVal(ioStandbyStateMask, ioStandbyStateShift)
}

// getIOStandbyTermination - return IO Standby Termination (IOSTERM)
// 0 = Same as functional mode (no change)
// 1 = Disable Pull-up and Pull-down (no on-die termination)
// 2 = Enable Pull-down
// 3 = Enable Pull-up
func (reg *Register) GetIOStandbyTermination() uint8 {
	return reg.getFieldVal(ioStandbyTerminationMask, ioStandbyTerminationShift)
}

// getTermination - returns the pad termination state defines the different weak
// pull-up and pull-down settings that are supported by the buffer
// 0000 = none; 0010 = 5k PD; 0100 = 20k PD; 1010 = 5k PU; 1100 = 20k PU;
// 1111 = Native controller selected
func (reg *Register) GetTermination() uint8 {
	return reg.getFieldVal(termMask, termShift)
}

// getInterruptSelect - returns Interrupt Line number from the GPIO controller
func (reg *Register) GetInterruptSelect() uint8 {
	return reg.getFieldVal(interruptSelectShift, 0)
}
