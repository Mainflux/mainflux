package api

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/mainflux/mainflux/certs"
)

func issueCert(svc certs.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(addCertsReq)
		if err := req.validate(); err != nil {
			return nil, err
		}
		res, err := svc.IssueCert(ctx, req.token, req.ThingID, req.Valid, req.RsaBits, req.KeyType)
		if err != nil {
			return certsResponse{Error: err.Error()}, nil
		}

		return res, nil
	}
}

func listCertificates(svc certs.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(listReq)
		if err := req.validate(); err != nil {
			return nil, err
		}

		page, err := svc.ListCertificates(ctx, req.token, req.thingID, req.offset, req.limit)
		if err != nil {
			return certsPageRes{
				Error: err.Error(),
			}, err
		}
		res := certsPageRes{
			pageRes: pageRes{
				Total:  page.Total,
				Offset: page.Offset,
				Limit:  page.Limit,
			},
			Certs: []certsResponse{},
		}

		for _, cert := range page.Certs {
			view := certsResponse{
				Serial:  cert.Serial,
				ThingID: cert.ThingID,
			}
			res.Certs = append(res.Certs, view)
		}

		return res, nil
	}
}

func revokeCertificate(svc certs.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(revokeReq)
		if err := req.validate(); err != nil {
			return nil, err
		}

		return svc.RevokeCert(ctx, req.token, req.ThingID, req.CertSerial)

	}
}
