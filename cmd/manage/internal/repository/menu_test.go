package repository

import (
	"testing"

	"example.com/go-service-kit/cmd/manage/internal/model"
)

func TestValidateMenuParent(t *testing.T) {
	rows := map[int]model.Menu{
		1: {Id: 1, Type: model.MENU_TYPE_MENU},
		2: {Id: 2, Type: model.MENU_TYPE_MENU, Pid: 1},
		3: {Id: 3, Type: model.MENU_TYPE_NODE, Pid: 2},
	}
	for _, test := range []struct {
		name   string
		menu   model.Menu
		wantOK bool
	}{
		{"root", model.Menu{Id: 1}, true},
		{"valid parent", model.Menu{Id: 3, Pid: 2}, true},
		{"self", model.Menu{Id: 1, Pid: 1}, false},
		{"descendant", model.Menu{Id: 1, Pid: 2}, false},
		{"permission node", model.Menu{Id: 1, Pid: 3}, false},
		{"missing parent", model.Menu{Id: 1, Pid: 4}, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			if err := validateMenuParent(rows, &test.menu); (err == nil) != test.wantOK {
				t.Fatalf("validateMenuParent returned %v, want success %t", err, test.wantOK)
			}
		})
	}
}
