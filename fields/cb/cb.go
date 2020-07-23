package cb

import "../../common"
import "../../config"

type FieldMacros struct {}

// field - data structure for creating a new bitfield macro object
// PAD_FUNC(NF3)
// prefix       : PAD_FUNC
// name         : NF3; this value will be overridden if the configurator is used
// unhide       : conditions for hiding macros
// configurator : method for determining the current configuration of the bit field
type field struct {
	prefix       string
	name         string
	unhide       bool
	configurator func()
}

// generate - wrapper for generating bitfield macros string
// fileds : field structure
func generate(macro *common.Macro, fileds ...*field) {
	for _, field := range fileds {
		if field.unhide {
			macro.Or()
			if field.prefix != "" {
				macro.Add(field.prefix).Add("(")
			}
			if field.name != "" {
				macro.Add(field.name)
			} else if field.configurator != nil {
				field.configurator()
			}
			if field.prefix != "" {
				macro.Add(")")
			}
		}
	}
}

// DecodeDW0 - decode value of DW0 register
func (FieldMacros) DecodeDW0(macro *common.Macro) {
	dw0 := macro.Register(common.PAD_CFG_DW0)
	generate(macro,
		&field {
			prefix : "PAD_FUNC",
			unhide : config.InfoLevelGet() <= 3 || dw0.GetPadMode() != 0,
			configurator : func() { macro.Padfn() },
		},

		&field {
			prefix : "PAD_RESET",
			unhide : dw0.GetResetConfig() != 0,
			configurator : func() { macro.Rstsrc() },
		},

		&field {
			prefix : "PAD_TRIG",
			unhide : dw0.GetRXLevelEdgeConfiguration() != 0,
			configurator : func() { macro.Trig() },
		},

		&field {
			prefix : "PAD_IRQ_ROUTE",
			name   : "IOAPIC",
			unhide : dw0.GetGPIOInputRouteIOxAPIC() != 0,
		},

		&field {
			prefix : "PAD_IRQ_ROUTE",
			name   : "SCI",
			unhide : dw0.GetGPIOInputRouteSCI() != 0,
		},

		&field {
			prefix : "PAD_IRQ_ROUTE",
			name   : "SMI",
			unhide : dw0.GetGPIOInputRouteSMI() != 0,
		},

		&field {
			prefix : "PAD_IRQ_ROUTE",
			name   : "NMI",
			unhide : dw0.GetGPIOInputRouteNMI() != 0,
		},

		&field {
			prefix : "PAD_RX_POL",
			unhide : dw0.GetRxInvert() != 0,
			configurator : func() { macro.Invert() },
		},

		&field {
			prefix : "PAD_BUF",
			unhide : dw0.GetGPIORxTxDisableStatus() != 0,
			configurator : func() { macro.Bufdis() },
		},

		&field {
			name   : "(1 << 29)",
			unhide : dw0.GetRXPadStateSelect() != 0,
		},

		&field {
			name   : "(1 << 28)",
			unhide : dw0.GetRXRawOverrideStatus() != 0,
		},

		&field {
			name   : "1",
			unhide : dw0.GetGPIOTXState() != 0,
		},
	)
}

// DecodeDW1 - decode value of DW1 register
func (FieldMacros) DecodeDW1(macro *common.Macro) {
	dw1 := macro.Register(common.PAD_CFG_DW1)
	generate(macro,
		&field {
			prefix : "PAD_PULL",
			unhide : dw1.GetTermination() != 0,
			configurator : func() { macro.Pull() },
		},

		&field {
			prefix : "PAD_IOSSTATE",
			unhide : dw1.GetIOStandbyState() != 0,
			configurator : func() { macro.IOSstate() },
		},

		&field {
			prefix : "PAD_IOSTERM",
			unhide : dw1.GetIOStandbyTermination() != 0,
			configurator : func() { macro.IOTerm() },
		},

		&field {
			prefix : "PAD_CFG_OWN_GPIO",
			unhide : macro.IsOwnershipDriver(),
			configurator : func() { macro.Own() },
		},
	)
}

// SetFieldsIface - set the interface for decoding configuration
// registers DW0 and DW1.
func SetFieldsIface(macro *common.Macro) {
	macro.Fields = FieldMacros{}
}
