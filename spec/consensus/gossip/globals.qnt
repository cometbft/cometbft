// Global parameters.
// Must be imported as `Globals.*`
module Globals {
    type ProcId = int

    //XXX Quint: Ideally this should be a const, but because it would have to be
    //instantiated on every import, there would be duplication.
    //All processes in the system
    pure val ProcIds = Set(1,2,3,4,5)

    pure val DEFAULT_MAX_INBOUND_CONNECTION_NUMBER = 4
    pure val DEFAULT_MAX_OUTBOUND_CONNECTION_NUMBER = 4
}
