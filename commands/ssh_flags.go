package commands

import "os"

// applySshFlagsToEnv promotes the apply/plan SSH-related CLI flags
// (--host, --sudo, --accept-new-host-keys) to the env vars that
// subprocess/ssh.go reads at dispatch time. Precedence is CLI flag >
// pre-existing env var > default. Once set here, the env value is the
// single source of truth for the rest of the run.
//
// We do not Unsetenv on exit: docket commands run as one-shot processes
// and the env mutation does not escape. If the package is later imported
// as a library, revisit.
func applySshFlagsToEnv(host string, sudo, acceptNewHostKeys bool) {
	if host != "" {
		_ = os.Setenv("DOKKU_HOST", host)
	}
	if sudo {
		_ = os.Setenv("DOKKU_SUDO", "1")
	}
	if acceptNewHostKeys {
		_ = os.Setenv("DOKKU_SSH_ACCEPT_NEW_HOST_KEYS", "1")
	}
}
