package config

// platform type constants
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
