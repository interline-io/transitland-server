package authz

import (
	"fmt"
	"os"
	"strings"

	"github.com/interline-io/transitland-lib/tlcsv"
	openfga "github.com/openfga/go-sdk"
)

type AuthzConfig struct {
	Auth0Domain       string
	Auth0ClientID     string
	Auth0ClientSecret string
	FGAStoreID        string
	FGAModelID        string
	FGAEndpoint       string
	FGATestModelPath  string
	FGATestTuplesPath string
}

type User struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email,omitempty"`
}

type TupleKey struct {
	UserType   string `json:"user_type,omitempty"`
	UserName   string `json:"user_name,omitempty"`
	Relation   string `json:"relation,omitempty"`
	ObjectType string `json:"object_type,omitempty"`
	ObjectName string `json:"object_name,omitempty"`
	Assert     bool   `json:"assert,omitempty"`
}

func LoadTuples(fn string) ([]TupleKey, error) {
	tkeys := []TupleKey{}
	if f, err := os.Open(fn); err != nil {
		return nil, err
	} else {
		tlcsv.ReadRows(f, func(row tlcsv.Row) {
			tk := TupleKey{}
			tk.UserType = csplit(row.GetString("user"))[0]
			tk.UserName = csplit(row.GetString("user"))[1]
			tk.ObjectType = csplit(row.GetString("object"))[0]
			tk.ObjectName = csplit(row.GetString("object"))[1]
			tk.Relation = row.GetString("relation")
			if row.GetString("assert") == "true" {
				tk.Assert = true
			}
			tkeys = append(tkeys, tk)
		})
	}
	return tkeys, nil
}

func (tk TupleKey) WithUser(user string) TupleKey {
	return TupleKey{
		UserType:   "user",
		UserName:   user,
		ObjectType: tk.ObjectType,
		ObjectName: tk.ObjectName,
		Relation:   tk.Relation,
	}
}

func (tk TupleKey) WithObject(objectType string, objectName string) TupleKey {
	return TupleKey{
		UserType:   tk.UserType,
		UserName:   tk.UserName,
		ObjectType: objectType,
		ObjectName: objectName,
		Relation:   tk.Relation,
	}

}

func (tk TupleKey) WithRelation(relation string) TupleKey {
	return TupleKey{
		UserType:   tk.UserType,
		UserName:   tk.UserName,
		ObjectType: tk.ObjectType,
		ObjectName: tk.ObjectName,
		Relation:   relation,
	}
}

func fromFGATupleKey(fgatk openfga.TupleKey) TupleKey {
	okeys := csplit(*fgatk.Object)
	ukeys := csplit(*fgatk.User)
	return TupleKey{
		UserType:   ukeys[0],
		UserName:   ukeys[1],
		ObjectType: okeys[0],
		ObjectName: okeys[1],
		Relation:   *fgatk.Relation,
	}
}

func (tk TupleKey) FGATupleKey() openfga.TupleKey {
	return openfga.TupleKey{
		User:     openfga.PtrString(cunsplit(tk.UserType, tk.UserName)),
		Relation: openfga.PtrString(tk.Relation),
		Object:   openfga.PtrString(cunsplit(tk.ObjectType, tk.ObjectName)),
	}
}

func csplit(v string) [2]string {
	a := strings.Split(v, ":")
	ret := [2]string{"", ""}
	if len(a) > 1 {
		ret[0] = a[0]
		ret[1] = a[1]
	} else if len(a) > 0 {
		ret[0] = a[0]
	}
	return ret
}

func cunsplit(a, b string) string {
	if b == "" {
		return a
	}
	return fmt.Sprintf("%s:%s", a, b)
}
