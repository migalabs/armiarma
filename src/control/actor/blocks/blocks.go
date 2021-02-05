package blocks

import (
	"errors"
	"github.com/protolambda/ask"
	bdb "github.com/protolambda/rumor/chain/db/blocks"
	"github.com/protolambda/rumor/control/actor/base"
	"github.com/protolambda/rumor/control/actor/blocks/dbcmd"
)

type DBState struct {
	CurrentDB bdb.DBID
}

type BlocksCmd struct {
	*base.Base
	bdb.DBs
	*DBState
}

func (c *BlocksCmd) Cmd(route string) (cmd interface{}, err error) {
	switch route {
	case "create":
		cmd = &CreateCmd{Base: c.Base, DBs: c.DBs, DBState: c.DBState}
	case "copy":
		cmd = &CopyCmd{Base: c.Base}
	case "switch":
		cmd = &SwitchCmd{Base: c.Base, DBState: c.DBState}
	case "rm":
		cmd = &RemoveCmd{Base: c.Base, DBs: c.DBs}
	case "list":
		cmd = &ListCmd{Base: c.Base, DBs: c.DBs, DBState: c.DBState}
	case "db":
		db, ok := c.DBs.Find(c.CurrentDB)
		if !ok {
			return nil, errors.New("current DB not available. Create one with 'blocks create'")
		}
		cmd = &dbcmd.DBCmd{Base: c.Base, DB: db}
	case "on":
		cmd = &OnCmd{Base: c.Base, DBs: c.DBs}
	default:
		return nil, ask.UnrecognizedErr
	}
	return cmd, nil
}

func (c *BlocksCmd) Routes() []string {
	return []string{"create", "copy", "switch", "rm", "list", "db", "on"}
}

func (c *BlocksCmd) Help() string {
	return "Manage and interact with blocks DBs"
}
