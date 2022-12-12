module CRDT {
    import Globals.* from "./globals"
    import Spells.* from "./spells"

    type Value = str //TODO

    //TODO: how to use Nat as a type?
    type NatNum = int

    type MsgType =  | {tag: "Proposal"}
                    | {tag: "PreVote"}
                    | {tag:"PreCommit"}

    type Entry = {height: NatNum, round: NatNum, msgType: MsgType, validator: ProcId, value: Value}

    pure def makeEntry(height: NatNum, round: NatNum, msgType: MsgType, validator: ProcId, value: Value): Entry =
        {height: height, round: round, msgType: msgType, validator: validator, value: value}

    type View = {addSet: Set[Entry], delSet: Set[Entry]}

    pure def makeView(adds: Set[Entry], dels: Set[Entry]): View =
        {addSet: adds, delSet: dels}

    pure val bot:View = {addSet:Set(), delSet:Set()}

    pure def removeStale(v: View): View = v //TODO: drop deprecated entries.

    def globalView(lVs: ProcId -> View): View =
        val viewUnion = lVs.keys().fold(bot, (acc, p) => {
            val view = lVs.get(p)
            {addSet:acc.addSet.union(view.addSet), delSet:acc.delSet.union(view.delSet)}
        })
        removeStale(makeView(viewUnion.addSet.exclude(viewUnion.delSet), viewUnion.delSet))

    pure def addEntry(v:View, e: Entry): View =
        if (v.delSet.contains(e))
            v
        else
            v.with("addSet", v.addSet.union(Set(e)))

    pure def delEntry(v:View, e: Entry): View =
        v.with("addSet", v.addSet.exclude(Set(e)))
         .with("delSet", v.delSet.union(Set(e)))

    pure def addToProcessView(lVs: ProcId -> View, p: ProcId, e: Entry): ProcId -> View =
        val lVp = lVs.getOrElse(p, bot).addEntry(e)
        lVs.set(p, lVp)

    pure def delFromProcessView(lVs: ProcId -> View, p: ProcId, e: Entry): ProcId -> View =
        val lVp = lVs.getOrElse(p, bot).delEntry(e)
        lVs.set(p, lVp)

    pure def merge(lhs: View, rhs: View): View =
        val dels = lhs.delSet.union(rhs.delSet)
        val adds = lhs.addSet.union(rhs.addSet).exclude(dels)
        removeStale(makeView(adds, dels))





    run tests = {
        val entry11 = makeEntry(1, 1, {tag:"Proposal"}, 1, "nada")
        val entry12 = makeEntry(1, 2, {tag:"Proposal"}, 1, "nada")
        val entry21 = makeEntry(2, 1, {tag:"Proposal"}, 1, "nada")
        val entry22 = makeEntry(2, 2, {tag:"Proposal"}, 1, "nada")
        val view11 = bot.addEntry(entry11)
        val view12 = bot.addEntry(entry12)
        val view112 = bot.addEntry(entry11).addEntry(entry12)
        val view121 = bot.addEntry(entry11).addEntry(entry12).delEntry(entry11)
        all {
            assert(view11 == makeView(Set(entry11), Set())),
            assert(view12 == makeView(Set(entry12), Set())),
            assert(view112 == makeView(Set(entry11, entry12), Set())),
            assert(view121 == makeView(Set(entry12), Set(entry11))),
            assert(view11.merge(view12) == view112),

            assert(Map().addToProcessView(1, entry11).addToProcessView(2, entry12) == Map(1 -> makeView(Set(entry11), Set()), 2 -> makeView(Set(entry12), Set()))),

            assert(Map().addToProcessView(1, entry11).addToProcessView(1, entry12).delFromProcessView(1, entry11) ==
                    Map(1 -> makeView(Set(entry12), Set(entry11)))),
        }
    }
}