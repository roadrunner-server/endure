package off

type serviceGraph struct {
	nodes map[string]interface{}
}

func (sg *serviceGraph) Push(name string, node interface{}) {

}

func (sg *serviceGraph) Depends(name string, depends ...string) {

}

func (sg *serviceGraph) Provides(name string, provides ...string) {

}
