package localrun

//
// NOTE: currently unused, but might need something similar later?
//

var suburbs = []string{
	"suburbs",
	"ready",
	"start",
	"modern",
	"man",
	"rococo",
	"empty",
	"room",
	"city",
	"children",
	"half",
	"light",
	"war",
	"month",
	"may",
	"wasted",
	"hours",
	"deep",
	"blue",
	"wait",
	"sprawl",
}

var sonic = []string{
	"emerald",
	"hill",
	"chemical",
	"plant",
	"aquatic",
	"ruin",
	"casino",
	"night",
	"hill",
	"top",
	"mystic",
	"cave",
	"oil",
	"ocean",
	"metropolis",
	"wing",
	"fortress",
	"death",
	"egg",
}

var deschutes = []string{
	"black",
	"butte",
	"mirror",
	"pond",
	"inversion",
	"obsidian",
	"chainbreaker",
	"deschutes",
	"river",
	"red",
	"chair",
	"twilight",
	"jubelale",
	"hop",
	"henge",
	"trip",
	"fresh",
	"squeezed",
	"chasin",
	"freshies",
	"abyss",
	"dissident",
	"stoic",
	"green",
	"monster",
}

// containers need to be given unique names or the docker
// daemon will complain
func GenerateContainerName(seed int) string {
	suburbsVal := seed % len(suburbs)
	sonicVal := (seed / len(suburbs)) % len(sonic)
	deschutesVal := ((seed / len(suburbs)) / len(sonic)) % len(deschutes)

	return "earthkit_" + suburbs[suburbsVal] + "_" + sonic[sonicVal] + "_" + deschutes[deschutesVal]
}