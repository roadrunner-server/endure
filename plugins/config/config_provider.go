package config

type Provider struct {

}

func (p *Provider) Init() error {
	return nil
}

func (p *Provider) Serve() chan error {
	errCh := make(chan error)
	return errCh
}

func (p *Provider) Configure() error {
	return nil
}

func (p *Provider) Stop() error {
	return nil
}
