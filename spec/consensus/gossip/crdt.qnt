// -*- mode: Bluespec; -*-

module CRDT {
    import Spells.* from "./spells"
    import Globals.* from "./globals"

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

    pure def makeEntry(height: NatNum, round: NatNum, msgType: str, validator: ProcId, value: Value): Entry =
        {height: height, round: round, msgType: {tag:msgType}, validator: validator, value: value}

    type View = Set[Entry]
    pure val bot:View = Set()

    pure def isSupersededBy(v1: View, v2: View): bool = true //TODO

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

    run isSupersededByTest = all {
        //Higher hight
        assert(Set(makeEntry(1,1,"Proposal",1,"")).isSupersededBy(Set(makeEntry(2,1,"Proposal",1,"")))),
        assert(Set(makeEntry(1,1,"Proposal",1,"")).isSupersededBy(Set(makeEntry(2,2,"Proposal",2,"")))),

        //Different validator from the same hight
        assert(not(Set(makeEntry(1,1,"Proposal",1,"")).isSupersededBy(Set(makeEntry(1,1,"Proposal",2,""))))),
        /*
        assert(Set((1,1,2)).isSupersededBy(Set((1,1,2)))),
        assert(Set((1,1,1)).isSupersededBy(Set((1,1,2)))),
        assert(Set((2,1,1)).isSupersededBy(Set((2,1,1)))),
        assert(not(Set((2,1,1)).isSupersededBy(Set((2,2,2))))),
        assert(not(Set((2,1,1)).isSupersededBy(Set((1,1,1))))),
        assert(not(Set((2,1,1)).isSupersededBy(Set((1,2,2))))),

        assert(makeEntry(1,1,"Proposal",1,"")).isSupersededBy(Set((1,1,1,true)))),
        */
    }

    run removeStaleTest = all {
        true,
        /*
        assert(Set((1,1,1)).removeStale() 
                == (Set((1,1,1)), bot)),
        assert(Set((1,1,1), (2,1,1), (2,1,2), (2,2,2)).removeStale() 
                == (Set((2,1,2), (2,2,2)), Set((1,1,1),(2,1,1)))),
        */
    }

    run mergeTest = all {
        true,
        /*
        assert(Set((1,1,1)).merge(Set((1,1,1))) == Set((1,1,1))),
        assert(Set((1,1,1), (2,1,2)).merge(Set((2,1,1), (2,1,2), (2,2,2))) == Set((2,1,2), (2,2,2))),
        assert(Set((1,1,1), (1,2,1)).merge(Set((1,3,1), (1,1,2), (1,2,2))) == Set((1,3,1), (1,1,2), (1,2,2))),
        */
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
}