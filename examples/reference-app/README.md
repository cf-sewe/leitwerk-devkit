# reference-app

A minimal repository already onboarded onto Leitwerk. It exists so the gate can
be run and seen turning green — the worked example the whitepaper points to.

```
cd examples/reference-app
leitwerk verify --tier T0      # PASS (lint has nothing to run here, skips cleanly)
leitwerk tier leitwerk/constitution.md   # -> T0
```

`leitwerk/constitution.md` is a filled-in constitution. As this app grows, wire
the checks in `core/checks/*` to a real toolchain and the gate stops skipping.
