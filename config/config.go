package config

import "os"

const (
	TempInteltool  int  = 0
	TempGpioh      int  = 1
	TempSpec       int  = 2
)

var template int = 0

func TemplateSet(temp int) bool {
	if temp > TempSpec {
		return false
	} else {
		template = temp
		return true
	}
}

func TemplateGet() int {
	return template
}

const (
	SunriseType   uint8  = 0
	LewisburgType uint8  = 1
	ApolloType    uint8  = 2
)

var key uint8 = SunriseType

var platform = map[string]uint8{
	"snr": SunriseType,
	"lbg": LewisburgType,
	"apl": ApolloType}
func PlatformSet(name string) int {
	if platformType, valid := platform[name]; valid {
		key = platformType
		return 0
	}
	return -1
}
func PlatformGet() uint8 {
	return key
}
func IsPlatform(platformType uint8) bool {
	return platformType == key
}
func IsPlatformApollo() bool {
	return IsPlatform(ApolloType)
}
func IsPlatformSunrise() bool {
	return IsPlatform(SunriseType)
}
func IsPlatformLewisburg() bool {
	return IsPlatform(LewisburgType)
}

var InputRegDumpFile *os.File = nil
var OutputGenFile *os.File = nil

var ignoredFieldsFormat bool = false
func IgnoredFieldsFlagSet(flag bool) {
	ignoredFieldsFormat = flag
}
func AreFieldsIgnored() bool {
	return ignoredFieldsFormat
}

var nonCheckingFlag bool = false
func NonCheckingFlagSet(flag bool) {
	nonCheckingFlag = flag
}
func IsNonCheckingFlagUsed() bool {
	return nonCheckingFlag
}

var infolevel uint8 = 0
func InfoLevelSet(lvl uint8) {
	infolevel = lvl
}
func InfoLevelGet() uint8 {
	return infolevel
}

var fldstyle uint8 = CorebootFlds
const (
	NoFlds       uint8  = 0
	CorebootFlds uint8  = 1
	FspFlds      uint8  = 2
	RawFlds      uint8  = 3
)
var fldstylemap = map[string]uint8{
	"none" : NoFlds,
	"cb"   : CorebootFlds,
	"fsp"  : FspFlds,
	"raw"  : RawFlds}
func FldStyleSet(name string) int {
	if style, valid := fldstylemap[name]; valid {
		fldstyle = style
		return 0
	}
	return -1
}
func FldStyleGet() uint8 {
	return fldstyle
}
func IsFieldsMacroUsed() bool {
	return FldStyleGet() != NoFlds
}
func IsCorebootStyleMacro() bool {
	return FldStyleGet() == CorebootFlds
}
func IsFspStyleMacro() bool {
	return FldStyleGet() == FspFlds
}
func IsRawFields() bool {
	return FldStyleGet() == RawFlds
}
