package authn

import (
	"os"

	"github.com/interline-io/transitland-lib/tlcsv"
	openfga "github.com/openfga/go-sdk"
)

func LoadTuples(fn string) ([]TupleKey, error) {
	tkeys := []TupleKey{}
	if f, err := os.Open(fn); err != nil {
		return nil, err
	} else {
		tlcsv.ReadRows(f, func(row tlcsv.Row) {
			tk := TupleKey{
				User:     row.GetString("user"),
				Relation: row.GetString("relation"),
				Object:   row.GetString("object"),
			}
			if row.GetString("assert") == "true" {
				tk.Assert = true
			}
			tkeys = append(tkeys, tk)
		})
	}
	return tkeys, nil
}

type TupleKey struct {
	User     string
	Relation string
	Object   string
	Assert   bool
}

func (tk TupleKey) FGATupleKey() openfga.TupleKey {
	return openfga.TupleKey{
		User:     openfga.PtrString(tk.User),
		Relation: openfga.PtrString(tk.Relation),
		Object:   openfga.PtrString(tk.Object),
	}
}