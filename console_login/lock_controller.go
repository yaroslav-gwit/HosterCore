// Copyright 2023s The Hoster Authors. All rights reserved.
// Use of this source code is governed by a Apache License 2.0
// license that can be found in the LICENSE fil

package main

import (
	"time"

	"github.com/gcla/gowid"
)

type LockController struct {
	view   *LockView
	ticker *time.Ticker
}

func NewLockController() *LockController {
	res := &LockController{nil, nil}
	view := NewLockView(res)
	res.view = view

	return res
}

func (c *LockController) ResetLock(app gowid.IApp) {
	c.view.timeout = pinTimeout
	c.view.UpdateLock(app)
}

func (c *LockController) AnimateLock(app gowid.IApp) {
	c.ticker = time.NewTicker(time.Second * 1)
	go func() {
		for range c.ticker.C {
			app.Run(gowid.RunFunction(func(app gowid.IApp) {
				c.view.UpdateLock(app)
				app.Redraw()
			}))
		}
	}()
}

func (g *LockController) StopAnimation() {
	g.ticker.Stop()
}
