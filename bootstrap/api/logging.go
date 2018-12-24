//
// Copyright (c) 2018
// Mainflux
//
// SPDX-License-Identifier: Apache-2.0
//

// +build !test

package api

import (
	"fmt"
	"time"

	"github.com/mainflux/mainflux/bootstrap"

	log "github.com/mainflux/mainflux/logger"
)

var _ bootstrap.Service = (*loggingMiddleware)(nil)

type loggingMiddleware struct {
	logger log.Logger
	svc    bootstrap.Service
}

// NewLoggingMiddleware adds logging facilities to the core service.
func NewLoggingMiddleware(svc bootstrap.Service, logger log.Logger) bootstrap.Service {
	return &loggingMiddleware{logger, svc}
}

func (lm *loggingMiddleware) Add(key string, thing bootstrap.Config) (saved bootstrap.Config, err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method add for key %s and thing %s took %s to complete", key, saved.ID, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())

	return lm.svc.Add(key, thing)
}

func (lm *loggingMiddleware) View(key, id string) (saved bootstrap.Config, err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method view for key %s and thing %s took %s to complete", key, saved.ID, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())

	return lm.svc.View(key, id)
}

func (lm *loggingMiddleware) Update(key string, thing bootstrap.Config) (err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method update for key %s and thing %s took %s to complete", key, thing.ID, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())

	return lm.svc.Update(key, thing)
}

func (lm *loggingMiddleware) List(key string, filter bootstrap.Filter, offset, limit uint64) (res []bootstrap.Config, err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method list for key %s and offset %d and limit %d took %s to complete", key, offset, limit, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())

	return lm.svc.List(key, filter, offset, limit)
}

func (lm *loggingMiddleware) Remove(key, id string) (err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method remove for key %s and thing %s took %s to complete", key, id, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())

	return lm.svc.Remove(key, id)
}

func (lm *loggingMiddleware) Bootstrap(externalKey, externalID string) (cfg bootstrap.Config, err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method bootstrap for thing with external id %s took %s to complete", externalID, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())

	return lm.svc.Bootstrap(externalKey, externalID)
}

func (lm *loggingMiddleware) ChangeState(key, id string, state bootstrap.State) (err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method changeState for key %s and thing %s took %s to complete", key, id, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())

	return lm.svc.ChangeState(key, id, state)
}
