package switcher

//SwitchDevice represents a basic switch usually in an "on"/"off" state
// but able to be in any number of states the implementations require
type SwitchDevice interface {
	CurrentStatus() (status *string, err error)
	TurnOn() (err error)
	TurnOff() (err error)
}
