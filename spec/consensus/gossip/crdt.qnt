// -*- mode: Bluespec; -*-

module CRDT {
    import Spells.* from "./spells"
    import Globals.* from "./globals"
    import OInt as OInt from "./option"
    import OStr as OStr from "./option"

    type Value = str //TODO: sum type

    //TODO: how to use Nat as a type?
    type NatNum = int

    type MsgType =  | {tag: "Proposal"}
                    | {tag: "PreVote"}
                    | {tag: "PreCommit"}

    type Entry = {
        height: NatNum,
        round: NatNum,
        msgType: MsgType,
        validator: ProcId,
        value: Value
    }

    //An entry with optional fields, which correspond to * in a query, if not set.
    type EntryO = {
        height: OInt::OType,
        round: OInt::OType,
        msgType: OStr::OType,
        validator: OInt::OType,
        value: OStr::OType
    }

    pure def makeEntry(height: NatNum, round: NatNum, msgType: str, validator: ProcId, value: Value): Entry =
        {height: height, round: round, msgType: {tag:msgType}, validator: validator, value: value}

    type View = Set[Entry]
    pure val bot:View = Set()

    /*
    Some considerations wrt to supersession.
    - Latest decided hight
        - If h is the last decided height, then anything from h' < h is stale.
        - Keeping information from a few decided heights might be useful for helping nodes catch up.
        - This is an optimization.
        - How to determine the last decided height (h) and the current height (h+1)
        - How to prevent bizantine nodes from sending messages from higher heights, in a way compatible with CRDT?
    - Latest round
        - How to determine the current round based on the CRDT?
        - How to prevent bizantine nodes from sending messages from higher rounds, in a way compatible with CRDT?
    - Merging any subset of the state does not compromise correctness
        - Allows for optimizations.
        - As long as all messages are possibly merged, eventual strong consistency is ensured.
    - Ordering of gossip heuristics
        - Are they just heuristics or have they been proved correct anywhere or at least discussed?
    */
    pure def isSupersededBy(v1: View, v2: View): bool = 
    {
        val currentHeight
        //Ignore votes from higher heights until you are ready for them.
        
        val currentRound
        //Ignore votes from higher rounds until you are ready for them.

        //ignore pre-votes if you have seen pre-commits.
        

    }

    pure def matches(e: Entry, oe:EntryO):bool = all {
        OInt::nsoe(oe.height, e.height),
        OInt::nsoe(oe.round, e.round),
        OStr::nsoe(oe.msgType, e.msgType.tag),
        OInt::nsoe(oe.validator, e.validator),
        OStr::nsoe(oe.value, e.value),
    }

    pure def removeStale(view: View): (View,View) =
        val superseded = view.filter(e => Set(e).isSupersededBy(view.excludeE(e)))
        val alive = view.exclude(superseded)
        (alive, superseded)

    pure def merge(lhs: View, rhs: View): View =
        val newViewAndStale = lhs.union(rhs).removeStale()
        newViewAndStale._1

    // Adds an entry to the view.
    pure def addEntry(v:View, e: Entry): View =
        val newViewAndStale = v.unionE(e).removeStale()
        newViewAndStale._1

    pure def query(view, oe) =
        view.filter(e => e.matches(oe))

    pure def hasEntry(v: View, e:Entry):bool = v.contains(e)

    //Unit tests
    run makeEntryTest = all {
        val entry1 = {
            height: 1, 
            round: 1, 
            msgType: {tag:"Proposal"}, 
            validator: 1, 
            value: ""
        }
        assert(makeEntry(1, 1, "Proposal", 1, "") == entry1)
    }

    run matchesTest = all {
        val entry1 = makeEntry(1,1,"Proposal",1,"")
        all {
            assert(entry1
                .matches({
                    height: OInt::none, 
                    round: OInt::none,
                    msgType: OStr::none,
                    validator: OInt::none,
                    value: OStr::none
                })),
            assert(entry1
                .matches({
                    height: OInt::valueOf(1),
                    round: OInt::valueOf(1),
                    msgType: OStr::valueOf("Proposal"),
                    validator: OInt::valueOf(1),
                    value: OStr::valueOf("")
                })),
            assert(entry1
                .matches({
                    height: OInt::none, 
                    round: OInt::valueOf(1),
                    msgType: OStr::valueOf("Proposal"),
                    validator: OInt::valueOf(1),
                    value: OStr::valueOf("")
                })),
            assert(entry1
                .matches({
                    height: OInt::none, 
                    round: OInt::valueOf(1),
                    msgType: OStr::none,
                    validator: OInt::valueOf(1),
                    value: OStr::valueOf("")
                })),
        }
    }

    run isSupersededByTest = all {
        //Higher hight
        assert(Set(makeEntry(1,1,"Proposal",1,"")).isSupersededBy(Set(makeEntry(2,1,"Proposal",1,"")))),
        assert(Set(makeEntry(1,1,"Proposal",1,"")).isSupersededBy(Set(makeEntry(2,2,"Proposal",2,"")))),

        //Different validator from the same hight
        assert(not(Set(makeEntry(1,1,"Proposal",1,"")).isSupersededBy(Set(makeEntry(1,1,"Proposal",2,""))))),
    }

    run removeStaleTest = all {
        false,
    }

    run mergeTest = all {
        false,
    }

    run addEntryTest = all {
        assert(bot.addEntry(makeEntry(1, 1, "Proposal", 1, "")) == Set(makeEntry(1, 1, "Proposal", 1, ""))),
        assert(bot.addEntry(makeEntry(1, 1, "Proposal", 1, ""))
                  .addEntry(makeEntry(1, 1, "Proposal", 2, "")) == Set(makeEntry(1, 1, "Proposal", 1, ""),
                                                                       makeEntry(1, 1, "Proposal", 2, ""))),
        assert(bot.addEntry(makeEntry(1, 1, "Proposal", 1, ""))
                  .addEntry(makeEntry(2, 1, "Proposal", 2, "")) == Set(makeEntry(2, 1, "Proposal", 2, ""))),
    }

    run hasEntryTest = all {
        assert(bot.addEntry(makeEntry(1, 1, "Proposal", 1, "")).hasEntry(makeEntry(1, 1, "Proposal", 1, ""))),
        assert(bot.addEntry(makeEntry(1, 1, "Proposal", 1, ""))
                  .addEntry(makeEntry(1, 1, "Proposal", 2, ""))
                  .hasEntry(makeEntry(1, 1, "Proposal", 1, ""))),
        assert(bot.addEntry(makeEntry(1, 1, "Proposal", 1, ""))
                  .addEntry(makeEntry(1, 1, "Proposal", 2, ""))
                  .hasEntry(makeEntry(1, 1, "Proposal", 2, ""))),
        assert(not(bot.addEntry(makeEntry(1, 1, "Proposal", 1, "")).hasEntry(makeEntry(1, 1, "Proposal", 2, "")))),
    }

    run queryTest = all {
        val view = Set(
                    makeEntry(1, 1, "Proposal", 1, ""), 
                    makeEntry(1, 1, "Proposal", 2, ""), 
                    makeEntry(2, 1, "Proposal", 2, ""))
        val view1nn = Set(
                    makeEntry(1, 1, "Proposal", 1, ""), 
                    makeEntry(1, 1, "Proposal", 2, ""))
        val viewn1n = view
        val viewnn2 = Set(
                    makeEntry(1, 1, "Proposal", 2, ""), 
                    makeEntry(2, 1, "Proposal", 2, ""))
        all {
            assert(view.query({
                    height: OInt::valueOf(1),
                    round: OInt::none, 
                    validator: OInt::none,
                    msgType: OStr::none,
                    value: OStr::valueOf("")
                }) == view1nn),
            assert(view.query({
                    height: OInt::none, 
                    round: OInt::valueOf(1),
                    validator: OInt::none,
                    msgType: OStr::none,
                    value: OStr::valueOf("")
                }) == viewn1n),
            assert(view.query({
                    height: OInt::none, 
                    round: OInt::none,
                    validator: OInt::valueOf(2),
                    msgType: OStr::none,
                    value: OStr::valueOf("")
                }) == viewnn2)
        }
    }
}