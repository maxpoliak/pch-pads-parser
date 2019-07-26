package main

const (
	PADRSTCFG_SHIFT uint8  = 30
	PADRSTCFG_MASK  uint32 = 0x3 << PADRSTCFG_SHIFT

	RXPADSTSEL_SHIFT uint8  = 29
	RXPADSTSEL_MASK  uint32 = 0x1 << RXPADSTSEL_SHIFT

	RXRAW1_SHIFT uint8  = 28
	RXRAW1_MASK  uint32 = 0x1 << RXRAW1_SHIFT

	RXEVCFG_SHIFT uint8  = 25
	RXEVCFG_MASK  uint32 = 0x3 << RXEVCFG_SHIFT

	RXINV_SHIFT uint8  = 23
	RXINV_MASK  uint32 = 0x1 << RXINV_SHIFT

	GPIROUTIOXAPIC_SHIFT uint8  = 20
	GPIROUTIOXAPIC_MASK  uint32 = 0x1 << GPIROUTIOXAPIC_SHIFT

	GPIROUTSCI_SHIFT uint8  = 19
	GPIROUTSCI_MASK  uint32 = 0x1 << GPIROUTSCI_SHIFT

	GPIROUTSMI_SHIFT uint8  = 18
	GPIROUTSMI_MASK  uint32 = 0x1 << GPIROUTSMI_SHIFT

	GPIROUTNMI_SHIFT uint8  = 17
	GPIROUTNMI_MASK  uint32 = 0x1 << GPIROUTNMI_SHIFT

	PMODE_SHIFT uint8  = 10
	PMODE_MASK  uint32 = 0x7 << PMODE_SHIFT

	GPIORXTXDIS_SHIFT uint8  = 8
	GPIORXTXDIS_MASK  uint32 = 0x3 << GPIORXTXDIS_SHIFT

	GPIORXSTATE_SHIFT uint8  = 1
	GPIORXSTATE_MASK  uint32 = 0x1 << GPIORXSTATE_SHIFT

	GPIOTXSTATE_MASK uint32 = 0x1
)

const (
	TERM_SHIFT  uint8  = 10
	TERM_MASK   uint32 = 0xF << TERM_SHIFT
	INTSEL_MASK uint32 = 0xFF
)

const (
	MAX_DW = 2
)

type configData struct {
	dw [MAX_DW]uint32
}

func (data *configData) getFieldVal(regNum uint8, mask uint32, shift uint8) uint8 {
	if regNum < MAX_DW {
		return uint8((data.dw[regNum] & mask) >> shift)
	}
	return 0
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
func (data *configData) getResetConfig() uint8 {
	return data.getFieldVal(0, PADRSTCFG_MASK, PADRSTCFG_SHIFT)
}

/*
RX Pad State Select (RXPADSTSEL): Determines from which node the RX pad
state for native function should be taken from. This field only affects the
pad state value being fanned out to native function(s) and isn`t meaningful
if the pad is in GPIO mode (i.e. Pad Mode = 0)
0 = Raw RX pad state directly from RX buffer
1 = Internal RX pad state (subject to RXINV and PreGfRXSel settings)
*/
func (data *configData) getRXPadStateSelect() uint8 {
	return data.getFieldVal(0, RXPADSTSEL_MASK, RXPADSTSEL_SHIFT)
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
func (data *configData) getRXRawOverrideStatus() uint8 {
	return data.getFieldVal(0, RXRAW1_MASK, RXRAW1_SHIFT)
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
func (data *configData) getRXLevelEdgeConfiguration() uint8 {
	return data.getFieldVal(0, RXEVCFG_MASK, RXEVCFG_SHIFT)
}

/*
RX Invert (RXINV): This bit determines if the selected pad state should
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
func (data *configData) getRXLevelConfiguration() bool {
	return data.getFieldVal(0, RXINV_MASK, RXINV_SHIFT) != 0
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
func (data *configData) getGPIOInputRouteIOxAPIC() bool {
	return data.getFieldVal(0, GPIROUTIOXAPIC_MASK, GPIROUTIOXAPIC_SHIFT) != 0
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
func (data *configData) getGPIOInputRouteSCI() bool {
	return data.getFieldVal(0, GPIROUTSCI_MASK, GPIROUTSCI_SHIFT) != 0
}

/*
GPIO Input Route SMI (GPIROUTSMI): Determines if the pad can be routed
to cause SMI when configured in GPIO input mode
This bit only applies to a GPIO that has SMI capability.
Otherwise, the bit is RO.
*/
func (data *configData) getGPIOInputRouteSMI() bool {
	return data.getFieldVal(0, GPIROUTSMI_MASK, GPIROUTSMI_SHIFT) != 0
}

/*
GPIO Input Route NMI (GPIROUTNMI): Determines if the pad can be routed
to cause NMI when configured in GPIO input mode
This bit only applies to a GPIO that has NMI capability. Otherwise, the
bit is RO.
*/
func (data *configData) getGPIOInputRouteNMI() bool {
	return data.getFieldVal(0, GPIROUTNMI_MASK, GPIROUTNMI_SHIFT) != 0
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
func (data *configData) getPadMode() uint8 {
	return data.getFieldVal(0, PMODE_MASK, PMODE_SHIFT)
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
func (data *configData) getGPIORxTxDisableStatus() uint8 {
	return data.getFieldVal(0, GPIORXTXDIS_MASK, GPIORXTXDIS_SHIFT)
}

/*
GPIO RX State (GPIORXSTATE): This is the current internal RX pad
state after Glitch Filter logic stage and is not affected by PMode and
RXINV settings.
*/
func (data *configData) getGPIORXState() uint8 {
	return data.getFieldVal(0, GPIORXSTATE_MASK, GPIORXSTATE_SHIFT)
}

/*
GPIO TX State (GPIOTXSTATE):
0 = Drive a level '0' to the TX output pad.
1 = Drive a level '1' to the TX output pad
*/
func (data *configData) getGPIOTXState() int {
	return int(data.getFieldVal(0, GPIOTXSTATE_MASK, 0))
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
func (data *configData) getTermination() uint8 {
	return data.getFieldVal(1, TERM_MASK, TERM_SHIFT)
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
func (data *configData) getInterruptSelect() uint8 {
	return data.getFieldVal(1, INTSEL_MASK, 0)
}
