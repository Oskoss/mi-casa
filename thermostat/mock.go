package thermostat

type MockThermostat struct {
	Temperature float64
}

func (device *MockThermostat) CurrentTemp() (temp *float64, err error) {
	return &device.Temperature, nil
}

func (device *MockThermostat) Connect() (err error) {

	return nil
}
