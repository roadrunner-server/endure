package cascade

// + Init() <<--

type Cascade struct {
	providers []Provider
	registers []Register
	services  *serviceGraph
}

func NewContainer() *Cascade {
	return &Cascade{
		registers: nil,
		providers: nil,
		services:  nil,
	}
}

func (c *Cascade) Register(name string, service Service) {
	c.services.Push(name, service)

	if r, ok := service.(Provider); ok {
		c.providers = append(c.providers, r)
	}

	if r, ok := service.(Register); ok {
		c.registers = append(c.registers, r)
	}
}

func (c *Cascade) Init() {
	for name, svc := range c.services.nodes {

	}

	// todo: SORT BASED ON INIT AND BASED ON PROVIDERS AND BASED ON REGISTERS
	//for _, e := range c.services.Sort() {
	//	// inject service dependencies
	//	if ok, err := c.initService(e.svc, cfg.Get(e.name)); err != nil {
	//		// soft error (skipping)
	//		if err == errNoConfig {
	//			c.log.Debugf("[%s]: disabled", e.name)
	//			continue
	//		}
	//
	//		return errors.Wrap(err, fmt.Sprintf("[%s]", e.name))
	//	} else if ok {
	//		e.setStatus(StatusOK)
	//	} else {
	//		c.log.Debugf("[%s]: disabled", e.name)
	//	}
	//}
}

//// calls Init method with automatically resolved arguments.
//func (c *container) initService(s interface{}, segment Config) (bool, error) {
//	r := reflect.TypeOf(s)
//
//	m, ok := r.MethodByName(InitMethod)
//	if !ok {
//		// no Init method is presented, assuming service does not need initialization.
//		return true, nil
//	}
//
//	if err := c.verifySignature(m); err != nil {
//		return false, err
//	}
//
//	// hydrating
//	values, err := c.resolveValues(s, m, segment)
//	if err != nil {
//		return false, err
//	}
//
//	// initiating service
//	out := m.Func.Call(values)
//
//	if out[1].IsNil() {
//		return out[0].Bool(), nil
//	}
//
//	return out[0].Bool(), out[1].Interface().(error)
//}
//
//// resolveValues returns slice of call arguments for service Init method.
//func (c *container) resolveValues(s interface{}, m reflect.Method, cfg Config) (values []reflect.Value, err error) {
//	for i := 0; i < m.Type.NumIn(); i++ {
//		v := m.Type.In(i)
//
//		switch {
//		case v.ConvertibleTo(reflect.ValueOf(s).Type()): // service itself
//			values = append(values, reflect.ValueOf(s))
//
//		case v.Implements(reflect.TypeOf((*Container)(nil)).Elem()): // container
//			values = append(values, reflect.ValueOf(c))
//
//		case v.Implements(reflect.TypeOf((*logrus.StdLogger)(nil)).Elem()),
//			v.Implements(reflect.TypeOf((*logrus.FieldLogger)(nil)).Elem()),
//			v.ConvertibleTo(reflect.ValueOf(c.log).Type()): // logger
//			values = append(values, reflect.ValueOf(c.log))
//
//		case v.Implements(reflect.TypeOf((*HydrateConfig)(nil)).Elem()): // injectable config
//			sc := reflect.New(v.Elem())
//
//			if dsc, ok := sc.Interface().(DefaultsConfig); ok {
//				err := dsc.InitDefaults()
//				if err != nil {
//					return nil, err
//				}
//				if cfg == nil {
//					values = append(values, sc)
//					continue
//				}
//
//			} else if cfg == nil {
//				return nil, errNoConfig
//			}
//
//			if err := sc.Interface().(HydrateConfig).Hydrate(cfg); err != nil {
//				return nil, err
//			}
//
//			values = append(values, sc)
//
//		case v.Implements(reflect.TypeOf((*Config)(nil)).Elem()): // generic config section
//			if cfg == nil {
//				return nil, errNoConfig
//			}
//
//			values = append(values, reflect.ValueOf(cfg))
//
//		default: // dependency on other service (resolution to nil if service can't be found)
//			value, err := c.resolveValue(v)
//			if err != nil {
//				return nil, err
//			}
//
//			values = append(values, value)
//		}
//	}
//
//	return
//}
//
//// verifySignature checks if Init method has valid signature
//func (c *container) verifySignature(m reflect.Method) error {
//	if m.Type.NumOut() != 2 {
//		return fmt.Errorf("method Init must have exact 2 return values")
//	}
//
//	if m.Type.Out(0).Kind() != reflect.Bool {
//		return fmt.Errorf("first return value of Init method must be bool type")
//	}
//
//	if !m.Type.Out(1).Implements(reflect.TypeOf((*error)(nil)).Elem()) {
//		return fmt.Errorf("second return value of Init method value must be error type")
//	}
//
//	return nil
//}
//
//func (c *container) resolveValue(v reflect.Type) (reflect.Value, error) {
//	value := reflect.Value{}
//	for _, e := range c.services {
//		if !e.hasStatus(StatusOK) {
//			continue
//		}
//
//		if v.Kind() == reflect.Interface && reflect.TypeOf(e.svc).Implements(v) {
//			if value.IsValid() {
//				return value, fmt.Errorf("disambiguous dependency `%s`", v)
//			}
//
//			value = reflect.ValueOf(e.svc)
//		}
//
//		if v.ConvertibleTo(reflect.ValueOf(e.svc).Type()) {
//			if value.IsValid() {
//				return value, fmt.Errorf("disambiguous dependency `%s`", v)
//			}
//
//			value = reflect.ValueOf(e.svc)
//		}
//	}
//
//	if !value.IsValid() {
//		// placeholder (make sure to check inside the method)
//		value = reflect.New(v).Elem()
//	}
//
//	return value, nil
//}
