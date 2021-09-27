package testdata

import (
	"context"

	foobar "github.com/KazanExpress/tracegen/tracegen/testdata/bar"
	"github.com/KazanExpress/tracegen/tracegen/testdata/foo"
)

type Full interface {
	// A does something important.
	// @trace Int64:id, String:text
	A(ctx context.Context, id int64, text string) (bool, error)

	// B does something important too.
	// @trace foo
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

	// H does something really important.
	// @trace foos
	H(ctx context.Context, foos ...foo.Foo) error

	// I does something strange.
	// @trace foos
	I(ctx context.Context, foos ...[]foo.Foo) error

	// J does something really strange.
	// @trace foos
	J(ctx context.Context, foos ...*[]*foo.Foo) []*[]*foo.Foo
}
