package articlemd

import (
	"io"

	markdown "github.com/teekennedy/goldmark-markdown"
	"github.com/yuin/goldmark"
	goldmarkAst "github.com/yuin/goldmark/ast"
	goldmarkExt "github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	goldmarkText "github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

func TransformDocument(w io.Writer, content []byte, walker func(n goldmarkAst.Node, entering bool) error) error {
	renderer := goldmark.New()
	docAst := renderer.Parser().Parse(goldmarkText.NewReader(content))

	err := goldmarkAst.Walk(docAst, func(n goldmarkAst.Node, entering bool) (goldmarkAst.WalkStatus, error) {
		if !entering {
			return goldmarkAst.WalkContinue, nil
		}

		err := walker(n, entering)
		if err != nil {
			return goldmarkAst.WalkStop, err
		}

		return goldmarkAst.WalkContinue, nil
	})

	if err != nil {
		return err
	}

	mdRenderer := goldmark.New(goldmark.WithRenderer(markdown.NewRenderer()))
	return mdRenderer.Renderer().Render(w, content, docAst)
}

func ParseDocument(content []byte, w io.Writer) error {
	exts := goldmark.WithExtensions(goldmarkExt.Table, goldmarkExt.GFM, goldmarkExt.TaskList)
	renderers := renderer.WithNodeRenderers(
		util.Prioritized(html.NewRenderer(), 200),
		util.Prioritized(&LazyImageRenderer{}, 100),
	)

	renderer := goldmark.New(
		exts,
		goldmark.WithRenderer(renderer.NewRenderer(html.WithUnsafe(), renderers)),
	)
	doc := renderer.Parser().Parse(goldmarkText.NewReader(content))

	return renderer.Renderer().Render(w, content, doc)
}
