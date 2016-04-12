package provider

type IaaS interface {
	SetVip() error
	DeleteVip() error
}
