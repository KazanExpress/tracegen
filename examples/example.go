package tracegen

import (
	"context"

	foobar "github.com/KazanExpress/tracegen/examples/bar"
	"github.com/KazanExpress/tracegen/examples/foo"
)

//go:generate tracegen -source $GOFILE -destination "example_gen.go"

type Example interface {
	// A does something important.
	// @trace Int64:id, String:text
	A(ctx context.Context, id int64, text string) (bool, error)

	// B does something important too.
	// @trace foo, String:foo.Name
	B(ctx context.Context, foo *foo.Foo) foobar.Bar

	// C does something important?
	// @trace String:foo.Name
	C(ctx context.Context, foo *foo.Foo) error

	D(ctx context.Context)

	// E does something important.
	// @trace bar
	E(ctx context.Context, bar []foobar.Bar) []foobar.Bar

	// F does something important.
	// @trace bars
	F(ctx context.Context, bars []*foobar.Bar) []*foobar.Bar

	// G does something important.
	// @trace bars
	G(ctx context.Context, bars *[]*foobar.Bar) *[]*foobar.Bar
}
