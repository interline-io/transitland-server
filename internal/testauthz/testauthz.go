package testauthz

import (
	"context"
	"testing"

	"github.com/interline-io/transitland-server/internal/generated/azpb"
)

type TupleKey = azpb.TupleKey

type FGAProvider interface {
	WriteTuple(context.Context, TupleKey) error
	CreateStore(context.Context, string) (string, error)
	CreateModel(context.Context, string) (string, error)
}

func LoadTestTuples(t testing.TB, fgaClient FGAProvider, modelFile string, testTuples []TestTuple) error {
	if _, err := fgaClient.CreateStore(context.Background(), "test"); err != nil {
		t.Fatal(err)
	}
	if _, err := fgaClient.CreateModel(context.Background(), modelFile); err != nil {
		t.Fatal(err)
	}
	for _, tk := range testTuples {
		if err := fgaClient.WriteTuple(context.Background(), tk.TupleKey()); err != nil {
			t.Fatal(err)
		}
	}
	return nil
}

type TestTuple struct {
	Subject            azpb.EntityKey
	Object             azpb.EntityKey
	Action             azpb.Action
	Relation           azpb.Relation
	Expect             string
	Notes              string
	ExpectError        bool
	ExpectUnauthorized bool
	CheckAsUser        string
	ExpectActions      []azpb.Action
	ExpectKeys         []azpb.EntityKey
}

func (tk *TestTuple) TupleKey() TupleKey {
	return TupleKey{Subject: tk.Subject, Object: tk.Object, Relation: tk.Relation, Action: tk.Action}
}

func (tk *TestTuple) String() string {
	a := tk.TupleKey().String()
	if tk.CheckAsUser != "" {
		a = a + "|checkuser:" + tk.CheckAsUser
	}
	return a
}
