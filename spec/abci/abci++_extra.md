---
order: 6
title: ABCI++ extra
---
# Introduction

In the section [Tendermint's expected behaviour](./abci++_tmint_expected_behavior.md), we presented the most common behaviour, usually referred to as the good case. However, the grammar specified in the same section is more general and covers more scenarios that an Application designer needs to account for. 

In this section, we give more information about these possible scenarios. We focus on methods introduced by ABCI++: `PrepareProposal` and `ProcessProposal`. Specifically, we concentrate on the part of the grammar presented below.  

```abnf
consensus-height    = *consensus-round decide commit
consensus-round     = proposer / non-proposer

proposer            = [prepare-proposal process-proposal]
non-proposer        = [process-proposal]
```

We can see from the grammar that we can have several rounds before deciding and committing a block. The reasons why one round may not be enough are:
* network asynchrony, and
* Byzantine process being the proposer. 

If we assume that the Tendermint algorithm calls `Decide` and `Commit` for block $X$ in round $r$, in the rounds $r' <= r$, it can exhibit any of the following behaviours:

1. Call `PrepareProposal` and/or `ProcessProposal` for block $X$. 
1. Call `PrepareProposal` and/or `ProcessProposal` for block $Y \neq X$.
1. Does not call `PrepareProposal` and/or `ProcessProposal`.

In the rounds when it is the proposer, Tendermint's `PrepareProposal` call is always followed by the `ProcessProposal` call. The reason is that the process always delivers the proposal to itself, which triggers the `ProcessProposal` call. 

As the number of rounds Tendermint will run is unknown, the Application needs to account for any number of rounds, where each round can exhibit any of these three behaviours. Recall that the Application is unaware of the Tendermint's round. 

# Possible scenarios
The unknown number of rounds we can have in Tendermint yields a vast number of scenarios we can expect. Listing them all is impossible. However, here we give several of them and draw the main conclusions. Specifically, we will show that before `Decide` and `Commit` for block $X$ are called:
    
1. `PrepareProposal` may be called multiple times and for different blocks ([**Scenario 1**](#scenario-1)). 
1. `ProcessProposal` may be called multiple times and for different blocks ([**Scenario 2**](#scenario-2)).
1. `PrepareProposal` and `ProcessProposal` for block $X$ may not be called ([**Scenario 3**](#scenario-3)).
1. `PrepareProposal` and `ProcessProposal` may not be called at all ([**Scenario 4**](#scenario-4)).


## Basic information

Each scenario is presented from the perspective of a process $p$. More precisely, we show what happens in each round's $step$ of the [Tendermint algorithm](https://arxiv.org/pdf/1807.04938.pdf). The legend is below: 

### Round X:

1. **Propose:** Describes what happens while $step_p = propose$.
1. **Prevote:** Describes what happens while $step_p = prevote$.
1. **Precommit:** Describes what happens while $step_p = precommit$.

## Scenario 1

### Round 0:
    
1. **Propose:** Proposer of this round is a Byzantine process, and it chooses not to send the proposal message. Therefore, $p$'s $timeoutPropose$ expires, it sends $Prevote$ for $nil$, and it does not call `ProcessProposal`. All correct processes do the same. 
1. **Prevote:** $p$ eventually receives $2f+1$ $Prevote$ messages for $nil$ and starts $timeoutPrevote$. When $timeoutPrevote$ expires it sends $Precommit$ for $nil$. 
1. **Precommit:** $p$ eventually receives $2f+1$ $Precommit$ messages for $nil$ and starts $timeoutPrecommit$. When it expires, it moves to the next round. 
    
### Round 1: 

1. **Propose:** Correct process is the proposer in this round. Its $validValue$ is $nil$, and it is free to generate and propose new block $Y$. Process $p$ receives this proposal in time, calls `ProcessProposal` for block $Y$, and broadcasts a $Prevote$ message for it. 
1. **Prevote:** Due to network asynchrony less than $2f+1$ processes send $Prevote$ for this block. Therefore, $p$ does not update $validValue$ in this round. 
1. **Precommit:** Since less than $2f+1$ processes send $Prevote$, no correct process will lock on this block and send $Precommit$ message. As a consequence, $p$ does not decide on $Y$. 

### Round 2: 

1. **Propose:** Same as in [**Round 1**](#round-1), just another correct process is the proposer, and it proposes another value $Z$. Process $p$ receives the proposal in time, calls `ProcessProposal` for new block $Z$, and broadcasts a $Prevote$ message for it. 
1. **Prevote:** Same as in [**Round 1**](#round-1).
1. **Precommit:** Same as in [**Round 1**](#round-1).


Rounds like these can continue until we have a round in which process $p$ updates its $validValue$ or until we reach round $r$ where process $p$ decides on a block. After that, it will not call `ProcessProposal` any more for this height. 

## Scenario 2 

### Round 0: 

1. **Propose:** Process $p$ is the proposer in this round. Its $validValue$ is $nil$, and it is free to generate and propose new block $Y$. Before proposing, it calls `PrepareProposal` for $Y$. After that, it broadcasts the proposal, delivers it to itself, calls `ProcessProposal` and broadcasts $Prevote$ for it. 
1. **Prevote:** Due to network asynchrony less than $2f+1$ processes receives the proposal in time and send $Prevote$ for it. Therefore, $p$ does not update $validValue$ in this round. 
1. **Precommit:** Since less than $2f+1$ processes send $Prevote$, no correct process will lock on this block and send $Precommit$ message. As a consequence, $p$ does not decide on $Y$. 

After this round, we can have multiple rounds like those in [Scenario 1](#scenario-1). The important thing is that process $p$ should not update its $validValue$. Consequently, when process $p$ reaches the round when it is again the proposer, it will ask the mempool for the new block again, and the mempool may return a different block $Z$, and we can have the same round as [Round 0](#round-0-1) just for a different block. As a result, process $p$ calls `PrepareProposal` again but for a different value. When it reaches round $r$ 
some process will propose block $X$ and if $p$ receives $2f+1$ $Precommit$ messages, it will decide on this value. 


## Scenario 3 

In this scenario, in all rounds before $r$ we can have any round presented in [Scenario 1](#scenario-1) or [Scenario 2](#scenario-2). What is important is that:
- no proposer proposed block $X$ or if it did, process $p$, due to asynchrony, did not receive it in time, so it did not call `ProcessProposal`, and
- if $p$ was the proposer it proposed some other value $\neq X$. 

### Round $r$: 

1. **Propose:** Correct process is the proposer in this round, and it proposes block $X$. 
Due to asynchrony, the proposal message arrives to process $p$ after its $timeoutPropose$
expires and it sends $Prevote$ for $nil$. Consequently, process $p$ does not call
`ProcessProposal` for block $X$. However, the same proposal arrives at other processes
before their $timeoutPropose$ expires, and they send $Prevote$ for this proposal.
1. **Prevote:** Process $p$ receives $2f+1$ $Prevote$ messages for proposal $X$, updates correspondingly its $validValue$ and $lockedValue$ and sends $Precommit$ message. All correct processes do the same. 
1. **Precommit:** Finally, process $p$ receives $2f+1$ $Precommit$ messages, and decides on block $X$. As a result, it calls `Decide` and `Commit` for block $X$.



## Scenario 4

[Scenario 3](#scenario-3) can be translated into a scenario where $p$ does not call `PrepareProposal` and `ProcessProposal` at all. For this, it is necessary that process $p$ is not the proposer in any of the rounds $0 <= r' <= r$ and that due to network asynchrony or Byzantine proposer, it does not receive the proposal before $timeoutPropose$ expires. As a result, it will enter round $r$ without calling `PrepareProposal` and `ProcessProposal` before it, and as shown in [Round $r$ of Scenario 3](#scenario-3) it will decide in this round. Again without calling any of these two calls.  






 




