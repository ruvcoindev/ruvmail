# ruvmail

It's email, but not as you know it.

## Introduction

ruvmail is a single-binary all-in-one mail transfer agent which sends and receives email natively over the [Ruvchain Network](https://Ruvchain-network.github.io/).

* ruvmail runs just about anywhere you like — your inbox is stored right on your own machine;
* Implements IMAP and SMTP protocols for sending and receiving mail, so you can use your favourite client (hopefully);
* Mails are exchanged between ruvmail users using built-in Ruvchain connectivity;
* All mail exchange traffic between any two ruvmail nodes is always end-to-end encrypted without exception;
* Ruvchain and ruvmail nodes on the same network are discovered automatically using multicast or you can configure a static Ruvchain peer.

Email addresses are based on your public key, like `89cd1ea25d99b8ccf29e454280313128c234ffb82aa0eb2e3496f6f156d063d0@ruvmail`.

## Why?

There are all sorts of messaging services in the world but there is still a lot of value in asynchronous communication. Email is something that a lot of people understand reasonably well and there is still a huge volume of software in the world which supports email. ruvmail is designed to comply with the standards that people know and expect.

Ruvchain is well-suited for ad-hoc mail delivery and allows ruvmail to work even in closed networks, where Internet or other connectivity is restricted or simply not available. It guarantees end-to-end encryption and handles networks with changing topologies reasonably well.

## Quickstart

Use a recent version of Go to install ruvmail:

```
go install github.com/ruvcoindev/ruvmail/cmd/ruvmail@latest
```

It will then be installed into your `GOPATH`, so add that to your environment:

```
export PATH=$PATH:`go env GOPATH`/bin
```

Create a mailbox and set your password. Your ruvmail database will automatically be created in your working directory if it doesn't already exist:

```
ruvmail -password
```

Start ruvmail, using the database in your working directory, with either multicast enabled, an [Ruvchain static peer](https://publicpeers.neilalexander.dev/) specified or both:

```
ruvmail -multicast
ruvmail -peer=tls://...
ruvmail -multicast -peer=tls://...
```

Your mail address will be printed in the log at startup. You will also use this as your username when you log into SMTP/IMAP.

Connect your mail client to ruvmail. In the above example:

* SMTP is listening on `localhost` port 1025, username is your mail address, plain password authentication, no SSL/TLS
* IMAP is listening on `localhost` port 1143, username is your mail address, plain password authentication, no SSL/TLS

Then try sending a mail to another ruvmail user!

## Parameters

The following command line switches are supported by the `ruvmail` binary:

* `-peer=tls://...` or `-peer=tcp://...` — connect to a specific Ruvchain node, like one of the [Public Peers](https://publicpeers.neilalexander.dev/);
* `-multicast` - enable multicast peer discovery for Ruvchain nodes on your LAN
* `-database=/path/to/ruvmail.db` — use a specific database file;
* `-smtp=listenaddr:port` — listen for SMTP on a specific address/port
* `-imap=listenaddr:port` — listen for IMAP on a specific address/port;
* `-password` — set your IMAP/SMTP password (doesn't matter if ruvmail is running or not, just make sure that ruvmail is pointing at the right database file or that you are in the right working directory).

## Notes

There are a few important notes:

* ruvmail needs to be running in order to receive inbound emails — it's therefore important to run ruvmail somewhere that will have good uptime;
* ruvmail tries to guarantee that senders are who they say they are. Your `From` address must be your ruvmail address;
* You can only email other ruvmail users, not regular email addresses on the public Internet;
* You may need to configure your client to allow "insecure" or "plaintext" authentication to IMAP/SMTP — this is because we don't support SSL/TLS on the IMAP/SMTP listeners yet;
* ruvmail won't transport mails larger than 1MB right now.

## Bugs

There are probably all sorts of bugs, but the ones that we know of are:

* IMAP behaviour might not be entirely spec-compliant in all cases, so your mileage with mail clients might vary;
* IMAP search isn't implemented yet and will instead return all mails.

The code's also a bit of a mess, so sorry about that.
