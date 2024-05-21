/*
 *  Copyright (c) 2021 Neil Alexander
 *  Copyright (c) 2024 ruvcoindev
 *  This Source Code Form is subject to the terms of the Mozilla Public
 *  License, v. 2.0. If a copy of the MPL was not distributed with this
 *  file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

package smtpserver

import (
	"github.com/emersion/go-smtp"
	"github.com/ruvcoindev/ruvmail/internal/imapserver"
)

type SMTPServer struct {
	server  *smtp.Server
	backend smtp.Backend
	notify  *imapserver.IMAPNotify
}

func NewSMTPServer(backend smtp.Backend, notify *imapserver.IMAPNotify) *SMTPServer {
	s := &SMTPServer{
		server:  smtp.NewServer(backend),
		backend: backend,
		notify:  notify,
	}
	return s
}
