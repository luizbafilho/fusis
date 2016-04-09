package infra

type IaaS interface {
	SetVip() error
	DeleteVip() error
}
