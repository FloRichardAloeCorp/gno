package present

import (
	"std"
	"strings"

	"gno.land/p/demo/avl"
)

var (
	adminAddr     std.Address
	moderatorList avl.Tree
	inPause       bool
)

func init() {
	// adminAddr = std.GetOrigCaller() // FIXME: find a way to use this from the main's genesis.
	adminAddr = "g1u7y667z64x2h7vc6fmpcprgey4ck233jaww9zq"
}

func AdminSetAdminAddr(addr std.Address) {
	assertIsAdmin()
	adminAddr = addr
}

func AdminSetInPause(state bool) {
	assertIsAdmin()
	inPause = state
}

func AdminAddModerator(addr std.Address) {
	assertIsAdmin()
	moderatorList.Set(addr.String(), true)
}

func AdminRemoveModerator(addr std.Address) {
	assertIsAdmin()
	moderatorList.Set(addr.String(), false) // XXX: delete instead?
}

func ModAddPost(slug, title, body, tags string) {
	assertIsModerator()

	caller := std.GetOrigCaller()
	tagList := strings.Split(tags, ",")
	err := b.NewPost(caller, slug, title, body, tagList)
	checkErr(err)
}

func ModEditPost(slug, title, body, tags string) {
	assertIsModerator()

	tagList := strings.Split(tags, ",")
	err := b.GetPost(slug).Update(title, body, tagList)
	checkErr(err)
}

func isAdmin(addr std.Address) bool {
	return addr == adminAddr
}

func isModerator(addr std.Address) bool {
	_, found := moderatorList.Get(addr.String())
	return found
}

func assertIsAdmin() {
	caller := std.GetOrigCaller()
	if !isAdmin(caller) {
		panic("access restricted.")
	}
}

func assertIsModerator() {
	caller := std.GetOrigCaller()
	if isAdmin(caller) || isModerator(caller) {
		return
	}
	panic("access restricted")
}

func assertNotInPause() {
	if inPause {
		panic("access restricted (pause)")
	}
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
