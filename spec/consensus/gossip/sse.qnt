module Lexy1 {
    import Spells.* from "./spells"

    //User provided
    type Entry = (int, int, int)

    type View = Set[Entry]
    pure val bot:View = Set()

    //User provided
    pure def isSupersededBy(v1: View, v2: View): bool = 
        v1.forall(e1 => 
            v2.exists(e2 => 
                or (
                    e1._1 < e2._1, 
                    and (
                        e1._1 == e2._1,
                        e1._2 == e2._2,
                        e1._3 <= e2._3
                    )
                )
            )
        )

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





    run isSupersededByTest = all {
        assert(Set((1,1,1)).isSupersededBy(Set((2,1,1)))),
        assert(Set((1,1,1)).isSupersededBy(Set((2,2,2)))),
        assert(Set((1,1,2)).isSupersededBy(Set((1,1,2)))),
        assert(Set((1,1,1)).isSupersededBy(Set((1,1,2)))),
        assert(Set((2,1,1)).isSupersededBy(Set((2,1,1)))),
        assert(not(Set((2,1,1)).isSupersededBy(Set((1,1,1))))),
        assert(not(Set((2,1,1)).isSupersededBy(Set((1,2,2))))),
    }

    run removeStaleTest = all {
        assert(Set((1,1,1)).removeStale() 
                == (Set((1,1,1)), bot)),
        assert(Set((1,1,1), (2,1,1), (2,1,2), (2,2,2)).removeStale() 
                == (Set((2,1,2), (2,2,2)), Set((1,1,1),(2,1,1)))),
    }

    run mergeTest = all {
        assert(Set((1,1,1)).merge(Set((1,1,1))) == Set((1,1,1))),
        assert(Set((1,1,1), (2,1,2)).merge(Set((2,1,1), (2,1,2), (2,2,2))) == Set((2,1,2), (2,2,2))),
    }

    run addEntryTest = all {
        assert(bot.addEntry((1,1,1)) == Set((1,1,1))),
        assert(not(bot.addEntry((1,1,1)) == Set((1,1,2)))),
        assert(bot.addEntry((1,1,1)).addEntry((1,1,2)) == Set((1,1,2),(1,1,2))),
    }

    run hasEntryTest = all {
        assert(bot.addEntry((1,1,1)).hasEntry((1,1,1))),
        assert(not(bot.addEntry((1,1,1)).hasEntry((1,1,2)))),
        assert(bot.addEntry((1,1,1)).addEntry((1,1,2)).hasEntry((1,1,2))),
    }

}


/*

module Lexy2 {
    import Spells.* from "./spells"

    type Entry = (int, int, bool)

    pure def isTs(ets: EntryOrTs): bool = ets._2

    pure def makeTs(ets: EntryOrTs): EntryOrTs = (ets._1, true)

    pure def makeLive(ets: EntryOrTs): EntryOrTs = (ets._1, false)


    type View = Set[EntryOrTs]

    pure val bot:View = Set()

    pure def hasEntry(v: View, e:Entry):bool = v.contains((e, false))

    pure def isSupersededBy(ets: EntryOrTs, view: View): bool = any {
        //There is a tombstone for this entry
        not (ets.isTs()) and view.contains(ets.makeTs()),

        //val bothHs = 
        //val tombstones = bothHs.filter(ets => ets.isTs())
        //val live = bothHs.filter(ets => not(ets.isTs()))
        //                 .filter(ets => not(tombstones.contains(ets.makeLive())))

        //Aplication specific rules
    }

    // Application specific.
    pure def removeStale(view: View): (View,View) =
        val superseded = view.filter(ets => ets.isSupersededBy(view.excludeE(ets)))
        val alive = view.exclude(superseded)
        (alive, superseded)

    // Adds an entry to the view.
    pure def addEntry(v:View, e: Entry): View =
        val newViewAndStale = v.unionE((e, false)).removeStale()
        newViewAndStale._1

    pure def merge(lhs: View, rhs: View): View =
        val newViewAndStale = lhs.union(rhs).removeStale()
        newViewAndStale._1

    // Removes entry from the view and adds a tombstone.
    pure def delEntry(v:View, e: Entry): View =
        val newViewAndStale = v.unionE((e, true)).removeStale()
        newViewAndStale._1

    run Test = all {
        val entry11 = (1, 1)
        val entry12 = (1, 2)
        val entry21 = (2, 1)
        val entry22 = (2, 2)
        val view11 = bot.addEntry(entry11)
        val view12 = bot.addEntry(entry12)
        val view112 = bot.addEntry(entry11).addEntry(entry12)
        val view121 = bot.addEntry(entry11).addEntry(entry12).delEntry(entry11)
        all {
            assert( not( bot.hasEntry(entry11))),
            assert(view11.hasEntry(entry11)),
            assert( not( view12.hasEntry(entry11))),
            assert(view12.hasEntry(entry12)),
            assert( not( view11.hasEntry(entry12))),

            assert( view121.hasEntry(entry12)),
            assert( not(view121.hasEntry(entry11))  ),

            assert(view11.merge(view12) == view112)
        }
    }
}
*/