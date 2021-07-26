package maasclient

import "net/url"

type params struct {
	values url.Values
}

func ParamsBuilder() Params {
	return &params{values: url.Values{}}
}

func (p *params) Add(key, value string) Params {
	p.values.Add(key, value)
	return p
}

func (p *params) Set(key, value string) Params {
	p.values.Set(key, value)
	return p
}

func (p *params) Reset() {
	p.values = url.Values{}
}

func (p *params) Values() url.Values {
	return p.values
}

func (p *params) Copy(in Params) {
	for key, value := range in.Values() {
		p.values[key] = value
	}
}

type Params interface {
	Add(key, value string) Params
	Set(key, value string) Params
	Reset()
	Values() url.Values
	Copy(in Params)
}
