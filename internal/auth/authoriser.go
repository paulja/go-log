package auth

import (
	_ "embed"
	"fmt"

	"github.com/casbin/casbin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	//go:embed model.conf
	ACLModel string
	//go:embed policy.csv
	ACLPolicy string
)

func NewDefault() *Authoriser {
	return New(ACLModel, ACLPolicy)
}

func New(model, policy string) *Authoriser {
	m := casbin.NewModel(ACLModel)
	a := NewStringAdapter(ACLPolicy)

	enf := &casbin.Enforcer{}
	enf.InitWithModelAndAdapter(m, a)

	return &Authoriser{
		enforcer: enf,
	}
}

type Authoriser struct {
	enforcer *casbin.Enforcer
}

func (a *Authoriser) Authorise(subject, object, action string) error {
	if !a.enforcer.Enforce(subject, object, action) {
		m := fmt.Sprintf("%s not permitted to %s to %s", subject, action, object)
		s := status.New(codes.PermissionDenied, m)
		return s.Err()
	}
	return nil
}
