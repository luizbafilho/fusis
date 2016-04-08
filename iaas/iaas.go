package iaas

type IaaS interface {
	SetVip() error
	DeleteVip() error
}
