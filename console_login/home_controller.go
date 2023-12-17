// Copyright 2023s The Hoster Authors. All rights reserved.
// Use of this source code is governed by a Apache License 2.0
// license that can be found in the LICENSE fil

package main

type HomeController struct {
	view *HomeView
}

func NewHomeController() *HomeController {
	res := &HomeController{nil}
	view := NewHomeView(res)
	res.view = view

	return res
}
