package pkg

import (
	"context"

	"github.com/cloudskiff/driftctl/pkg/remote"
	"github.com/pkg/errors"

	"github.com/cloudskiff/driftctl/pkg/parallel"
	"github.com/sirupsen/logrus"

	"github.com/cloudskiff/driftctl/pkg/alerter"
	"github.com/cloudskiff/driftctl/pkg/resource"
)

type Scanner struct {
	resourceSuppliers []resource.Supplier
	runner            *parallel.ParallelRunner
	alerter           *alerter.Alerter
}

func NewScanner(resourceSuppliers []resource.Supplier, alerter *alerter.Alerter) *Scanner {
	return &Scanner{
		resourceSuppliers: resourceSuppliers,
		runner:            parallel.NewParallelRunner(context.TODO(), 10),
		alerter:           alerter,
	}
}

func (s *Scanner) Resources() ([]resource.Resource, error) {
	for _, resourceProvider := range s.resourceSuppliers {
		supplier := resourceProvider
		s.runner.Run(func() (interface{}, error) {
			res, err := supplier.Resources()
			if err != nil {
				err := remote.HandleResourceEnumerationError(err, s.alerter)
				if err == nil {
					return []resource.Resource{}, nil
				}
				return nil, err
			}
			for _, resource := range res {
				logrus.WithFields(logrus.Fields{
					"id":   resource.TerraformId(),
					"type": resource.TerraformType(),
				}).Debug("Found cloud resource")
			}
			return res, nil
		})
	}

	results := make([]resource.Resource, 0)
loop:
	for {
		select {
		case resources, ok := <-s.runner.Read():
			if !ok || resources == nil {
				break loop
			}
			for _, res := range resources.([]resource.Resource) {
				normalisable, ok := res.(resource.NormalizedResource)
				if ok {
					normalizedRes, err := normalisable.NormalizeForProvider()

					if err != nil {
						logrus.Errorf("Could not normalize remote for res %s: %+v", res.TerraformId(), err)
						results = append(results, res)
					}

					if err == nil {
						results = append(results, normalizedRes)
					}
				}

				if !ok {
					results = append(results, res)
				}
			}
		case <-s.runner.DoneChan():
			break loop
		}
	}
	return results, s.runner.Err()
}

func (s *Scanner) Stop() {
	logrus.Debug("Stopping scanner")
	s.runner.Stop(errors.New("interrupted"))
}
