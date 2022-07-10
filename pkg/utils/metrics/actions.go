package metrics

// Action labels for different pod/container/image operations.
const (
	ActionCreateLabel    = "create"
	ActionDeleteLabel    = "delete"
	ActionRemoveLabel    = "remove"
	ActionUpdateLabel    = "update"
	ActionUpgradeLabel   = "upgrade"
	ActionInfoLabel      = "info"
	ActionListLabel      = "list"
	ActionStatusLabel    = "status"
	ActionStartLabel     = "start"
	ActionStopLabel      = "stop"
	ActionKillLabel      = "kill"
	ActionRenameLabel    = "rename"
	ActionRestartLabel   = "restart"
	ActionRunLabel       = "run"
	ActionPullLabel      = "pull"
	ActionStatsLabel     = "stats"
	ActionStatsListLabel = "stats_list"
	ActionPauseLabel     = "pause"
	ActionUnpauseLabel   = "unpause"
)
