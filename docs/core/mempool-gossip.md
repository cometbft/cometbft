---
order: 12
---

# Mempool: propagation protocol

## Inbound messages

```
def Receive(msg: {txs: bytes, sender: Id}) {
    for tx in msg.txs:
        res = mempool.CheckTx(tx);
        addSender(tx, getId(sender))
}
```

## Outbound messages

```
def broadcastTxRoutine(peer) {
    iterator := memR.mempool.NewIterator()
}
```