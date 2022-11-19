package types

const CONFIGURATION_FILENAME = "cody.yml"

type PortRange struct {
	Start int
	End   int
}

type Configuration struct {
	Ports PortRange
}

func (c Configuration) IsRangeValid() bool {
	return c.Ports.Start != 0 && c.Ports.End != 0
}
