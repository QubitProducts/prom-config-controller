package main

import (
	"time"

	"github.com/prometheus/prometheus/promql"
)

type ruleDAGNode struct {
	exp  promql.Expr
	deps []string
}

func findDeps(exp promql.Expr) []string {
	var deps []string
	promql.Inspect(exp, func(node promql.Node, path []promql.Node) error {
		switch n := node.(type) {
		case *promql.VectorSelector:
			deps = append(deps, n.Name)
		case *promql.MatrixSelector:
			deps = append(deps, n.Name)
		}
		return nil
	})

	return deps
}

func calcMaxOffset(exp promql.Expr) time.Duration {
	var maxOffset time.Duration
	promql.Inspect(exp, func(node promql.Node, path []promql.Node) error {
		subqOffset := cumulativeSubqueryOffset(path)
		switch n := node.(type) {
		case *promql.VectorSelector:
			if maxOffset < promql.LookbackDelta+subqOffset {
				maxOffset = promql.LookbackDelta + subqOffset
			}
			if n.Offset+promql.LookbackDelta+subqOffset > maxOffset {
				maxOffset = n.Offset + promql.LookbackDelta + subqOffset
			}
		case *promql.MatrixSelector:
			if maxOffset < n.Range+subqOffset {
				maxOffset = n.Range + subqOffset
			}
			if n.Offset+n.Range+subqOffset > maxOffset {
				maxOffset = n.Offset + n.Range + subqOffset
			}
		}
		return nil
	})

	return maxOffset
}

func cumulativeSubqueryOffset(path []promql.Node) time.Duration {
	var subqOffset time.Duration
	for _, node := range path {
		switch n := node.(type) {
		case *promql.SubqueryExpr:
			subqOffset += n.Range + n.Offset
		}
	}
	return subqOffset
}
