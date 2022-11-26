# statuspages
Status pages library for go programs.

Back-end services usually export metrics and logs, but to be able to
understand what is going on inside them, it makes sense to create
status pages. These are simple web pages served directly by the binary
and exposing internal state, potentially also allowing to act on the
service for maintenance purposes: restart, change elected master,
reload configuration, etc.

This library provides tools to easily:
- Generate status pages with simple HTML code, forms and tables.
- Make your status pages secure, by authenticating the user or using
  a proxy for authentication. Generate also audit logs for any
  user action.

# Current status of the project

Draft

# Usage

TODO

# Login plugin

TODO

# Audit log plugin

TODO

