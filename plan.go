package main

import (
	"fmt"

	tfjson "github.com/hashicorp/terraform-json"
)

type ChangeSummary struct {
	Add    int
	Change int
	Remove int
	Import int
}

func (c *ChangeSummary) String() string {
	if c.Import > 0 {
		return fmt.Sprintf("& %d to import, + %d to add, ~ %d to change, - %d to destroy.", c.Import, c.Add, c.Change, c.Remove)
	}
	return fmt.Sprintf("+ %d to add, ~ %d to change, - %d to destroy.", c.Add, c.Change, c.Remove)
}

type Action rune

const (
	NoOp             Action = 0
	Create           Action = '+'
	Read             Action = '←'
	Update           Action = '~'
	DeleteThenCreate Action = '∓'
	CreateThenDelete Action = '±'
	Delete           Action = '-'
)

func UnmarshalActions(actions tfjson.Actions) Action {
	if len(actions) == 2 {
		if actions[0] == "create" && actions[1] == "delete" {
			return CreateThenDelete
		}

		if actions[0] == "delete" && actions[1] == "create" {
			return DeleteThenCreate
		}
	}

	if len(actions) == 1 {
		switch actions[0] {
		case "create":
			return Create
		case "delete":
			return Delete
		case "update":
			return Update
		case "read":
			return Read
		case "no-op":
			return NoOp
		}
	}

	panic("unrecognized action slices")
}

func (a Action) Symbol() string {
	switch a {
	case DeleteThenCreate:
		return "-/+"
	case CreateThenDelete:
		return "+/-"
	case Create:
		return "+"
	case Delete:
		return "-"
	case Read:
		return "<="
	case Update:
		return "~"
	case NoOp:
		return "   "
	default:
		return "  ?"
	}
}
