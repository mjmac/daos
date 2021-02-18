//
// (C) Copyright 2020-2021 Intel Corporation.
//
// SPDX-License-Identifier: BSD-2-Clause-Patent
// +build !firmware

package server

import (
	"context"
	"testing"

	"github.com/pkg/errors"

	"github.com/mjmac/soad/src/control/common"
	"github.com/mjmac/soad/src/control/logging"
	"github.com/mjmac/soad/src/control/server/config"
)

func TestCtlSvc_FirmwareQuery_Disabled(t *testing.T) {
	log, buf := logging.NewTestLogger(t.Name())
	defer common.ShowBufferOnFailure(t, buf)

	emptyCfg := config.DefaultServer()
	cs := mockControlService(t, log, emptyCfg, nil, nil, nil)

	result, err := cs.FirmwareQuery(context.TODO(), nil)
	if result != nil {
		t.Errorf("expected nil response, got %+v", result)
	}

	common.CmpErr(t, errors.New("not implemented"), err)
}

func TestCtlSvc_FirmwareUpdate_Disabled(t *testing.T) {
	log, buf := logging.NewTestLogger(t.Name())
	defer common.ShowBufferOnFailure(t, buf)

	emptyCfg := config.DefaultServer()
	cs := mockControlService(t, log, emptyCfg, nil, nil, nil)

	result, err := cs.FirmwareUpdate(context.TODO(), nil)
	if result != nil {
		t.Errorf("expected nil response, got %+v", result)
	}

	common.CmpErr(t, errors.New("not implemented"), err)
}
