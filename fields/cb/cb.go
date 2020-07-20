package cb

import "../../common"
import "../../config"

type fields struct {}

// DecodeDW0 - decode value of DW0 register
func (fields) DecodeDW0(macro *common.Macro) {
	dw0 := macro.Register(common.PAD_CFG_DW0)

	irqroute := func(prefix string) {
		macro.Or().Add("PAD_IRQ_ROUTE(").Add(prefix).Add(")")
	}

	somebits := func(prefix string) {
		macro.Or().Add(prefix)
	}

	for _, field := range []struct {
		prefix   string
		getvalue func() uint8
		generate func(prefix string)
	} {
		{	// PAD_FUNC(NF3)
			"PAD_FUNC",
			func() uint8 {
				if dw0.GetPadMode() != 0 || config.InfoLevelGet() <= 3 {
					return 1
				}
				return 0
			},
			func(prefix string) {
				macro.Or().Add(prefix).Add("(").Padfn().Add(")")
			},
		},
		{	// PAD_RESET(DEEP)
			"PAD_RESET",
			dw0.GetResetConfig,
			func(prefix string) {
				macro.Or().Add(prefix).Add("(").Rstsrc().Add(")")
			},
		},
		{	// PAD_TRIG(OFF)
			"PAD_TRIG",
			dw0.GetRXLevelEdgeConfiguration,
			func(prefix string) {
				macro.Or().Add(prefix).Add("(").Trig().Add(")")
			},
		},
		{	// PAD_IRQ_ROUTE(IOAPIC)
			"IOAPIC",
			dw0.GetGPIOInputRouteIOxAPIC,
			irqroute,
		},
		{	// PAD_IRQ_ROUTE(SCI)
			"SCI",
			dw0.GetGPIOInputRouteSCI,
			irqroute,
		},
		{	// PAD_IRQ_ROUTE(SMI)
			"SMI",
			dw0.GetGPIOInputRouteSMI,
			irqroute,
		},
		{	// PAD_IRQ_ROUTE(NMI)
			"NMI",
			dw0.GetGPIOInputRouteNMI,
			irqroute,
		},
		{	// PAD_RX_POL(EDGE_SINGLE)
			"PAD_RX_POL",
			dw0.GetRxInvert,
			func(prefix string) {
				macro.Or().Add(prefix).Add("(").Invert().Add(")")
			},
		},
		{	// PAD_BUF(TX_RX_DISABLE)
			"PAD_BUF",
			dw0.GetGPIORxTxDisableStatus,
			func(prefix string) {
				macro.Or().Add(prefix).Add("(").Bufdis().Add(")")
			},
		},
		{	// (1 << 29)
			"(1 << 29)",
			dw0.GetRXPadStateSelect,
			somebits,
		},
		{	// (1 << 28)
			"(1 << 28)",
			dw0.GetRXRawOverrideStatus,
			somebits,
		},
		{	// 1
			"1",
			dw0.GetGPIOTXState,
			somebits,
		},
	} {
		if field.getvalue() != 0 {
			field.generate(field.prefix)
		}
	}
}

// DecodeDW1 - decode value of DW1 register
func (fields) DecodeDW1(macro *common.Macro) {
	dw1 := macro.Register(common.PAD_CFG_DW1)
	for _, field := range []struct {
		prefix   string
		getvalue func() uint8
		generate func(prefix string)
	} {
		{	// PAD_IOSSTATE(HIZCRx0)
			"PAD_PULL",
			dw1.GetTermination,
			func(prefix string) {
				macro.Or().Add(prefix).Add("(").Pull().Add(")")
			},
		},
		{	// PAD_IOSSTATE(HIZCRx0)
			"PAD_IOSSTATE",
			dw1.GetIOStandbyState,
			func(prefix string) {
				macro.Or().Add(prefix).Add("(").IOSstate().Add(")")
			},
		},
		{	// PAD_IOSTERM(ENPU)
			"PAD_IOSTERM",
			dw1.GetIOStandbyTermination,
			func(prefix string)  {
				macro.Or().Add(prefix).Add("(").IOTerm().Add(")")
			},
		},
		{	// PAD_CFG_OWN_GPIO(DRIVER)
			"PAD_CFG_OWN_GPIO",
			func() uint8 {
				if macro.IsOwnershipDriver() {
					return 1
				}
				return 0
			},
			func(prefix string)  {
				macro.Or().Add(prefix).Add("(").Own().Add(")")
			},
		},
	} {
		if field.getvalue() != 0 {
			field.generate(field.prefix)
		}
	}
}

// SetFieldsIface - set the interface for decoding configuration
// registers DW0 and DW1.
func SetFieldsIface(macro *common.Macro) {
	macro.Fields = fields{}
}
