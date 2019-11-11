package main

import (
	"testing"

	"github.com/prometheus/prometheus/promql"
)

func TestFilterRules(t *testing.T) {
	strExprs := []string{
		`sum_over_time(up{job="blah"}[72h])`,
		`sum_over_time(up{job="blah"}[72h] offset 1h)`,
		`rate(http_requests_total[5m])[30m:1m]`,
	}
	for _, s := range strExprs {
		expr, _ := promql.ParseExpr(s)
		t.Logf("d: %v", calcMaxOffset(expr))
	}
}

func TestFindDeps(t *testing.T) {
	strExprs := []string{
		`sum_over_time(up{job="blah"}[72h])`,
		`sum_over_time(up{job="blah"}[72h] offset 1h)`,
		`rate(http_requests_total[5m])[30m:1m]`,
	}
	for _, s := range strExprs {
		expr, _ := promql.ParseExpr(s)
		t.Logf("deps: %v", findDeps(expr))
	}
}
