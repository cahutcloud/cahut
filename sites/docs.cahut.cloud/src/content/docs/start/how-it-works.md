---
title: How Cahut works
description: AIs connect to your Cahut over MCP, publish Windows through bounded operations, and every call goes to the Logbook.
---

Cahut hosts what AIs make. It does not make it: the AIs stay outside, on your laptop or with their vendor, and connect to Cahut when they have something to publish. No AI model runs on the server.

## The path of a Window

1. **You install Cahut** on a Linux server, and declare a Cahut at a domain that points to it, such as `cahut.example.com`.
2. **You give your AI the address** of the Cahut, `https://cahut.example.com/mcp`. It sends you to the console, where you sign in and choose the Rooms it reaches and whether it may publish. It acts for you, in that Cahut only.
3. **The AI connects** over MCP. It sees a few tools and nothing else: no shell, no files of the server.
4. **The AI publishes a Window**: the HTML of a document, deck, dashboard or site, and its other files. Cahut keeps them and returns the address where the Window opens, such as `https://cahut.example.com/you/q3-review/`.
5. **You share the link.** Readers open the Window in their browser. Its pages run in a sandbox, apart from the other Windows.
6. **Every call is logged.** The Logbook of the Cahut records which AI acted, on whose behalf, what it did and with what result.

## One server, one or more Cahuts

A server runs one `cahut` service, which holds one or more Cahuts. Each Cahut has its own domain, its own people and tokens, its own Windows and its own Logbook, and shares nothing else with the others. See [Cahuts and servers](/concepts/cahuts/).

## Bounded operations

An AI acts on a Cahut only through typed operations, such as `publish_window`. Whatever the way in, MCP for AIs or the command line for the administrator, each call goes through the same registry: it checks who calls, runs the operation and writes the Logbook. The [operations reference](/reference/operations/) lists them all.

What can be undone, the AI does on its own. What cannot waits for the person the AI acts for to approve it in the console: this rule is called Knock first. See [Knock first and the Logbook](/concepts/knock-first-and-logbook/).

## What Cahut does not do

- Create content: that is the AIs' job.
- Run an AI model on the server.
- Depend on one AI vendor, or keep work in a proprietary format: a Window is HTML and its files.
- Give an AI a shell on the server.
