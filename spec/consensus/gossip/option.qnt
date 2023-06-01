// -*- mode: Bluespec; -*-

module OInt {
    type OType = (int,bool)
    pure def valueOf(v: int):OType = (v, true)
    pure val none:OType = (0, false)

    pure def isSet(opt:OType):bool = (opt._2)

    //Not Set or is Equal To: nsoe
    pure def nsoe(opt:OType, v:int):bool = 
        or(not(opt.isSet()), opt._1 == v)
}

module OStr {
    type OType = (str,bool)
    pure def valueOf(v: str):OType = (v, true)
    pure val none:OType = ("", false)

    pure def isSet(opt:OType):bool = (opt._2)

    //Not Set or is Equal To: nsoe
    pure def nsoe(opt:OType, v:str):bool = 
        or(not(opt.isSet()), opt._1 == v)
}

module OBool {
    type OType = (bool,bool)
    pure def valueOf(v: bool):OType = (v, true)
    pure val none:OType = (false, false)

    pure def isSet(opt:OType):bool = (opt._2)

    //Not Set or is Equal To: nsoe
    pure def nsoe(opt:OType, v:bool):bool = 
        or(not(opt.isSet()), opt._1 == v)
}

module OptionTest {
    import OInt as OInt
    import OStr as OStr
    import OBool as OBool

    run optIntTest = all {
        val optSet = OInt::valueOf(3)
        assert(OInt::isSet(optSet)),

        val optSet = OInt::valueOf(5)
        assert(OInt::isSet(optSet)),

        val optSet = OInt::none
        assert(not(OInt::isSet(optSet))),
    }

    run optStrTest = all {
        val optSet = OStr::valueOf("3")
        assert(OStr::isSet(optSet)),

        val optSet = OStr::valueOf("5")
        assert(OStr::isSet(optSet)),

        val optSet = OStr::none
        assert(not(OStr::isSet(optSet))),
    }

    run optBoolTest = all {
        val optSet = OBool::valueOf(true)
        assert(OBool::isSet(optSet)),

        val optSet = OBool::valueOf(false)
        assert(OBool::isSet(optSet)),

        val optSet = OBool::none
        assert(not(OBool::isSet(optSet))),
    }
}