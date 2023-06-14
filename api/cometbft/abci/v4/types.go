package v4

// IsAccepted returns true if Code is ACCEPT
func (r ResponseProcessProposal) IsAccepted() bool {
	return r.Status == PROCESS_PROPOSAL_STATUS_ACCEPT
}

// IsStatusUnknown returns true if Code is UNKNOWN
func (r ResponseProcessProposal) IsStatusUnknown() bool {
	return r.Status == PROCESS_PROPOSAL_STATUS_UNKNOWN
}

func (r ResponseVerifyVoteExtension) IsAccepted() bool {
	return r.Status == VERIFY_VOTE_EXTENSION_STATUS_ACCEPT
}

// IsStatusUnknown returns true if Code is Unknown
func (r ResponseVerifyVoteExtension) IsStatusUnknown() bool {
	return r.Status == VERIFY_VOTE_EXTENSION_STATUS_UNKNOWN
}
