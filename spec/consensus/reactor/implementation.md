# Current Implementation

## [REQ-CONS-GOSSIP-KEEP_NON_SUPERSEDED]
GOSSIP reacts to messages by adding them to sets of similar messages, within the GOSSIP internal state, and then evaluating if Tendermint conditions are met and triggering changes to CONS, or by itself reacting to implement the gossip communication.

Starting a new height produces a message that supersedes all previous messages and allows the sets to be reset and memory contained.

Starting a new round of the same height has a similar but weaker effect.


## [DEF-SUPERSESSION]
Currently the knowledge of message supersession is embedded in GOSSIP, which decides which messages to retransmit based on the CONS' state and the GOSSIP's state.
 
Even though there is no specific superseding operator implemented, superseding happens by advancing steps, rounds and heights.

> @josef-wider
> In the past we looked a bit into communication closure w.r.t. consensus. Roughly, the lexicographical order over the tuples (height, round, step) defines a notion of logical time, and when I am in a certain height, round and step, I don't care about messages from "the past". Tendermint consensus is mostly communication-closed in that we don't care about messages from the past. An exception is line 28 in the arXiv paper where we accept prevote messages from previous rounds vr for the same height.
> 
> I guess a precise constructive definition of "supersession" can be done along these lines.