package thermostat

//ThermostatDevice is an interface which abstracts
//the underlying device complexities of a thermostat
type ThermostatDevice interface {
	CurrentTemp() (temp *float64, err error)
	Connect() (err error)
}
