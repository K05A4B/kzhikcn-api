package articlemd

import (
	"bytes"
	"fmt"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

type LazyImageRenderer struct{}

func (r *LazyImageRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.KindImage, r.renderImage)
}

func (r *LazyImageRenderer) renderImage(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		img := node.(*ast.Image)
		_, _ = w.WriteString(`<img `)

		fmt.Fprintf(w, "data-src=%q alt=%q", img.Destination, img.Title)

		attrs := node.Attributes()
		for _, attr := range attrs {
			if !bytes.HasPrefix(attr.Name, []byte("hAttr-")) {
				continue
			}

			attrName := bytes.TrimPrefix(attr.Name, []byte("hAttr-"))
			fmt.Fprintf(w, "%s=%q", attrName, attr.Value)
		}

		w.WriteString("/>")

		return ast.WalkSkipChildren, nil
	}

	return ast.WalkContinue, nil
}
