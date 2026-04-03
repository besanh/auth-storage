package data

import (
	"context"

	"server/internal/biz"
	db "server/internal/data/db/generated"

	"github.com/go-kratos/kratos/v2/log"
)

type m2mAuthRepo struct {
	data *Data
	log  *log.Helper
}

func NewM2MAuthRepo(data *Data, logger log.Logger) biz.M2MAuthRepo {
	return &m2mAuthRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (r *m2mAuthRepo) GetMachineClientByID(ctx context.Context, clientID string) (*db.MachineClient, error) {
	client, err := r.data.Query.GetMachineClientByID(ctx, clientID)
	if err != nil {
		return nil, err
	}
	return &client, nil
}
