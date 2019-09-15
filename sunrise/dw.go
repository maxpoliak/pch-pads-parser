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

// getFieldVal - get bit fields value in pad config DW register
// regNum : DW register number
// mask   : bit field mask
// shift  : shift of the bit field mask in 32-bit DW register
// Returns true if the macro is generated correctly
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

/*
Pad Reset Config (PADRSTCFG): This register controls which reset is used
to reset GPIO pad register fields in PAD_CFG_DW0 and PAD_CFG_DW1 registers.
This register can be used for Sx isolation of the associated signal if
needed.
00 = RSMRST#
01 = Host deep reset. This reset occurs when any of the following occur:
Host reset (with or without power cycle) is initiated, Global reset is
initiated, or Deep Sx entry. This reset does NOT occur as part of S3/S4/S5
entry.
10 = PLTRST#
11 = Reserved
*/
func (data *cfgDW) getResetConfig() uint8 {
	return data.getFieldVal(0, padRstCfgMask, padRstCfgShift)
}

/*
RX Pad State Select (RXPADSTSEL): Determines from which node the RX pad
state for native function should be taken from. This field only affects the
pad state value being fanned out to native function(s) and isn`t meaningful
if the pad is in GPIO mode (i.e. Pad Mode = 0)
0 = Raw RX pad state directly from RX buffer
1 = Internal RX pad state (subject to RXINV and PreGfRXSel settings)
*/
func (data *cfgDW) getRXPadStateSelect() uint8 {
	return data.getFieldVal(0, rxPadStateSelectMask, rxPadStateSelectShift)
}

/*
RX Raw Override to '1' (RXRAW1): This bit determines if the selected pad
state is being overridden to '1'. This field is only applicable when the RX
buffer is configured as an input in either GPIO Mode or native function
mode. The override takes place at the internal pad state directly from
buffer and before the RXINV
0 = No Override
1 = RX drive 1 internally
*/
func (data *cfgDW) getRXRawOverrideStatus() uint8 {
	return data.getFieldVal(0, rxRawOverrideTo1Mask, rxRawOverrideTo1Shift)
}

/*
RX Level/Edge Configuration (RXEVCFG): Determines if the internal RX pad
state (synchronized, filtered vs non-filtered version as determined by
PreGfRXSel, and is further subject to RXInv) should be passed on to the
next logic stage as is, as a pulse, or level signal. This field does not
affect the received pad state (to GPIORXState or native functions) but how
the interrupt or wake triggering events should be delivered to the GPIO
Community Controller
0h = Level
1h = Edge
2h = Drive '0'
3h = Reserved (implement as setting 0h)
*/
func (data *cfgDW) getRXLevelEdgeConfiguration() uint8 {
	return data.getFieldVal(0, rxLevelEdgeConfigurationMask, rxLevelEdgeConfigurationShift)
}

/*
RX Invert (RсonfigDataXINV): This bit determines if the selected pad state should
go through the polarity inversion stage. This field is only applicable when
the RX buffer is configured as an input in either GPIO Mode or native
function mode. The polarity inversion takes place at the mux node of raw vs
filtered or non-filtered RX pad state, as determined by PreGfRXsel and
RXPadStSel This bit does not affect GPIORXState. During host ownership GPIO
Mode, when this bit is set to '1', then the RX pad state is inverted as it
is sent to the GPIO-to-IOxAPIC, GPE/SCI, SMI, NMI logic or GPI_IS[n] that
is using it. This is used to allow active-low and active-high inputs to
cause IRQ, SMI#, SCI or NMI
0 = No inversion
1 = Inversion
*/
func (data *cfgDW) getRXLevelConfiguration() bool {
	return data.getFieldVal(0, rxInvertMask, rxInvertShift) != 0
}

/*
GPIO Input Route IOxAPIC (GPIROUTIOXAPIC): Determines if the pad can be
routed to cause peripheral IRQ when configured in GPIO input mode. If the
pad is not configured in GPIO input mode, this field has no effect.
0 = Routing does not cause peripheral IRQ
1 = Routing can cause peripheral IRQ
Note: This bit does not affect any interrupt status bit within GPIO, but
is used as the last qualifier for the peripheral IRQ indication to the
intended recipient(s).
*/
func (data *cfgDW) getGPIOInputRouteIOxAPIC() bool {
	return data.getFieldVal(0, gpioInputRouteIOxApicMask, gpioInputRouteIOxApicShift) != 0
}

/*
GPIO Input Route SCI (GPIROUTSCI): Determines if the pad can be routed
to cause SCI when configured in GPIO input mode. If the pad is not
configured in GPIO input mode, this field has no effect.
0 = Routing does not cause SCI.
1 = Routing can cause SCI
Note: This bit does not affect any interrupt status bit within GPIO, but is
used as the last qualifier for the GPE indication to the intended
recipient(s).
*/
func (data *cfgDW) getGPIOInputRouteSCI() bool {
	return data.getFieldVal(0, gpioInputRouteSCIMask, gpioInputRouteSCIShift) != 0
}

/*
GPIO Input Route SMI (GPIROUTSMI): Determines if the pad can be routed
to cause SMI when configured in GPIO input mode
This bit only applies to a GPIO that has SMI capability.
Otherwise, the bit is RO.
*/
func (data *cfgDW) getGPIOInputRouteSMI() bool {
	return data.getFieldVal(0, gpioInputRouteSMIMask, gpioInputRouteSMIShift) != 0
}

/*
GPIO Input Route NMI (GPIROUTNMI): Determines if the pad can be routed
to cause NMI when configured in GPIO input mode
This bit only applies to a GPIO that has NMI capability. Otherwise, the
bit is RO.
*/
func (data *cfgDW) getGPIOInputRouteNMI() bool {
	return data.getFieldVal(0, gpioInputRouteNMIMask, gpioInputRouteNMIShift) != 0
}

/*
This three-bit field determines whether the Pad is controlled by GPIO
controller logic or one of the native functions muxed onto the Pad.
0h = GPIO control the Pad.
1h = native function 1, if applicable, controls the Pad
2h = native function 2, if applicable, controls the Pad
3h = native function 3, if applicable, controls the Pad
4h = enable GPIO blink/PWM capability if applicable (note that not all
GPIOs have
blink/PWM capability)
Dedicated (unmuxed) GPIO shall report RO of all 0’s in this register
field If GPIO vs. native mode is configured via soft strap, this bit has
no effect. Default value is determined by the default functionality of the
pad.
*/
func (data *cfgDW) getPadMode() uint8 {
	return data.getFieldVal(0, padModeMask, padModeShift)
}

/*
GPIO RX Disable (GPIORXDIS): 0 = Enable the input buffer (active low
enable) of the pad.
1 = Disable the input buffer of the pad.
Notes: When the input buffer is disabled, the internal pad state is always
driven to '0'

GPIO TX Disable (GPIOTXDIS): 0 = Enable the output buffer (active low
enable) of the pad.

1 = Disable the output buffer of the pad; i.e. Hi-Z
*/
func (data *cfgDW) getGPIORxTxDisableStatus() uint8 {
	return data.getFieldVal(0, gpioRxTxDisableMask, gpioRxTxDisableShift)
}

/*
GPIO RX State (GPIORXSTATE): This is the current internal RX pad
state after Glitch Filter logic stage and is not affected by PMode and
RXINV settings.
*/
func (data *cfgDW) getGPIORXState() uint8 {
	return data.getFieldVal(0, gpioRxStateMask, gpioRxStateShift)
}

/*
GPIO TX State (GPIOTXSTATE):
0 = Drive a level '0' to the TX output pad.
1 = Drive a level '1' to the TX output pad
*/
func (data *cfgDW) getGPIOTXState() int {
	return int(data.getFieldVal(0, gpioTxStateMask, 0))
}

/*
Termination (TERM): The Pad Termination state defines the different weak
pull-up and pull-down settings that are supported by the buffer. The
settings for [13:10] correspond to:
0000: none
0010: 5k PD
0100: 20k PD
1010: 5k PU
1100: 20k PU
1111: Native controller selected by Pad Mode controls the Termination. This
setting needs to be set only for GPP_A1/LAD0, GPP_A2/LAD1, GPP_A3/LAD2,
GPP_A4/LAD3, GPP_D5/I2S0_SFRM, GPP_D6/I2S0_TXD, GPP_D7/I2S0_RXD, GPP_D8/
I2S0_CLK, GPD1 / ACPRESENT, and GPD2 / LAN_WAKE#, when the signals are used
as native function. Otherwise, the setting is reserved.
NOTES:
1. All other bit encodings are reserved.
2. If a reserved value is programmed, pad may malfunction.
3. The setting of this field is applicable in all Pad Mode including GPIO.
As each Pad Mode may require different termination and isolation, care must
be taken in SW in the transition with appropriate register programming. The
actual transition sequence requirement may vary on case by case basis
depending on the native functions involved. For example, before changing
the pad from output to input direction, PU/PD settings should be programmed
first to ensure the input does not float momentarily.
*/
func (data *cfgDW) getTermination() uint8 {
	return data.getFieldVal(1, termMask, termShift)
}

/*
Interrupt Select (INTSEL): The Interrupt Select defines which interrupt
line driven from the GPIO Controller toggles when an interrupt is detected
on this pad.
0 = Interrupt Line 0
1 = Interrupt Line 1
...
Up to the max IOxAPIC IRQ supported
*/
func (data *cfgDW) getInterruptSelect() uint8 {
	return data.getFieldVal(1, interruptSelectShift, 0)
}
