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
        assert(Set((1,1,1), (1,2,1)).merge(Set((1,3,1), (1,1,2), (1,2,2))) == Set((1,3,1), (1,1,2), (1,2,2))),
    }

    run addEntryTest = all {
        assert(bot.addEntry((1,1,1)) == Set((1,1,1))),
        assert(not(bot.addEntry((1,1,1)) == Set((1,1,2)))),
        assert(bot.addEntry((1,1,1)).addEntry((1,1,2)) == Set((1,1,2))),
    }

    run hasEntryTest = all {
        assert(bot.addEntry((1,1,1)).hasEntry((1,1,1))),
        assert(not(bot.addEntry((1,1,1)).hasEntry((1,1,2)))),
        assert(bot.addEntry((1,1,1)).addEntry((1,1,2)).hasEntry((1,1,2))),
    }

}



module Lexy2 {
    import Spells.* from "./spells"

    //User provided
    type Entry = (int, int, int, bool)

    //User provided
    pure def toTombstone(e: Entry): Entry = (e._1, e._1, e._3, true)

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
                        e1._3 < e2._3
                    ),
                    and (
                        e1._1 == e2._1,
                        e1._2 == e2._2,
                        e1._3 == e2._3,
                        or (
                            e1._4 == e2._4,
                            e2._4 == true
                        )
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

    // Removes entry from the view and adds a tombstone.
    pure def delEntry(v:View, e: Entry): View =
        v.addEntry(e.toTombstone())

    run isSupersededByTest = all {
        assert(Set((1,1,1,false)).isSupersededBy(Set((1,1,1,true)))),
        assert(not(Set((1,1,1,true)).isSupersededBy(Set((1,1,1,false))))),
        assert(Set((1,1,1,false)).isSupersededBy(Set((2,1,1,false)))),
        assert(Set((1,1,1,true)).isSupersededBy(Set((2,1,1,false)))),
        assert(Set((1,1,1,false)).isSupersededBy(Set((2,1,1,true)))),
        assert(Set((1,1,1,true  )).isSupersededBy(Set((2,1,1,true)))),
        assert(Set((1,1,1,false)).isSupersededBy(Set((2,2,2,false)))),
        assert(Set((1,1,1,false)).isSupersededBy(Set((2,2,2,true)))),
        assert(Set((1,1,2,false)).isSupersededBy(Set((1,1,2,false)))),
        assert(Set((1,1,2,false)).isSupersededBy(Set((1,1,2,true)))),
        assert(Set((1,1,1,false)).isSupersededBy(Set((1,1,2,false)))),
        assert(Set((1,1,1,false)).isSupersededBy(Set((1,1,2,true)))),
        assert(Set((2,1,1,false)).isSupersededBy(Set((2,1,1,false)))),
        assert(Set((2,1,1,false)).isSupersededBy(Set((2,1,1,true)))),
        assert(not(Set((2,1,1,false)).isSupersededBy(Set((1,1,1,false))))),
        assert(not(Set((2,1,1,false)).isSupersededBy(Set((1,1,1,true))))),
        assert(not(Set((2,1,1,false)).isSupersededBy(Set((1,2,2,false))))),
        assert(not(Set((2,1,1,false)).isSupersededBy(Set((1,2,2,true))))),
    }

    run removeStaleTest = all {
        assert(Set((1,1,1,false)).removeStale() 
                == (Set((1,1,1,false)), bot)),
        assert(Set((1,1,1,true)).removeStale() 
                == (Set((1,1,1,true)), bot)),
        assert(Set((1,1,1,false),(1,1,1,true)).removeStale() 
                == (Set((1,1,1,true)), (Set((1,1,1,false))))),
        assert(Set((1,1,1,false), (2,1,1,false), (2,1,2,false), (2,2,2,false)).removeStale() 
                == (Set((2,1,2,false), (2,2,2,false)), Set((1,1,1,false),(2,1,1,false)))),
        assert(Set((1,1,1,false), (2,1,1,false), (2,1,2,false), (2,2,2,false), (4,4,4,true)).removeStale() 
                == (Set((4,4,4,true)), Set((2,1,2,false), (2,2,2,false), (1,1,1,false),(2,1,1,false)))),
    }
    run mergeTest = all {
        assert(Set((1,1,1,false)).merge(Set((1,1,1,false))) == Set((1,1,1,false))),
        assert(Set((1,1,1,false)).merge(Set((1,1,1,true))) == Set((1,1,1,true))),
        assert(Set((1,1,1,false), (2,1,2,false)).merge(Set((2,1,1,false), (2,1,2,false), (2,2,2,false)))
                == Set((2,1,2,false), (2,2,2,false))),
        assert(Set((1,1,1,false), (1,2,1,false)).merge(Set((1,3,1,false), (1,1,2,false), (1,2,2,false)))
                == Set((1,3,1,false), (1,1,2,false), (1,2,2,false))),

        assert(Set((1,1,1,false), (2,1,2,false)).merge(Set((2,1,1,true), (2,1,2,false), (2,2,2,false)))
                == Set((2,1,2,false), (2,2,2,false))),
        assert(Set((1,1,1,false), (1,2,1,false)).merge(Set((1,3,1,true), (1,1,2,false), (1,2,2,false)))
                == Set((1,3,1,true), (1,1,2,false), (1,2,2,false))),
    }

    run addEntryTest = all {
        assert(bot.addEntry((1,1,1,false)) == Set((1,1,1,false))),
        assert(bot.addEntry((1,1,1,false)).addEntry((1,1,1,true)) == Set((1,1,1,true))),
        assert(not(bot.addEntry((1,1,1,false)) == Set((1,1,2,false)))),
        assert(not(bot.addEntry((1,1,1,false)) == Set((1,1,1,true)))),
        assert(bot.addEntry((1,1,1,false)).addEntry((1,1,2,false)) == Set((1,1,2,false))),
    }

    run hasEntryTest = all {
        assert(bot.addEntry((1,1,1,false)).hasEntry((1,1,1,false))),
        assert(not(bot.addEntry((1,1,1,false)).hasEntry((1,1,1,true)))),
        assert(not(bot.addEntry((1,1,1,true)).hasEntry((1,1,1,false)))),
        assert(not(bot.addEntry((1,1,1,false)).hasEntry((1,1,1,true)))),
        assert(not(bot.addEntry((1,1,1,false)).hasEntry((1,1,2,false)))),
        assert(bot.addEntry((1,1,1,false)).addEntry((1,2,1,false)).hasEntry((1,2,1,false))),
        assert(bot.addEntry((1,1,1,false)).addEntry((1,2,1,false)).hasEntry((1,1,1,false))),
        assert(bot.addEntry((1,1,1,false)).addEntry((1,2,1,true)).hasEntry((1,2,1,true))),
        assert(bot.addEntry((1,1,1,false)).addEntry((1,2,1,true)).hasEntry((1,1,1,false))),
    }
}