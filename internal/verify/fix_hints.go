package verify

// Fix hint constants — copy-paste commands shown under failed checks.
// One constant per check ID family. Keep them short (one command, no prose).
const (
	FixInstallEngram     = "architect-ai install --component engram"
	FixInstallContext7   = "architect-ai install --component context7"
	FixInstallNotebookLM = "architect-ai install --component notebooklm"
	FixInstallGGA        = "architect-ai install --component gga"
	FixSync              = "architect-ai sync"
	FixRepairPermissions = "architect-ai sync --repair-permissions"
	FixInstallNode       = "Install Node.js 18+ → https://nodejs.org"
	FixBuild             = "architect-ai build"
	FixMigrateV03        = "architect-ai migrate-v03"
	FixDiagnoseEngram    = "architect-ai diagnose engram"
)
