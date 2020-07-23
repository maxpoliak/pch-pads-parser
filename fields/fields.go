package fields

import "../config"
import "../common"

import "./fsp"
import "./cb"

// InterfaceSet - set the interface for decoding configuration
// registers DW0 and DW1.
func InterfaceSet() {
	macro := common.GetInstanceMacro()
	if config.IsFspStyleMacro() {
		macro.Fields = fsp.FieldMacros{}
	} else {
		macro.Fields = cb.FieldMacros{}
	}
}
