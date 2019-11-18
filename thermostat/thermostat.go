package thermostat

type ThermostatDevice interface {
	CurrentTemp() (temp *float64, err error)
	Connect() (err error)
}
