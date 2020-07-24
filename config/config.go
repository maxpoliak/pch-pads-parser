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

var useRawFormat bool = false
func RawFormatFlagSet(flag bool) {
	useRawFormat = flag
}
func IsRawFormatUsed() bool {
	return useRawFormat
}

var useCorebootFiledsFlag bool = false
func CorebootFiledsFormatSet(flag bool) {
	useCorebootFiledsFlag = flag
}
func IsCorebootFiledsFormatUsed() bool {
	return useCorebootFiledsFlag
}

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

var fspStyleUsed bool = false
func FspStyleMacroUsed(flag bool) {
	fspStyleUsed = flag
}
func IsFspStyleMacro() bool {
	return fspStyleUsed
}
