- `[state/kvindex]` Querying event attributes that are bigger than int64 is now enabled as well as properly converting float event values 
from the database into int64 values ,even if the float is bigger than an int64. 
  ([\#771](https://github.com/cometbft/cometbft/pull/771))