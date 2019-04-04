# Prometheus Configuration Controller

This controller generates configuration for a promethues instance.
It provides new Custom Resources, and a validating admission webhook,  for
prometheus rule groups, and for scrape definitions. This allows the definitions
for rules and alerts to be kept with, and managed with, the applications for
which they are relevant.

The controller can be used as a sidecar, or as a standalone service.



