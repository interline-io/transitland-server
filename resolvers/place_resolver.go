package resolvers

import (
	"context"
	"encoding/json"
	"fmt"

	dataloader "github.com/graph-gophers/dataloader/v7"
	"github.com/interline-io/transitland-server/model"
)

type placeResolver struct{ *Resolver }

func (r *placeResolver) Operators(ctx context.Context, obj *model.Place) ([]*model.Operator, error) {
	var ret []*model.Operator
	var thunks []dataloader.Thunk[*model.Operator]
	for _, oid := range obj.OperatorOnestopIDs {
		fmt.Println("creating thunk for operator:", oid)
		t := For(ctx).OperatorsByOnestopID.Load(ctx, oid.Val)
		thunks = append(thunks, t)
	}
	for i := 0; i < len(obj.OperatorOnestopIDs); i++ {
		oid := obj.OperatorOnestopIDs[i].Val
		o, err := thunks[i]()
		if err != nil {
			panic(err)
		}
		if o != nil {
			oj, _ := json.Marshal(o)
			fmt.Println("got operator for:", oid, "json:", string(oj))
			ret = append(ret, o)
		} else {
			fmt.Println("no operator for:", oid)
		}
	}
	return ret, nil
}

func (r *placeResolver) Count(ctx context.Context, obj *model.Place) (int, error) {
	return len(obj.OperatorOnestopIDs), nil
}
