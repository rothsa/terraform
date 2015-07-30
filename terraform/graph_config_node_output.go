package terraform

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform/config"
	"github.com/hashicorp/terraform/dag"
	"github.com/hashicorp/terraform/dot"
)

// GraphNodeConfigOutput represents an output configured within the
// configuration.
type GraphNodeConfigOutput struct {
	Output *config.Output
}

func (n *GraphNodeConfigOutput) Name() string {
	return fmt.Sprintf("output.%s", n.Output.Name)
}

func (n *GraphNodeConfigOutput) ConfigType() GraphNodeConfigType {
	return GraphNodeConfigTypeOutput
}

func (n *GraphNodeConfigOutput) OutputName() string {
	return n.Output.Name
}

func (n *GraphNodeConfigOutput) DependableName() []string {
	return []string{n.Name()}
}

func (n *GraphNodeConfigOutput) DependentOn() []string {
	vars := n.Output.RawConfig.Variables
	result := make([]string, 0, len(vars))
	for _, v := range vars {
		if vn := varNameForVar(v); vn != "" {
			result = append(result, vn)
		}
	}

	return result
}

// GraphNodeEvalable impl.
func (n *GraphNodeConfigOutput) EvalTree() EvalNode {
	return &EvalOpFilter{
		Ops: []walkOperation{walkRefresh, walkPlan, walkApply, walkDestroy},
		Node: &EvalSequence{
			Nodes: []EvalNode{
				&EvalWriteOutput{
					Name:  n.Output.Name,
					Value: n.Output.RawConfig,
				},
			},
		},
	}
}

// GraphNodeProxy impl.
func (n *GraphNodeConfigOutput) Proxy() bool {
	return true
}

// GraphNodeDestroyEdgeInclude impl.
func (n *GraphNodeConfigOutput) DestroyEdgeInclude(dag.Vertex) bool {
	return false
}

// GraphNodeDotter impl.
func (n *GraphNodeConfigOutput) DotNode(name string, opts *GraphDotOpts) *dot.Node {
	return dot.NewNode(name, map[string]string{
		"label":     n.OutputName(),
		"shape":     "invhouse",
		"style":     "filled",
		"fillcolor": "#FFB2B2",
		"fontname":  "courier",
	})
}

// GraphNodeFlattenable impl.
func (n *GraphNodeConfigOutput) Flatten(p []string) (dag.Vertex, error) {
	return &GraphNodeConfigOutputFlat{
		GraphNodeConfigOutput: n,
		PathValue:             p,
	}, nil
}

// Same as GraphNodeConfigOutput, but for flattening
type GraphNodeConfigOutputFlat struct {
	*GraphNodeConfigOutput

	PathValue []string
}

func (n *GraphNodeConfigOutputFlat) Name() string {
	return fmt.Sprintf(
		"%s.%s", modulePrefixStr(n.PathValue), n.GraphNodeConfigOutput.Name())
}

func (n *GraphNodeConfigOutputFlat) Path() []string {
	return n.PathValue
}

func (n *GraphNodeConfigOutputFlat) DependableName() []string {
	return modulePrefixList(
		n.GraphNodeConfigOutput.DependableName(),
		modulePrefixStr(n.PathValue))
}

func (n *GraphNodeConfigOutputFlat) DependentOn() []string {
	prefix := modulePrefixStr(n.PathValue)
	return modulePrefixList(
		n.GraphNodeConfigOutput.DependentOn(),
		prefix)
}

// GraphNodeDotter impl.
func (n *GraphNodeConfigOutputFlat) DotNode(name string, opts *GraphDotOpts) *dot.Node {
	label := fmt.Sprintf("%s.%s", strings.Join(n.PathValue[1:], "."), n.OutputName())
	return dot.NewNode(name, map[string]string{
		"label":     label,
		"shape":     "invhouse",
		"style":     "filled",
		"fillcolor": "#FFB2B2",
		"fontname":  "courier",
	})
}
