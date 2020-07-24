package fields

import "../config"
import "../platforms/common"

import "./fsp"
import "./cb"

// InterfaceSet - set the interface for decoding configuration
// registers DW0 and DW1.
func InterfaceGet() common.Fields {
	if config.IsFspStyleMacro() {
		return fsp.FieldMacros{}
	} else {
		return cb.FieldMacros{}
	}
}
