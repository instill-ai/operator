package json

import (
	"fmt"
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/gofrs/uuid"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/instill-ai/x/errmsg"
)

const asJSON = `{"foo":"bar"}`

var asMap = map[string]any{"foo": "bar"}

func TestOperator_Execute(t *testing.T) {
	c := qt.New(t)

	testcases := []struct {
		name string

		task    string
		in      map[string]any
		want    map[string]any
		wantErr string
	}{
		{
			name: "ok - marshal",

			task: taskMarshal,
			in:   map[string]any{"json": asMap},
			want: map[string]any{"string": asJSON},
		},
		{
			name: "nok - marshal",

			task:    taskMarshal,
			in:      map[string]any{},
			wantErr: "Couldn't convert the provided object to JSON.",
		},
		{
			name: "ok - unmarshal",

			task: taskUnmarshal,
			in:   map[string]any{"string": asJSON},
			want: map[string]any{"json": asMap},
		},
		{
			name: "nok - unmarshal",

			task:    taskUnmarshal,
			in:      map[string]any{"string": `{`},
			wantErr: "Couldn't parse the JSON string. Please check the syntax is correct.",
		},
	}

	logger := zap.NewNop()
	operator := Init(logger)
	defID := uuid.Must(uuid.NewV4())
	config := &structpb.Struct{}

	for _, tc := range testcases {
		c.Run(tc.name, func(c *qt.C) {
			exec, err := operator.CreateExecution(defID, tc.task, config, logger)
			c.Assert(err, qt.IsNil)

			pbIn, err := structpb.NewStruct(tc.in)
			c.Assert(err, qt.IsNil)

			got, err := exec.Execute([]*structpb.Struct{pbIn})
			if tc.wantErr != "" {
				c.Check(errmsg.Message(err), qt.Matches, tc.wantErr)
				return
			}

			c.Check(err, qt.IsNil)
			c.Assert(got, qt.HasLen, 1)

			gotJSON, err := got[0].MarshalJSON()
			c.Assert(err, qt.IsNil)
			c.Check(gotJSON, qt.JSONEquals, tc.want)
		})
	}
}

func TestOperator_CreateExecution(t *testing.T) {
	c := qt.New(t)

	logger := zap.NewNop()
	operator := Init(logger)
	defID := uuid.Must(uuid.NewV4())

	c.Run("nok - unsupported task", func(c *qt.C) {
		task := "FOOBAR"
		want := fmt.Sprintf("%s task is not supported.", task)

		_, err := operator.CreateExecution(defID, task, new(structpb.Struct), logger)
		c.Check(err, qt.IsNotNil)
		c.Check(errmsg.Message(err), qt.Equals, want)
	})
}
