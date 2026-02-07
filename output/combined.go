package generated

type Car struct {
	Vehicle
	Engine Engine
	Wheel  [4]*Wheel
	Garage *Garage
	vin    string
}

func (c *Car) Drive(destination string) error {
	panic("TODO: implement")
}

type Engine struct {
}

func (e *Engine) Start() bool {
	panic("TODO: implement")
}

type Garage struct {
	Location string
}

type Person struct {
	Name string
}

func (p *Person) DriveCar(c Car) error {
	panic("TODO: implement")
}

type Vehicle struct {
}

type Wheel struct {
	Size int
}
