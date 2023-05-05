module CRDT {
    import Spells.* from "./spells"

    //TODO: how to use Nat as a type?
    type NatNum = int

    type Entry = (NatNum, NatNum)

    //True if this entry is a tombstone
    type EntryOrTs = (Entry, bool)

    pure def isTs(ets: EntryOrTs): bool = ets._2

    pure def makeTs(ets: EntryOrTs): EntryOrTs = (ets._1, true)

    pure def makeLive(ets: EntryOrTs): EntryOrTs = (ets._1, false)

    type View = Set[EntryOrTs]

    pure val bot:View = Set()

    pure def hasEntry(v: View, e:Entry):bool = v.contains((e, false))

    pure def isSupersededIn(ets: EntryOrTs, view: View): bool = any {
        //There is a tombstone for this entry
        not (ets.isTs()) and view.contains(ets.makeTs()),

        //Aplication specific rules
    }

    // Application specific.
    pure def removeStale(view: View): (View,View) =
        val superseded = view.filter(ets => ets.isSupersededIn(view.excludeE(ets)))
        val alive = view.exclude(superseded)
        (alive, superseded)

    // Adds an entry to the view.
    pure def addEntry(v:View, e: Entry): View =
        val newViewAndStale = v.unionE((e, false)).removeStale()
        newViewAndStale._1

    pure def merge(lhs: View, rhs: View): View =
        val bothHs = lhs.union(rhs)
        val tombstones = bothHs.filter(ets => ets.isTs())
        val live = bothHs.filter(ets => not(ets.isTs()))
                         .filter(ets => not(tombstones.contains(ets.makeLive())))

        val newViewAndStale = removeStale(live)
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