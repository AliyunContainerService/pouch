package networkdb

import "github.com/hashicorp/memberlist"

type eventDelegate struct {
	nDB *NetworkDB
}

func (e *eventDelegate) NotifyJoin(mn *memberlist.Node) {
	e.nDB.Lock()
	// In case the node is rejoining after a failure or leave,
	// wait until an explicit join message arrives before adding
	// it to the nodes just to make sure this is not a stale
	// join. If you don't know about this node add it immediately.
	_, fOk := e.nDB.failedNodes[mn.Name]
	_, lOk := e.nDB.leftNodes[mn.Name]
	if fOk || lOk {
		e.nDB.Unlock()
		return
	}

	e.nDB.nodes[mn.Name] = &node{Node: *mn}
	e.nDB.Unlock()
}

func (e *eventDelegate) NotifyLeave(mn *memberlist.Node) {
	e.nDB.deleteNodeTableEntries(mn.Name)
	e.nDB.deleteNetworkEntriesForNode(mn.Name)
	e.nDB.Lock()
	if n, ok := e.nDB.nodes[mn.Name]; ok {
		delete(e.nDB.nodes, mn.Name)
		e.nDB.failedNodes[mn.Name] = n
	}
	e.nDB.Unlock()
}

func (e *eventDelegate) NotifyUpdate(n *memberlist.Node) {
}
